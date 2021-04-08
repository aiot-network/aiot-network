package status

import (
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/types"
)

type IStatus interface {
	InitRoots(actRoot, dPosRoot, tokenRoot arry.Hash) error
	Commit() (arry.Hash, arry.Hash, arry.Hash, error)
	SetConfirmed(confirmed uint64)
	CheckMsg(msg types.IMessage, strict bool) error
	Change(msgs []types.IMessage, block types.IBlock) error
	Account(address arry.Address) types.IAccount
	Token(address arry.Address) (types.IToken, error)
	Candidates() types.ICandidates
	CycleSupers(cycle uint64) types.ICandidates
	CycleReword(cycle uint64) []types.IReword
}
