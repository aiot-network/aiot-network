package types

import (
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/rlp"
	"github.com/aiot-network/aiot-network/types"
)

// Super nodes
type Supers struct {
	Candidates []*Member
	PreHash    arry.Hash
}

func NewSupers() *Supers {
	return &Supers{Candidates: make([]*Member, 0)}
}

func (s *Supers) Len() int {
	return len(s.Candidates)
}

func (s *Supers) List() []types.ICandidate {
	iCans := make([]types.ICandidate, s.Len())
	for i, mem := range s.Candidates {
		iCans[i] = mem
	}
	return iCans
}

func (s *Supers) Get(i int) types.ICandidate {
	return s.Candidates[i]
}

func (s *Supers) GetPreHash() arry.Hash {
	return s.PreHash
}

type Member struct {
	Signer   arry.Address
	PeerId   string
	Weight   uint64
	MntCount uint32
	Voters   []arry.Address
}

func (m *Member) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(m)
	return bytes
}

func (m *Member) GetPeerId() string {
	return m.PeerId
}

func (m *Member) GetSinger() arry.Address {
	return m.Signer
}

func DecodeMember(bytes []byte) (*Member, error) {
	var mem *Member
	err := rlp.DecodeBytes(bytes, &mem)
	if err != nil {
		return nil, err
	}
	return mem, nil
}

type Candidates struct {
	Members []*Member
}

func NewCandidates() *Candidates {
	return &Candidates{Members: make([]*Member, 0)}
}

func (c *Candidates) Set(newMem *Member) {
	c.Members = append(c.Members, newMem)
}

func (c *Candidates) Remove(reMem *Member) {
	for i, mem := range c.Members {
		if mem.Signer.IsEqual(reMem.Signer) {
			c.Members = append(c.Members[0:i], c.Members[i+1:]...)
			return
		}
	}
}

func (c *Candidates) List() []types.ICandidate {
	iCans := make([]types.ICandidate, c.Len())
	for i, mem := range c.Members {
		iCans[i] = mem
	}
	return iCans
}

func (c *Candidates) Len() int {
	return len(c.Members)
}

func (c *Candidates) Get(i int) types.ICandidate {
	return c.Members[i]
}

func (c *Candidates) GetPreHash() arry.Hash {
	return arry.Hash{}
}

type SortableCandidates []*Member

func (p SortableCandidates) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p SortableCandidates) Len() int      { return len(p) }
func (p SortableCandidates) Less(i, j int) bool {
	if p[i].Weight < p[j].Weight {
		return false
	} else if p[i].Weight > p[j].Weight {
		return true
	} else {
		return p[i].Weight < p[j].Weight
	}
}
