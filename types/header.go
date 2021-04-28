package types

import (
	"github.com/aiot-network/aiotchain/tools/arry"
)

type IHeader interface {
	GetHash() arry.Hash
	GetPreHash() arry.Hash
	GetMsgRoot() arry.Hash
	GetActRoot() arry.Hash
	GetDPosRoot() arry.Hash
	GetTokenRoot() arry.Hash
	GetSigner() arry.Address
	GetSignature() ISignature
	GetHeight() uint64
	GetTime() uint64
	GetCycle() uint64
	ToRlpHeader() IRlpHeader
	Bytes() []byte
}
