package act_status

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/db/status/act_db"
	fmtypes "github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/utils"
	"github.com/aiot-network/aiotchain/types"
	"sync"
	"time"
)

const account_db = "account_db"

type ActStatus struct {
	db        IActDB
	mutex     sync.RWMutex
	confirmed uint64
}

func NewActStatus() (*ActStatus, error) {
	db, err := act_db.Open(config.Param.Data + "/" + account_db)
	if err != nil {
		return nil, err
	}
	return &ActStatus{db: db}, nil
}

// Initialize account balance root hash
func (a *ActStatus) SetTrieRoot(stateRoot arry.Hash) error {
	return a.db.SetRoot(stateRoot)
}

func (a *ActStatus) CheckMessage(msg types.IMessage, strict bool) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	now := uint64(utils.NowUnix())
	if msg.Time() > now+60*10 {
		return fmt.Errorf("incorrect message time, msg time = %d, now time = %d", msg.Time(), now)
	}

	account := a.Account(msg.From())
	if err := account.Check(msg, strict); err != nil {
		return err
	}
	return a.checkBody(msg)
}

func (a *ActStatus) checkBody(msg types.IMessage) error {
	switch fmtypes.MessageType(msg.Type()) {
	case fmtypes.Work:
		body, ok := msg.MsgBody().(*fmtypes.WorkBody)
		if !ok {
			return errors.New("wrong message")
		}
		cycle := msg.Time() / param.CycleInterval
		for _, work := range body.List {
			addrAct := a.db.Account(work.Address)
			work := addrAct.GetWorks()
			if cycle < work.GetCycle() {
				return errors.New("the work is overdue")
			}
			if body.StartTime < work.GetEndTime() {
				return errors.New("work start time overlaps with previous work")
			}
			if body.EndTime > uint64(time.Now().Unix()) {
				return errors.New("wong end time")
			}
			if body.EndTime <= body.StartTime {
				return errors.New("wong end time")
			}
		}
	}
	return nil
}

// Get account status, if the account status needs to be updated
// according to the effective block height, it will be updated,
// but not stored.
func (a *ActStatus) Account(address arry.Address) types.IAccount {
	a.mutex.RLock()
	account := a.db.Account(address)
	a.mutex.RUnlock()

	if account.NeedUpdate() {
		account = a.updateLocked(address)
	}
	return account
}

func (a *ActStatus) Nonce(address arry.Address) uint64 {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.db.Nonce(address)
}

func (a *ActStatus) WorkMessage(msg types.IMessage) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	body, ok := msg.MsgBody().(*fmtypes.WorkBody)
	if !ok {
		return errors.New("wrong message")
	}
	cycle := msg.Time() / param.CycleInterval
	for _, work := range body.List {
		addrAct := a.db.Account(work.Address)
		addrAct.WorkMessage(work.Address, work.Workload, cycle, work.EndTime)
		a.setAccount(addrAct)
	}
	return nil
}

// Update sender account status based on message information
func (a *ActStatus) FromMessage(msg types.IMessage, height uint64) error {
	if msg.IsCoinBase() {
		return nil
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	fromAct := a.db.Account(msg.From())
	err := fromAct.UpdateLocked(a.confirmed)
	if err != nil {
		return err
	}

	err = fromAct.FromMessage(msg, height)
	if err != nil {
		return err
	}

	a.setAccount(fromAct)
	return nil
}

// Update the receiver's account status based on message information
func (a *ActStatus) ToMessage(msg types.IMessage, height uint64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	msgBody := msg.MsgBody()
	receivers := msgBody.MsgTo().ReceiverList()

	switch fmtypes.MessageType(msg.Type()) {
	case fmtypes.Token:
		eater := a.db.Account(config.Param.EaterAddress)
		err := eater.UpdateLocked(a.confirmed)
		if err != nil {
			return err
		}
		err = eater.EaterMessage(height)
		if err != nil {
			return err
		}
		a.setAccount(eater)
		for _, re := range receivers {
			var toAct types.IAccount
			toAct = a.db.Account(re.Address)
			err := toAct.UpdateLocked(a.confirmed)
			if err != nil {
				return err
			}
			err = toAct.ToMessage(msg.Type(), re.Address, msgBody.MsgContract(), re.Amount, height)
			if err != nil {
				return err
			}
			// publish token need consume
			a.setAccount(toAct)
		}
	case fmtypes.Redemption:
		var toAct types.IAccount
		toAct = a.db.Account(msg.From())
		err := toAct.UpdateLocked(a.confirmed)
		if err != nil {
			return err
		}
		body, _ := msg.MsgBody().(*fmtypes.RedemptionBody)
		amount := body.RedemptionAmount()
		err = toAct.ToMessage(msg.Type(), msg.From(), config.Param.MainToken, amount, height)
		if err != nil {
			return err
		}
		// publish token need consume
		a.setAccount(toAct)
	default:
		for _, re := range receivers {
			var toAct types.IAccount
			toAct = a.db.Account(re.Address)
			err := toAct.UpdateLocked(a.confirmed)
			if err != nil {
				return err
			}
			err = toAct.ToMessage(msg.Type(), re.Address, msgBody.MsgContract(), re.Amount, height)
			if err != nil {
				return err
			}
			// publish token need consume

			a.setAccount(toAct)
		}
	}

	return nil
}

func (a *ActStatus) SetConfirmed(height uint64) {
	a.confirmed = height
}

func (a *ActStatus) Transfer(from, to, token arry.Address, amount uint64, height uint64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	fromAcc := a.Account(from)
	if err := fromAcc.TransferOut(token, amount, height); err != nil {
		return err
	}
	toAcc := a.Account(to)
	if err := toAcc.TransferIn(token, amount, height); err != nil {
		return err
	}
	a.setAccount(fromAcc)
	a.setAccount(toAcc)
	return nil
}

func (a *ActStatus) PreTransfer(from, to, token arry.Address, amount uint64, height uint64) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	fromAcc := a.Account(from)
	if err := fromAcc.TransferOut(token, amount, height); err != nil {
		return err
	}
	toAcc := a.Account(to)
	if err := toAcc.TransferIn(token, amount, height); err != nil {
		return err
	}
	return nil
}

func (a *ActStatus) PreBurn(from arry.Address, contract arry.Address, amount, height uint64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.preBurn(from, contract, amount, height)
}

func (a *ActStatus) preBurn(from arry.Address, contract arry.Address, amount, height uint64) error {
	var toAccount types.IAccount

	toAccount = a.db.Account(from)
	err := toAccount.UpdateLocked(a.confirmed)
	if err != nil {
		return err
	}

	return toAccount.TransferOut(contract, amount, height)
}

func (a *ActStatus) Burn(from arry.Address, contract arry.Address, amount, height uint64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.burn(from, contract, amount, height)
}

func (a *ActStatus) burn(from arry.Address, contract arry.Address, amount, height uint64) error {
	var toAccount types.IAccount

	toAccount = a.db.Account(from)
	err := toAccount.UpdateLocked(a.confirmed)
	if err != nil {
		return err
	}

	toAccount.TransferOut(contract, amount, height)
	a.setAccount(toAccount)
	return nil
}

func (a *ActStatus) Mint(reviver arry.Address, contract arry.Address, amount, height uint64) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.mint(&types.Receiver{
		Address: reviver,
		Amount:  amount,
	}, contract, height)
}

func (a *ActStatus) mint(receiver *types.Receiver, contract arry.Address, height uint64) error {
	var toAccount types.IAccount

	toAccount = a.db.Account(receiver.Address)
	err := toAccount.UpdateLocked(a.confirmed)
	if err != nil {
		return err
	}

	toAccount.ContractChangeTo(receiver, contract, height)
	a.setAccount(toAccount)
	return nil
}

// Verify the status of the trading account
func (a *ActStatus) Check(msg types.IMessage, strict bool) error {
	now := uint64(utils.NowUnix())
	if msg.Time() > now+60*10 {
		return fmt.Errorf("incorrect message time, msg time = %d, now time = %d", msg.Time(), now)
	}

	account := a.Account(msg.From())
	return account.Check(msg, strict)
}

func (a *ActStatus) Commit() (arry.Hash, error) {
	return a.db.Commit()
}

func (a *ActStatus) TrieRoot() arry.Hash {
	return a.db.Root()
}

func (a *ActStatus) Close() error {
	return a.db.Close()
}

func (a *ActStatus) setAccount(account types.IAccount) {
	a.db.SetAccount(account)
}

// Update the locked balance of an account
func (a *ActStatus) updateLocked(address arry.Address) types.IAccount {
	act := a.db.Account(address)
	act.UpdateLocked(a.confirmed)
	return act
}
