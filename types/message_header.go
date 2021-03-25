package types

import "github.com/aiot-network/aiot-network/tools/arry"

type IMessageHeader interface {
	Type() int
	Hash() arry.Hash
	From() arry.Address
	Nonce() uint64
	Fee() uint64
	Time() uint64
	IsCoinBase() bool
	Signature() string
	PublicKey() string
}
