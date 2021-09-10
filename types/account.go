package types

import (
	"github.com/aiot-network/aiotchain/tools/arry"
)

type IAccount interface {
	NeedUpdate() bool
	UpdateLocked(confirmed uint64) error
	FromMessage(msg IMessage, height uint64) error
	ToMessage(msgType int, address, token arry.Address, amount, height uint64) error
	WorkMessage(address arry.Address, workload, cycle, endTime uint64)
	EaterMessage(height uint64) error
	Check(msg IMessage, strict bool) error
	Bytes() []byte
	GetAddress() arry.Address
	GetBalance(tokenAddr arry.Address) uint64
	GetWorks() IWorks

	TransferOut(token arry.Address, amount, height uint64) error
	TransferIn(token arry.Address, amount, height uint64) error
	ContractChangeTo(re *Receiver, contract arry.Address, blockHeight uint64)
}

type IActStatus interface {
	Nonce(arry.Address) uint64
	SetTrieRoot(hash arry.Hash) error
	TrieRoot() arry.Hash
	SetConfirmed(confirmed uint64)
	Account(address arry.Address) IAccount
	CheckMessage(msg IMessage, strict bool) error
	FromMessage(msg IMessage, height uint64) error
	WorkMessage(msg IMessage) error
	ToMessage(msg IMessage, height uint64) error
	Commit() (arry.Hash, error)
}
