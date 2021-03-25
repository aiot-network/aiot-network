package dpos

import (
	"encoding/binary"
	"errors"
	"github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/common/blockchain"
	"github.com/aiot-network/aiot-network/common/config"
	"github.com/aiot-network/aiot-network/common/dpos"
	"github.com/aiot-network/aiot-network/common/param"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/crypto/hash"
	"math/rand"
	"sort"
)

var Err_Elected = errors.New("the election has been passed")

type Cycle struct {
	DPosStatus dpos.IDPosStatus
}

func (c *Cycle) CheckCycle(chain blockchain.IChain, preTime, time uint64) error {
	currentTerm := time / param.CycleInterval

	_, err := c.DPosStatus.CycleSupers(currentTerm)
	if err != nil {
		return nil
	}
	return Err_Elected
}

func (c *Cycle) Elect(time uint64, preHash arry.Hash, chain blockchain.IChain) error {
	curCycle := time / param.CycleInterval
	voters, err := c.calVotes(chain)
	if err != nil {
		return err
	}
	candidates := types.SortableCandidates{}
	for _, candidate := range voters {
		candidates = append(candidates, candidate)
	}
	if len(candidates) < config.Param.DPosSize {
		return errors.New("too few candidate")
	}

	sort.Sort(candidates)

	if len(candidates) > config.Param.SuperSize {
		candidates = candidates[:config.Param.SuperSize]
	}

	// Use the last block hash of the last cycle as a random number seed
	// to ensure that the election results of each node are consistent
	seed := int64(binary.LittleEndian.Uint32(hash.Hash(preHash.Bytes()).Bytes())) + int64(curCycle)
	r := rand.New(rand.NewSource(seed))
	for i := len(candidates) - 1; i > 0; i-- {
		j := int(r.Int31n(int32(i + 1)))
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}

	supers := &types.Supers{Candidates: candidates, PreHash: preHash}
	c.DPosStatus.SaveCycle(curCycle, supers)
	return nil
}

func (c *Cycle) calVotes(chain blockchain.IChain) ([]*types.Member, error) {
	iCans, err := c.DPosStatus.Candidates()
	if err != nil {
		return nil, errors.New("no candidate")
	}
	if iCans.Len() < config.Param.SuperSize {
		return nil, errors.New("not enough candidates")
	}
	cans := iCans.(*types.Candidates)
	voterMap := c.DPosStatus.Voters()
	for index, candidate := range cans.Members {
		voters, ok := voterMap[candidate.Signer]
		if ok {
			for _, voter := range voters {
				cans.Members[index].Weight += chain.Vote(voter)
			}
		}
	}
	return cans.Members, nil
}
