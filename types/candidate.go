package types

import "github.com/aiot-network/aiotchain/tools/arry"

type ICandidates interface {
	Len() int
	List() []ICandidate
	Get(int) ICandidate
	GetPreHash() arry.Hash
}

type ICandidate interface {
	GetPeerId() string
	GetSinger() arry.Address
	GetMntCount() uint32
}
