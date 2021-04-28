package types

import (
	"bytes"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/hash"
	"github.com/aiot-network/aiotchain/types"
)

type Messages []types.IMessage

func (m Messages) MsgList() []types.IMessage {
	return m
}

func (m Messages) Count() int {
	return len(m)
}

func MsgRoot(msgs []types.IMessage) arry.Hash {
	var hashes [][]byte
	for _, msg := range msgs {
		hashes = append(hashes, msg.Hash().Bytes())
	}
	hashBytes := bytes.Join(hashes, []byte{})
	return hash.Hash(hashBytes)
}

func CalculateFee(msgs []types.IMessage) uint64 {
	var sum uint64
	for _, msg := range msgs {
		sum += msg.Fee()
	}
	return sum
}
