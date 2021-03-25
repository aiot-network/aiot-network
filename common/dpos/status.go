package dpos

import (
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/types"
)

type IDPosStatus interface {
	SetTrieRoot(hash arry.Hash) error
	TrieRoot() arry.Hash
	Confirmed() (uint64, error)
	SetConfirmed(height uint64)
	Candidates() (types.ICandidates, error)
	Voters() map[arry.Address][]arry.Address
	CycleSupers(cycle uint64) (types.ICandidates, error)
	SaveCycle(cycle uint64, supers types.ICandidates)
	CheckMessage(msg types.IMessage) error
	AddCandidate(msg types.IMessage) error
	CancelCandidate(msg types.IMessage) error
	Voter(msg types.IMessage) error
	AddSuperBlockCount(cycle uint64, signer arry.Address)
	SuperBlockCount(cycle uint64, signer arry.Address) uint32
	Commit() (arry.Hash, error)
}
