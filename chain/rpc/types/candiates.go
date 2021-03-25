package types

import (
	chaintypes "github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/types"
)

type RpcMember struct {
	Signer   string `json:"address"`
	PeerId   string `json:"peerid"`
	Weight   uint64 `json:"votes"`
	MntCount uint32 `json:"mntcount"`
}

type RpcCandidates struct {
	Members []*RpcMember `json:"members"`
}

func CandidatesToRpcCandidates(candidates types.ICandidates) *RpcCandidates {
	rpcMems := &RpcCandidates{Members: make([]*RpcMember, 0)}
	cas := candidates.(*chaintypes.Candidates)
	for _, candidate := range cas.Members {
		rpcMem := &RpcMember{
			Signer: candidate.Signer.String(),
			PeerId: candidate.PeerId,
			Weight: candidate.Weight,
		}
		rpcMems.Members = append(rpcMems.Members, rpcMem)
	}
	return rpcMems
}

func SupersToRpcCandidates(candidates types.ICandidates) *RpcCandidates {
	rpcMems := &RpcCandidates{Members: make([]*RpcMember, 0)}
	supers := candidates.(*chaintypes.Supers)
	for _, candidate := range supers.Candidates {
		rpcMem := &RpcMember{
			Signer:   candidate.Signer.String(),
			PeerId:   candidate.PeerId,
			Weight:   candidate.Weight,
			MntCount: candidate.MntCount,
		}
		rpcMems.Members = append(rpcMems.Members, rpcMem)
	}
	return rpcMems
}
