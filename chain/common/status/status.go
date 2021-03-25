package status

import (
	"errors"
	chaintypes "github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/common/dpos"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/types"
)

const module = "chain"

type Status struct {
	actStatus   types.IActStatus
	dPosStatus  dpos.IDPosStatus
	tokenStatus types.ITokenStatus
}

func NewStatus(actStatus types.IActStatus, dPosStatus dpos.IDPosStatus, tokenStatus types.ITokenStatus) *Status {
	return &Status{
		actStatus:   actStatus,
		dPosStatus:  dPosStatus,
		tokenStatus: tokenStatus,
	}
}

func (f *Status) InitRoots(actRoot, dPosRoot, tokenRoot arry.Hash) error {
	if err := f.actStatus.SetTrieRoot(actRoot); err != nil {
		return err
	}
	if err := f.dPosStatus.SetTrieRoot(dPosRoot); err != nil {
		return err
	}
	if err := f.tokenStatus.SetTrieRoot(tokenRoot); err != nil {
		return err
	}
	return nil
}

func (f *Status) SetConfirmed(confirmed uint64) {
	f.actStatus.SetConfirmed(confirmed)
}

func (f *Status) Account(address arry.Address) types.IAccount {
	return f.actStatus.Account(address)
}

func (f *Status) CheckMsg(msg types.IMessage, strict bool) error {
	if err := msg.Check(); err != nil {
		return err
	}

	if err := f.dPosStatus.CheckMessage(msg); err != nil {
		return err
	}

	if err := f.actStatus.CheckMessage(msg, strict); err != nil {
		return err
	}

	if err := f.tokenStatus.CheckMessage(msg); err != nil {
		return err
	}
	return nil
}

func (f *Status) Change(msgs []types.IMessage, block types.IBlock) error {
	for _, msg := range msgs {
		switch chaintypes.MessageType(msg.Type()) {
		case chaintypes.Transaction:
			if err := f.actStatus.ToMessage(msg, block.GetHeight()); err != nil {
				return err
			}
		case chaintypes.Token:
			if err := f.actStatus.ToMessage(msg, block.GetHeight()); err != nil {
				return err
			}
			if err := f.tokenStatus.UpdateToken(msg, block.GetHeight()); err != nil {
				return err
			}
		case chaintypes.Vote:
			if err := f.dPosStatus.Voter(msg); err != nil {
				return nil
			}
		case chaintypes.Candidate:
			if err := f.dPosStatus.AddCandidate(msg); err != nil {
				return nil
			}
		case chaintypes.Cancel:
			if err := f.dPosStatus.CancelCandidate(msg); err != nil {
				return nil
			}
		case chaintypes.Work:
			if err := f.actStatus.WorkMessage(msg); err != nil {
				return nil
			}
		default:
			return errors.New("wrong message type")
		}
		if err := f.actStatus.FromMessage(msg, block.GetHeight()); err != nil {
			return err
		}

	}
	f.dPosStatus.AddSuperBlockCount(block.GetCycle(), block.GetSigner())
	return nil
}

func (f *Status) Commit() (arry.Hash, arry.Hash, arry.Hash, error) {
	actRoot, err := f.actStatus.Commit()
	if err != nil {
		return arry.Hash{}, arry.Hash{}, arry.Hash{}, err
	}
	tokenRoot, err := f.tokenStatus.Commit()
	if err != nil {
		return arry.Hash{}, arry.Hash{}, arry.Hash{}, err
	}
	dPosRoot, err := f.dPosStatus.Commit()
	if err != nil {
		return arry.Hash{}, arry.Hash{}, arry.Hash{}, err
	}
	return actRoot, tokenRoot, dPosRoot, nil
}

func (f *Status) Candidates() types.ICandidates {
	iCans, _ := f.dPosStatus.Candidates()
	cans := iCans.(*chaintypes.Candidates)
	voterMap := f.dPosStatus.Voters()
	for index, candidate := range cans.Members {
		voters, ok := voterMap[candidate.Signer]
		if ok {
			cans.Members[index].Voters = voters
		}
	}
	return cans
}

func (f *Status) CycleSupers(cycle uint64) types.ICandidates {
	candidates, err := f.dPosStatus.CycleSupers(cycle)
	if err != nil {
		return chaintypes.NewSupers()
	}
	supers := candidates.(*chaintypes.Supers)
	for i, s := range supers.Candidates {
		supers.Candidates[i].MntCount = f.dPosStatus.SuperBlockCount(cycle, s.Signer)
	}
	return supers
}

func (f *Status) Token(address arry.Address) (types.IToken, error) {
	return f.tokenStatus.Token(address)
}
