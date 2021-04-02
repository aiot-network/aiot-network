package act_status

import (
	"errors"
	"github.com/aiot-network/aiot-network/chain/db/status/act_db"
	fmtypes "github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/common/config"
	"github.com/aiot-network/aiot-network/common/param"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/utils"
	"github.com/aiot-network/aiot-network/types"
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
	a.mutex.RUnlock()

	if msg.Time() > uint64(utils.NowUnix()) {
		return errors.New("incorrect transaction time")
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
	for _, re := range receivers {
		var toAct types.IAccount
		toAct = a.db.Account(re.Address)
		err := toAct.UpdateLocked(a.confirmed)
		if err != nil {
			return err
		}
		err = toAct.ToMessage(msg.Type(), re.Address, msgBody.MsgToken(), re.Amount, height)
		if err != nil {
			return err
		}
		// publish token need consume
		if fmtypes.MessageType(msg.Type()) == fmtypes.Token {
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
		}
		a.setAccount(toAct)
	}

	return nil
}

func (a *ActStatus) SetConfirmed(height uint64) {
	a.confirmed = height
}

// Verify the status of the trading account
func (a *ActStatus) Check(msg types.IMessage, strict bool) error {
	if msg.Time() > uint64(utils.NowUnix()) {
		return errors.New("incorrect message time")
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
