package dpos_status

import (
	"github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/tools/arry"
)

type IDPosDB interface {
	SetRoot(hash arry.Hash) error
	Root() arry.Hash
	Commit() (arry.Hash, error)
	CandidatesCount() int
	Candidates() (*types.Candidates, error)
	AddCandidate(member *types.Member)
	CancelCandidate(signer arry.Address)
	CycleSupers(cycle uint64) (*types.Supers, error)
	SaveCycle(cycle uint64, supers *types.Supers)
	Voters() map[arry.Address][]arry.Address
	Confirmed() (uint64, error)
	SetConfirmed(uint64)
	Voter(from, to arry.Address)
	AddSuperBlockCount(cycle uint64, signer arry.Address)
	SuperBlockCount(cycle uint64, signer arry.Address) uint32
}
