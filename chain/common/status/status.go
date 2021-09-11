package status

import (
	"errors"
	"github.com/aiot-network/aiotchain/chain/common/kit"
	"github.com/aiot-network/aiotchain/chain/runner"
	chaintypes "github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/dpos"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/types"
)

const module = "chain"

type lastHeightFunc func() uint64

type Status struct {
	actStatus   types.IActStatus
	dPosStatus  dpos.IDPosStatus
	tokenStatus types.ITokenStatus
	runner      *runner.ContractRunner
}

func NewStatus(actStatus types.IActStatus, dPosStatus dpos.IDPosStatus, tokenStatus types.ITokenStatus, runner *runner.ContractRunner) *Status {
	return &Status{
		actStatus:   actStatus,
		dPosStatus:  dPosStatus,
		tokenStatus: tokenStatus,
		runner:      runner,
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

func (f *Status) CheckMsg(msg types.IMessage, strict bool, height uint64) error {
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

	if err := f.runner.Verify(msg, height); err != nil {
		return err
	}
	return nil
}

func (f *Status) CheckBlockMsg(msg types.IMessage, strict bool) error {
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
	coinBaseAddr := arry.Address{}
	for _, msg := range msgs {
		if msg.IsCoinBase() {
			coinBaseAddr = msg.MsgBody().MsgTo().ReceiverList()[0].Address
		}
		if err := f.actStatus.FromMessage(msg, block.GetHeight()); err != nil {
			return err
		}
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
		case chaintypes.TokenV2:
			if err := f.actStatus.ToMessage(msg, block.GetHeight()); err != nil {
				return err
			}
			if err := f.tokenStatus.UpdateToken(msg, block.GetHeight()); err != nil {
				return err
			}
		case chaintypes.Redemption:
			if err := f.actStatus.ToMessage(msg, block.GetHeight()); err != nil {
				return err
			}
			if err := f.tokenStatus.UpdateToken(msg, block.GetHeight()); err != nil {
				return err
			}
		case chaintypes.Contract:
			if err := f.runner.RunContract(msg, block.GetHeight(), block.GetTime()); err != nil {
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
			if err := f.dPosStatus.UpdateWork(msg); err != nil {
				return err
			}
		default:
			return errors.New("wrong message type")
		}

	}
	f.dPosStatus.AddSuperBlockCount(block.GetCycle(), block.GetSigner())
	f.dPosStatus.AddCoinBaseCount(block.GetCycle(), coinBaseAddr)
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

func (f *Status) CycleReword(cycle uint64) []types.IReword {
	rewords := make([]*chaintypes.Reword, 0)
	var allWork uint64
	for _, info := range *config.Param.CoinBaseAddressList {
		work, err := f.dPosStatus.AddressWork(cycle-1, arry.StringToAddress(info.Address))
		if err != nil {
			continue
		}
		allWork += work.GetWorkLoad()
		rewords = append(rewords, &chaintypes.Reword{
			Cycle:    cycle,
			EndTime:  work.GetEndTime(),
			Address:  info.Address,
			WorkLoad: work.GetWorkLoad(),
			Blocks:   uint64(f.dPosStatus.CoinBaseCount(cycle, arry.StringToAddress(info.Address))),
		})
	}
	for i, reword := range rewords {
		amount := kit.CalCoinBase(config.Param.NetWork, allWork, reword.GetWorkLoad()) * reword.Blocks
		rewords[i].Amount = amount
	}
	iReword := make([]types.IReword, len(rewords))
	for i, r := range rewords {
		iReword[i] = r
	}
	return iReword
}

func (f *Status) CycleWork(cycle uint64, address arry.Address) (types.IWorks, error) {
	return f.dPosStatus.AddressWork(cycle, address)
}

func (f *Status) Token(address arry.Address) (types.IToken, error) {
	return f.tokenStatus.Token(address)
}

func (f *Status) TokenList() []map[string]string {
	return f.tokenStatus.TokenList()
}

func (f *Status) SymbolContract(symbol string) (arry.Address, bool) {
	return f.tokenStatus.SymbolContract(symbol)
}

func (f *Status) Contract(address arry.Address) (types.IContract, error) {
	return f.tokenStatus.Contract(address)
}

func (f *Status) ContractState(msgHash arry.Hash) types.IStatus {
	return f.tokenStatus.ContractState(msgHash)
}
