package types

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiot-network/chain/common/kit"
	"github.com/aiot-network/aiot-network/common/config"
	"github.com/aiot-network/aiot-network/tools/amount"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/math"
	"github.com/aiot-network/aiot-network/types"
	"time"
)

const (
	PeerLength = 53
	MaxName    = 100
)

type Peer [PeerLength]byte

func (p Peer) String() string {
	return string(p[:])
}

func (p Peer) Bytes() []byte {
	return p[:]
}

type TransactionBody struct {
	TokenAddress arry.Address
	Receivers    *Receivers
}

type Receivers struct {
	List []*types.Receiver
}

func NewReceivers() *Receivers {
	return &Receivers{
		List: make([]*types.Receiver, 0),
	}
}

func (r *Receivers) Add(address arry.Address, amount uint64) {
	r.List = append(r.List, &types.Receiver{
		Address: address,
		Amount:  amount,
	})
}

func (r *Receivers) CheckAddress() error {
	for _, re := range r.List {
		if !kit.CheckAddress(config.Param.Name, re.Address.String()) {
			return fmt.Errorf("receive address %s verification failed", re.Address.String())
		}
	}
	return nil
}

func (r *Receivers) CheckAmount() error {
	var sum uint64
	for _, re := range r.List {
		if re.Amount < config.Param.MinimumTransfer {
			return fmt.Errorf("the minimum allowed transfer is %d", config.Param.MinimumTransfer)
		}
		if re.Amount > config.Param.MaximumTransfer {
			return fmt.Errorf("the maximum allowed transfer is %d", config.Param.MaximumTransfer)
		}
		sum += re.Amount
		if sum > config.Param.MaximumTransfer {
			return fmt.Errorf("the maximum allowed transfer is %d", config.Param.MaximumTransfer)
		}
	}
	return nil
}

func (r *Receivers) ReceiverList() []*types.Receiver {
	return r.List
}

func (t *TransactionBody) MsgTo() types.IReceiver {
	return t.Receivers
}

func (t *TransactionBody) CheckBody(from arry.Address) error {
	if len(t.Receivers.List) > config.Param.MaximumReceiver {
		return fmt.Errorf("the maximum number of receive addresses is %d", config.Param.MaximumReceiver)
	}
	if len(t.Receivers.List) == 0 {
		return fmt.Errorf("no receivers")
	}
	if err := t.Receivers.CheckAddress(); err != nil {
		return err
	}
	if !t.TokenAddress.IsEqual(config.Param.MainToken) {
		if !kit.CheckTokenAddress(config.Param.Name, t.TokenAddress.String()) {
			return errors.New("token address verification failed")
		}
	}
	if err := t.Receivers.CheckAmount(); err != nil {
		return err
	}
	return nil
}

func (t *TransactionBody) MsgAmount() uint64 {
	var sum uint64
	for _, re := range t.Receivers.List {
		sum += re.Amount
	}
	return sum
}

func (t *TransactionBody) MsgToken() arry.Address {
	return t.TokenAddress
}

type TokenBody struct {
	TokenAddress   arry.Address
	Receiver       arry.Address
	Name           string
	Shorthand      string
	IncreaseIssues bool
	Amount         uint64
}

func (t *TokenBody) MsgTo() types.IReceiver {
	recis := NewReceivers()
	recis.Add(t.Receiver, t.Amount)
	return recis
}

func (t *TokenBody) CheckBody(from arry.Address) error {
	if !kit.CheckAddress(config.Param.Name, t.Receiver.String()) {
		return errors.New("receive address verification failed")
	}
	if !kit.CheckTokenAddress(config.Param.Name, t.TokenAddress.String()) {
		return errors.New("token address verification failed")
	}
	toKenAddr, err := kit.GenerateTokenAddress(config.Param.Name, from.String(), t.Shorthand)
	if err != nil {
		return errors.New("token address verification failed")
	}
	if toKenAddr != t.TokenAddress.String() {
		return errors.New("token address verification failed")
	}
	if err := kit.CheckShorthand(t.Shorthand); err != nil {
		return fmt.Errorf("shorthand verification failed, %s", err.Error())
	}
	if len(t.Name) > MaxName {
		return fmt.Errorf("the maximum length of the token name is %d", MaxName)
	}
	if t.Amount > math.MaxInt64 {
		return fmt.Errorf("amount cannot be greater than %.8f", amount.Amount(math.MaxInt64).ToCoin())
	}
	fAmount := amount.Amount(t.Amount).ToCoin()
	if fAmount < config.Param.MinCoinCount || fAmount > config.Param.MaxCoinCount {
		return fmt.Errorf("the quantity of coins must be between %.8f and %.8f", config.Param.MinCoinCount, config.Param.MaxCoinCount)
	}
	return nil
}

func (t *TokenBody) MsgAmount() uint64 {
	return t.Amount
}

func (t *TokenBody) MsgToken() arry.Address {
	return t.TokenAddress
}

type CandidateBody struct {
	Peer Peer
}

func (c *CandidateBody) MsgTo() types.IReceiver {
	return NewReceivers()
}

func (c *CandidateBody) CheckBody(from arry.Address) error {
	return nil
}

func (c *CandidateBody) MsgAmount() uint64 {
	return 0
}

func (c *CandidateBody) MsgToken() arry.Address {
	return config.Param.MainToken
}

type CancelBody struct {
}

func (c *CancelBody) MsgTo() types.IReceiver {
	return NewReceivers()
}

func (c *CancelBody) CheckBody(from arry.Address) error {
	return nil
}

func (c *CancelBody) MsgToken() arry.Address {
	return config.Param.MainToken
}

func (c *CancelBody) MsgAmount() uint64 {
	return 0
}

type VoteBody struct {
	To arry.Address
}

func (v *VoteBody) MsgTo() types.IReceiver {
	recis := NewReceivers()
	recis.Add(v.To, 0)
	return recis
}

func (v *VoteBody) CheckBody(from arry.Address) error {
	if !kit.CheckAddress(config.Param.Name, v.To.String()) {
		return errors.New("wrong to address")
	}
	return nil
}

func (v *VoteBody) MsgToken() arry.Address {
	return config.Param.MainToken
}

func (v *VoteBody) MsgAmount() uint64 {
	return 0
}

type WorkBody struct {
	StartTime uint64
	EndTime   uint64
	List      []AddressWork
}

type AddressWork struct {
	Address  arry.Address
	Workload uint64
	EndTime  uint64
}

func (w *WorkBody) MsgTo() types.IReceiver {
	recis := NewReceivers()
	return recis
}

func (w *WorkBody) CheckBody(from arry.Address) error {
	if len(w.List) > config.Param.SuperSize {
		return fmt.Errorf("it cannot exceed the maximum number of supernodes %d", config.Param.SuperSize)
	}
	if len(w.List) == 0 {
		return fmt.Errorf("no wokrs")
	}
	for _, work := range w.List {
		if !kit.CheckAddress(config.Param.Name, work.Address.String()) {
			return errors.New("wrong to address")
		}
	}
	if w.EndTime > uint64(time.Now().Unix()) {
		return errors.New("end time error")
	}
	if w.StartTime >= w.EndTime {
		return errors.New("start time error")
	}
	return nil
}

func (w *WorkBody) MsgToken() arry.Address {
	return config.Param.MainToken
}

func (w *WorkBody) MsgAmount() uint64 {
	return 0
}
