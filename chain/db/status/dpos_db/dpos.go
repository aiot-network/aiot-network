package dpos_db

import (
	"bytes"
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/db/base"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/hash"
	"github.com/aiot-network/aiotchain/tools/rlp"
	"github.com/aiot-network/aiotchain/tools/trie"
	"strconv"
)

const (
	_cycleSupers    = "cycleSupers"
	_candidates     = "candidates"
	_voters         = "voters"
	_confirmed      = "confirmed"
	_blockCount     = "blockCount"
	_coinBaseCount  = "coinBaseCount"
	_superWork      = "superWork"
)

type DPosDB struct {
	base *base.Base
	trie *trie.Trie
}

func Open(path string) (*DPosDB, error) {
	var err error
	baseDB, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &DPosDB{base: baseDB}, nil
}

func (d *DPosDB) SetRoot(hash arry.Hash) error {
	t, err := trie.New(hash, d.base)
	if err != nil {
		return err
	}
	d.trie = t
	return nil
}

func (d *DPosDB) Root() arry.Hash {
	return d.trie.Hash()
}

func (d *DPosDB) Commit() (arry.Hash, error) {
	return d.trie.Commit()
}

func (d *DPosDB) Confirmed() (uint64, error) {
	bytes := d.trie.Get(base.Key(_confirmed, []byte(_confirmed)))
	var height uint64
	rlp.DecodeBytes(bytes, &height)
	return height, nil
}

func (d *DPosDB) SetConfirmed(height uint64) {
	bytes, _ := rlp.EncodeToBytes(height)
	d.trie.Update(base.Key(_confirmed, []byte(_confirmed)), bytes)
}

func (d *DPosDB) CandidatesCount() int {
	cans := types.NewCandidates()
	iter := d.trie.PrefixIterator(base.Prefix(_voters))
	for iter.Next(true) {
		if iter.Leaf() {
			value := iter.LeafBlob()
			mem, _ := types.DecodeMember(value)
			cans.Set(mem)
		}
	}
	return cans.Len()
}

func (d *DPosDB) AddCandidate(member *types.Member) {
	d.trie.Update(base.Key(_candidates, member.Signer.Bytes()), member.Bytes())
}

func (d *DPosDB) CancelCandidate(signer arry.Address) {
	d.trie.Delete(base.Key(_candidates, signer.Bytes()))
}

func (d *DPosDB) CycleSupers(cycle uint64) (*types.Supers, error) {
	var supers *types.Supers
	cycleBytes, err := rlp.EncodeToBytes(cycle)
	if err != nil {
		return nil, err
	}
	bytes, err := d.base.GetFromBucket(_cycleSupers, cycleBytes)
	if err := rlp.DecodeBytes(bytes, &supers); err != nil {
		return nil, err
	}
	return supers, nil
}

func (d *DPosDB) SaveCycle(cycle uint64, supers *types.Supers) {
	value, _ := rlp.EncodeToBytes(supers)
	key, _ := rlp.EncodeToBytes(cycle)
	d.base.PutInBucket(_cycleSupers, key, value)
}

func (d *DPosDB) Candidates() (*types.Candidates, error) {
	cans := types.NewCandidates()
	iter := d.trie.PrefixIterator(base.Prefix(_candidates))
	for iter.Next(true) {
		if iter.Leaf() {
			value := iter.LeafBlob()
			mem, _ := types.DecodeMember(value)
			cans.Set(mem)
		}
	}
	return cans, nil
}

func (d *DPosDB) Voters() map[arry.Address][]arry.Address {
	rs := make(map[arry.Address][]arry.Address)
	iter := d.trie.PrefixIterator(base.Prefix(_voters))
	for iter.Next(true) {
		if iter.Leaf() {
			key := iter.LeafKey()
			from := arry.BytesToAddress(base.LeafKeyToKey(_voters, key))
			value := iter.LeafBlob()
			to := arry.BytesToAddress(value)
			addrs, ok := rs[to]
			if !ok {
				rs[to] = []arry.Address{from}
			} else {
				rs[to] = append(addrs, from)
			}
		}
	}
	return rs
}

func (d *DPosDB) Voter(from, to arry.Address) {
	d.trie.Update(base.Key(_voters, from.Bytes()), to.Bytes())
}

func (d *DPosDB) AddSuperBlockCount(cycle uint64, signer arry.Address) {
	hash := cycleSuperCountKey(cycle, signer)
	cnt := d.SuperBlockCount(cycle, signer)
	cnt++
	bytes, _ := rlp.EncodeToBytes(cnt)
	d.trie.Update(base.Key(_blockCount, hash.Bytes()), bytes)
}

func (d *DPosDB) SuperBlockCount(cycle uint64, signer arry.Address) uint32 {
	hash := cycleSuperCountKey(cycle, signer)
	bytes := d.trie.Get(base.Key(_blockCount, hash.Bytes()))
	var count uint32
	rlp.DecodeBytes(bytes, &count)
	return count
}

func (d *DPosDB) AddCoinBaseCount(cycle uint64, address arry.Address) {
	hash := cycleSuperCountKey(cycle, address)
	cnt := d.CoinBaseCount(cycle, address)
	cnt++
	bytes, _ := rlp.EncodeToBytes(cnt)
	d.trie.Update(base.Key(_coinBaseCount, hash.Bytes()), bytes)
}

func (d *DPosDB) CoinBaseCount(cycle uint64, address arry.Address) uint32 {
	hash := cycleSuperCountKey(cycle, address)
	bytes := d.trie.Get(base.Key(_coinBaseCount, hash.Bytes()))
	var count uint32
	rlp.DecodeBytes(bytes, &count)
	return count
}


func cycleSuperCountKey(cycle uint64, signer arry.Address) arry.Hash {
	bytes := bytes.Join([][]byte{[]byte(strconv.FormatUint(cycle, 10)), signer.Bytes()}, []byte{})
	return hash.Hash(bytes)
}

func (d *DPosDB) AddSuperWork(cycle uint64, super arry.Address, works *types.Works) {
	hash := cycleSuperCountKey(cycle, super)
	workLast, err := d.SuperWork(cycle, super)
	if err == nil {
		works.WorkLoad += workLast.WorkLoad
	}
	bytes, _ := rlp.EncodeToBytes(works)
	d.trie.Update(base.Key(_superWork, hash.Bytes()), bytes)
}

func (d *DPosDB) SuperWork(cycle uint64, super arry.Address) (*types.Works, error) {
	hash := cycleSuperCountKey(cycle, super)
	bytes := d.trie.Get(base.Key(_superWork, hash.Bytes()))
	var work *types.Works
	err := rlp.DecodeBytes(bytes, &work)
	if err != nil {
		return nil, err
	}
	return work, nil
}
