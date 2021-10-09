package types

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/tools/amount"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/rlp"
	"github.com/aiot-network/aiotchain/types"
)

// token structure, issuing a contract with the same
// name is equivalent to reissuing the pass
type TokenRecord struct {
	Address        arry.Address
	Sender         arry.Address
	Name           string
	Shorthand      string
	IncreaseIssues bool
	PledgeRate     PledgeRate
	PledgeAmount   uint64
	Records        *RecordList
}

func NewToken() *TokenRecord {
	return &TokenRecord{Records: &RecordList{}}
}

func DecodeToken(bytes []byte) (*TokenRecord, error) {
	var token *TokenRecord
	if err := rlp.DecodeBytes(bytes, &token); err != nil {
		return nil, err
	}
	return token, nil
}

func (t *TokenRecord) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(t)
	return bytes
}

func (t *TokenRecord) IsExist(msgHash arry.Hash) bool {
	for _, r := range *t.Records {
		if msgHash.IsEqual(r.MsgHash) {
			return true
		}
	}
	return false
}

func (t *TokenRecord) Symbol() string {
	return t.Shorthand
}

func (t *TokenRecord) CheckToken(msg types.IMessage) error {
	body := msg.MsgBody().(*TokenBody)
	if !t.Sender.IsEqual(msg.From()) {
		return errors.New("the token already exists")
	}
	if !t.IncreaseIssues {
		return errors.New("token does not allow increase issuance")
	}
	if t.Shorthand != body.Shorthand {
		return errors.New("token shorthand is not consistent")
	}
	if t.Name != body.Name {
		return errors.New("token name is not consistent")
	}
	if !t.Address.IsEqual(body.TokenAddress) {
		return errors.New("token address is not consistent")
	}
	if t.IsExist(msg.Hash()) {
		return errors.New("duplicate message hash")
	}
	fAmount := amount.Amount(t.amount() + body.Amount).ToCoin()
	if fAmount < 0 {
		return fmt.Errorf("the total number of coins must not exceed %.8f", config.Param.MaxCoinCount)
	}
	if fAmount > config.Param.MaxCoinCount {
		return fmt.Errorf("the total number of coins must not exceed %.8f", config.Param.MaxCoinCount)
	}
	return nil
}

func (t *TokenRecord) CheckTokenV2(msg types.IMessage) error {
	body := msg.MsgBody().(*TokenV2Body)
	if !t.Sender.IsEqual(msg.From()) {
		return errors.New("the token already exists")
	}
	if !t.IncreaseIssues {
		return errors.New("token does not allow increase issuance")
	}
	if t.Shorthand != body.Shorthand {
		return errors.New("token shorthand is not consistent")
	}
	if t.Name != body.Name {
		return errors.New("token name is not consistent")
	}
	if !t.Address.IsEqual(body.TokenAddress) {
		return errors.New("token address is not consistent")
	}
	if t.IsExist(msg.Hash()) {
		return errors.New("duplicate message hash")
	}
	fAmount := amount.Amount(t.amount() + body.Amount).ToCoin()
	if fAmount < 0 {
		return fmt.Errorf("the total number of coins must not exceed %.8f", config.Param.MaxCoinCount)
	}
	if fAmount > config.Param.MaxCoinCount {
		return fmt.Errorf("the total number of coins must not exceed %.8f", config.Param.MaxCoinCount)
	}
	return nil
}

func (t *TokenRecord) CheckRedemption(msg types.IMessage) error {
	body := msg.MsgBody().(*RedemptionBody)
	if !t.Address.IsEqual(body.TokenAddress) {
		return errors.New("token address is not consistent")
	}
	if t.IsExist(msg.Hash()) {
		return errors.New("duplicate message hash")
	}
	fAmount := amount.Amount(t.amount() + body.Amount).ToCoin()
	if fAmount < 0 {
		return fmt.Errorf("the total number of coins must not exceed %.8f", config.Param.MaxCoinCount)
	}
	if fAmount > config.Param.MaxCoinCount {
		return fmt.Errorf("the total number of coins must not exceed %.8f", config.Param.MaxCoinCount)
	}
	if body.PledgeRate != t.PledgeRate {
		return fmt.Errorf("the redemption ratio of %d is not the same as the pledge ratio of %d", body.PledgeRate, t.PledgeRate)
	}
	return nil
}

func (t *TokenRecord) IncreaseRecord(record *Record) {
	t.Records.Set(record)
}

func (t *TokenRecord) FallBack(height uint64) error {
	for _, record := range *t.Records {
		if record.Height > height {
			t.Records.Remove(height)
		}
	}
	return nil
}

func (t *TokenRecord) amount() uint64 {
	var sum uint64
	for _, record := range *t.Records {
		sum += record.Amount
	}
	return sum
}

type Record struct {
	Height   uint64
	Type     string
	MsgHash  arry.Hash
	Receiver arry.Address
	Time     uint64
	Amount   uint64
}

type RecordList []*Record

func (r *RecordList) Get(height uint64) (*Record, bool) {
	for _, record := range *r {
		if height == record.Height {
			return record, true
		}
	}
	return &Record{}, false
}

func (r *RecordList) Set(newRecord *Record) {
	for i, record := range *r {
		if newRecord.Height == record.Height {
			(*r)[i] = newRecord
			return
		}
	}
	*r = append(*r, newRecord)
}

func (r *RecordList) Remove(height uint64) {
	for i, record := range *r {
		if record.Height == height {
			(*r) = append((*r)[0:i], (*r)[i+1:]...)
			return
		}
	}
}

func (r *RecordList) Len() int {
	return len(*r)
}
