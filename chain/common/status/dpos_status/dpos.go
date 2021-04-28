package dpos_status

import (
	"fmt"
	"github.com/aiot-network/aiotchain/chain/db/status/dpos_db"
	chaintypes "github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/types"
)

const dPosDB = "dpos_db"

type DPosStatus struct {
	db IDPosDB
}

func NewDPosStatus() (*DPosStatus, error) {
	db, err := dpos_db.Open(config.Param.Data + "/" + dPosDB)
	if err != nil {
		return nil, err
	}

	return &DPosStatus{db: db}, nil
}

func (d *DPosStatus) SetTrieRoot(hash arry.Hash) error {
	return d.db.SetRoot(hash)
}

func (d *DPosStatus) TrieRoot() arry.Hash {
	return d.db.Root()
}

func (d *DPosStatus) Commit() (arry.Hash, error) {
	return d.db.Commit()
}

// If the current number of candidates is less than or equal to the
// number of super nodes, it is not allowed to withdraw candidates.
func (d *DPosStatus) CheckMessage(msg types.IMessage) error {
	switch chaintypes.MessageType(msg.Type()) {
	case chaintypes.Cancel:
		if d.db.CandidatesCount() <= config.Param.SuperSize {
			return fmt.Errorf("candidate nodes are already in the minimum number. Cannot cancel the candidate status now, please wait")
		}
	}
	return nil
}

func (d *DPosStatus) UpdateWork(msg types.IMessage) error {
	body, ok := msg.MsgBody().(*chaintypes.WorkBody)
	if !ok {
		return fmt.Errorf("incrrect message type")
	}
	cycle := msg.Time() / param.CycleInterval
	for _, work := range body.List {
		d.db.AddSuperWork(cycle, work.Address, &chaintypes.Works{
			Cycle:    cycle,
			WorkLoad: work.Workload,
			EndTime:  work.EndTime,
		})
	}
	return nil
}

func (d *DPosStatus) CycleSupers(cycle uint64) (types.ICandidates, error) {
	return d.db.CycleSupers(cycle)
}

func (d *DPosStatus) SaveCycle(cycle uint64, supers types.ICandidates) {
	ss := supers.(*chaintypes.Supers)
	d.db.SaveCycle(cycle, ss)
}

func (d *DPosStatus) Candidates() (types.ICandidates, error) {
	return d.db.Candidates()
}

func (d *DPosStatus) Voters() map[arry.Address][]arry.Address {
	return d.db.Voters()
}

func (d *DPosStatus) Confirmed() (uint64, error) {
	return d.db.Confirmed()
}

func (d *DPosStatus) SetConfirmed(height uint64) {
	d.db.SetConfirmed(height)
}

func (d *DPosStatus) AddCandidate(msg types.IMessage) error {
	body := msg.MsgBody().(*chaintypes.CandidateBody)
	candidate := &chaintypes.Member{
		Signer: msg.From(),
		PeerId: body.Peer.String(),
		Weight: 0,
	}
	d.db.AddCandidate(candidate)
	d.db.Voter(msg.From(), msg.From())
	return nil
}

func (d *DPosStatus) CancelCandidate(msg types.IMessage) error {
	d.db.CancelCandidate(msg.From())
	return nil
}

func (d *DPosStatus) Voter(msg types.IMessage) error {
	receis := msg.MsgBody().MsgTo().ReceiverList()
	if len(receis) == 0 {
		return fmt.Errorf("not receiver")
	}
	d.db.Voter(msg.From(), msg.MsgBody().MsgTo().ReceiverList()[0].Address)
	return nil
}

func (d *DPosStatus) AddSuperBlockCount(cycle uint64, signer arry.Address) {
	d.db.AddSuperBlockCount(cycle, signer)
}

func (d *DPosStatus) SuperBlockCount(cycle uint64, signer arry.Address) uint32 {
	return d.db.SuperBlockCount(cycle, signer)
}
func (d *DPosStatus) AddSuperWork(cycle uint64, super arry.Address, works types.IWorks) {
	d.db.AddSuperWork(cycle, super, works.(*chaintypes.Works))
}
func (d *DPosStatus) SuperWork(cycle uint64, super arry.Address) (types.IWorks, error) {
	return d.db.SuperWork(cycle, super)
}
