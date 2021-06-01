package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/tools/amount"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	"github.com/aiot-network/aiotchain/tools/crypto/hash"
	"github.com/aiot-network/aiotchain/tools/rlp"
	"github.com/aiot-network/aiotchain/types"
)

var CoinBase = arry.StringToAddress("coinbase")

type Message struct {
	Header *MsgHeader
	Body   types.IMessageBody
}

func (m *Message) Type() int {
	return int(m.Header.Type)
}

func (m *Message) Hash() arry.Hash {
	return m.Header.Hash
}

func (m *Message) From() arry.Address {
	return m.Header.From
}

func (m *Message) Nonce() uint64 {
	return m.Header.Nonce
}

func (m *Message) Fee() uint64 {
	return m.Header.Fee
}

func (m *Message) Time() uint64 {
	return m.Header.Time
}

func (m *Message) IsCoinBase() bool {
	return m.Header.From.IsEqual(CoinBase)
}

func (m *Message) MsgTo() types.IReceiver {
	return m.Body.MsgTo()
}

func (m *Message) MsgBody() types.IMessageBody {
	return m.Body
}

func (m *Message) MsgAmount() uint64 {
	return m.Body.MsgAmount()
}

func (m *Message) Signature() string {
	return m.Header.Signature.SignatureString()
}

func (m *Message) PublicKey() string {
	return m.Header.Signature.PubKeyString()
}

func (m *Message) ToRlp() types.IRlpMessage {
	rlpMsg := &RlpMessage{}
	rlpMsg.MsgHeader = m.Header
	rlpMsg.MsgBody, _ = rlp.EncodeToBytes(m.Body)
	return rlpMsg
}

func (m *Message) Check() error {
	if m.Header == nil || m.Body == nil {
		return errors.New("incomplete message")
	}

	if err := m.checkHash(); err != nil {
		return err
	}

	if err := m.checkFees();err != nil{
		return err
	}

	if err := m.Header.Check(); err != nil {
		return err
	}

	if err := m.Body.CheckBody(m.From()); err != nil {
		return err
	}
	return nil
}

func (m *Message) CheckCoinBase(fee uint64, coinbase uint64) error {
	nTx := m.Body.(*TransactionBody)
	sumAmount := coinbase + fee
	if sumAmount != nTx.MsgAmount() {
		return fmt.Errorf("the fee of %d and the reward of %d are not consistent "+
			"with amount %d", fee, coinbase, nTx.MsgAmount())
	}
	return nil
}


func (m *Message) checkFees() error {
	if m.Header.Type == Work{
		return nil
	}
	fees := uint64(minFees * len(m.MsgTo().ReceiverList()))
	if m.Header.Fee < fees  {
		return fmt.Errorf("fees %.8f is less than the minimum poundage allowed %.8f", amount.Amount(m.Header.Fee).ToCoin(), amount.Amount(fees).ToCoin())
	}
	return nil
}

func (m *Message) checkHash() error {
	newMsg := m.copy()
	err := newMsg.SetHash()
	if err != nil {
		return err
	}
	if newMsg.Hash().IsEqual(m.Hash()) {
		return nil
	}
	return errors.New("error messages hash")
}

func (m *Message) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(m)
	return bytes
}

func (m *Message) SignMsg(key *secp256k1.PrivateKey) error {
	var err error
	if m.Header.Signature, err = Sign(key, m.Header.Hash); err != nil {
		return err
	}
	return nil
}

func (m *Message) SetHash() error {
	m.Header.Hash = arry.Hash{}
	m.Header.Signature = &Signature{}
	rpcMsg, err := MsgToRpcMsg(m)
	if err != nil {
		return err
	}
	mBytes, err := json.Marshal(rpcMsg)
	if err != nil {
		return err
	}
	m.Header.Hash = hash.Hash(mBytes)
	return nil
}

func (m *Message) SignMessage(key *secp256k1.PrivateKey) error {
	var err error
	if m.Header.Signature, err = Sign(key, m.Header.Hash); err != nil {
		return err
	}
	return nil
}

func (m *Message) copy() *Message {
	return &Message{
		Header: &MsgHeader{
			Hash:      m.Header.Hash,
			Type:      m.Header.Type,
			From:      m.Header.From,
			Nonce:     m.Header.Nonce,
			Fee:       m.Header.Fee,
			Time:      m.Header.Time,
			Signature: m.Header.Signature,
		},
		Body: m.Body,
	}
}
