package types

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/tools/arry"
)

type MessageType uint8

const (
	Transaction MessageType = iota
	Token
	Candidate
	Cancel
	Vote
	Work
	TokenV2
	Redemption
	Contract
)

const (
	minFees      = 1e4
	maxFees      = 1e9
	ContractFees = 1e5
)

type MsgHeader struct {
	Type      MessageType
	Hash      arry.Hash
	From      arry.Address
	Nonce     uint64
	Fee       uint64
	Time      uint64
	Signature *Signature
}

func (m *MsgHeader) Check() error {
	if err := m.checkType(); err != nil {
		return err
	}

	if err := m.checkFrom(); err != nil {
		return err
	}

	if err := m.checkSinger(); err != nil {
		return err
	}
	return nil
}

func (m *MsgHeader) checkType() error {
	switch m.Type {
	case Transaction:
		return nil
		//case Token:
		//case Candidate:
		//case Cancel:
		//case Vote:
	case Work:
		return nil
	case TokenV2:
		return nil
	case Redemption:
		return nil
	case Contract:
		return nil
	}
	return fmt.Errorf("there are no messages of type %d", m.Type)
}

func (m *MsgHeader) checkFrom() error {
	if !kit.CheckAddress(config.Param.Name, m.From.String()) {
		return fmt.Errorf("%s address illegal", m.From.String())
	}
	return nil
}

func (m *MsgHeader) checkSinger() error {
	if !Verify(m.Hash, m.Signature) {
		return errors.New("signature verification failed")
	}

	if !VerifySigner(config.Param.Name, m.From, m.Signature.PubKey) {
		return errors.New("signer and sender do not match")
	}

	if m.Type == Work {
		if m.From.String() != config.Param.WorkProofAddress {
			return errors.New("incorrect signature address")
		}
	}
	return nil
}
