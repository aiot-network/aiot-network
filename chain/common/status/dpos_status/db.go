package dpos_status

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/tools/arry"
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
	AddSuperWork(cycle uint64, super arry.Address, works *types.Works)
	SuperWork(cycle uint64, super arry.Address) (*types.Works, error)
}
