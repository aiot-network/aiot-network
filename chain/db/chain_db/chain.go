package chain_db

import (
	"fmt"
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/db/base"
	"github.com/aiot-network/aiotchain/tools/arry"
	"strconv"
)

const (
	_lastHeight   = "lastHeight"
	_header       = "header"
	_message      = "message"
	_heightHash   = "heightHash"
	_txIndex      = "txIndex"
	_actRoot      = "actRoot"
	_tokenRoot    = "tokenRoot"
	_dPosRoot     = "dPosRoot"
	_hisConfirmed = "hisConfirmed"
	_cycleHash    = "cycleHash"
)

type ChainDB struct {
	db *base.Base
}

func Open(path string) (*ChainDB, error) {
	var err error
	baseDB, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &ChainDB{db: baseDB}, nil
}

func (c *ChainDB) ActRoot() (arry.Hash, error) {
	rootBytes, err := c.db.GetFromBucket(_actRoot, []byte(_actRoot))
	if err != nil {
		return arry.Hash{}, err
	}
	return arry.BytesToHash(rootBytes), nil
}

func (c *ChainDB) DPosRoot() (arry.Hash, error) {
	rootBytes, err := c.db.GetFromBucket(_dPosRoot, []byte(_dPosRoot))
	if err != nil {
		return arry.Hash{}, err
	}
	return arry.BytesToHash(rootBytes), nil
}

func (c *ChainDB) TokenRoot() (arry.Hash, error) {
	rootBytes, err := c.db.GetFromBucket(_tokenRoot, []byte(_tokenRoot))
	if err != nil {
		return arry.Hash{}, err
	}
	return arry.BytesToHash(rootBytes), nil
}

func (c *ChainDB) LastHeight() (uint64, error) {
	bytes, err := c.db.GetFromBucket(_lastHeight, []byte(_lastHeight))
	if err != nil {
		return 0, err
	}
	height, _ := strconv.ParseUint(string(bytes), 10, 64)
	return height, nil
}

func (b *ChainDB) Close() error {
	return b.db.Close()
}

func (b *ChainDB) GetHeaderHeight(height uint64) (*types.Header, error) {
	hash, err := b.GetHashByHeight(height)
	if err != nil {
		return nil, err
	}
	return b.GetHeaderHash(hash)
}

func (b *ChainDB) GetHeaderHash(hash arry.Hash) (*types.Header, error) {
	bytes, err := b.db.GetFromBucket(_header, hash.Bytes())
	if err != nil {
		return nil, err
	}
	return types.DecodeHeader(bytes)
}

func (b *ChainDB) GetMsgIndex(hash arry.Hash) (*types.MsgIndex, error) {
	bytes, err := b.db.GetFromBucket(_txIndex, hash.Bytes())
	if err != nil {
		return nil, err
	}
	if bytes == nil || len(bytes) == 0 {
		return nil, fmt.Errorf("message %s is not exist", hash.String())
	}
	txIndex, err := types.DecodeTxIndex(bytes)
	return txIndex, err
}

func (b *ChainDB) GetMessages(txRoot arry.Hash) ([]*types.RlpMessage, error) {
	bytes, err := b.db.GetFromBucket(_message, txRoot.Bytes())
	if err != nil {
		return nil, err
	}
	return types.DecodeRlpMessages(bytes)
}

func (b *ChainDB) GetMessage(hash arry.Hash) (*types.RlpMessage, error) {
	txIndex, err := b.GetMsgIndex(hash)
	if err != nil {
		return nil, err
	}
	txs, err := b.GetMessages(txIndex.MsgRoot)
	if err != nil {
		return nil, err
	}
	return txs[txIndex.Index], nil
}

func (b *ChainDB) GetHashByHeight(height uint64) (arry.Hash, error) {
	hash, err := b.db.GetFromBucket(_heightHash, []byte(strconv.FormatUint(height, 10)))
	if err != nil {
		return arry.Hash{}, nil
	}
	return arry.BytesToHash(hash), nil
}

func (b *ChainDB) GetConfirmedHeight(height uint64) (uint64, error) {
	heightBytes, err := b.db.GetFromBucket(_hisConfirmed, []byte(strconv.FormatUint(height, 10)))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(heightBytes), 10, 64)
}

func (b *ChainDB) CycleLastHash(cycle uint64) (arry.Hash, error) {
	bytes := []byte(strconv.FormatUint(cycle, 10))
	bytes, err := b.db.GetFromBucket(_cycleHash, bytes)
	if err != nil {
		return arry.Hash{}, err
	}
	return arry.BytesToHash(bytes), nil
}

func (b *ChainDB) SaveLastHeight(height uint64) {
	bytes := []byte(strconv.FormatUint(height, 10))
	b.db.PutInBucket(_lastHeight, []byte(_lastHeight), bytes)
}

func (b *ChainDB) SaveHeader(header *types.Header) {
	b.db.PutInBucket(_header, header.Hash.Bytes(), header.Bytes())
	b.SaveHeightHash(header.Height, header.Hash)
}

func (b *ChainDB) SaveMessages(msgRoot arry.Hash, iTxs []*types.RlpMessage) {
	bytes := types.EncodeRlpMessages(iTxs)
	b.db.PutInBucket(_message, msgRoot.Bytes(), bytes)
}

func (b *ChainDB) SaveMsgIndex(msgIndexs map[arry.Hash]*types.MsgIndex) {
	for hash, loc := range msgIndexs {
		b.db.PutInBucket(_txIndex, hash.Bytes(), loc.Bytes())
	}
}

func (b *ChainDB) SaveHeightHash(height uint64, hash arry.Hash) {
	key := []byte(strconv.FormatUint(height, 10))
	b.db.PutInBucket(_heightHash, key, hash.Bytes())
}

func (b *ChainDB) SaveActRoot(hash arry.Hash) {
	b.db.PutInBucket(_actRoot, []byte(_actRoot), hash.Bytes())
}

func (b *ChainDB) SaveTokenRoot(hash arry.Hash) {
	b.db.PutInBucket(_tokenRoot, []byte(_tokenRoot), hash.Bytes())
}

func (b *ChainDB) SaveDPosRoot(hash arry.Hash) {
	b.db.PutInBucket(_dPosRoot, []byte(_dPosRoot), hash.Bytes())
}

func (b *ChainDB) SaveConfirmedHeight(height uint64, confirmed uint64) {
	heightBytes := []byte(strconv.FormatUint(height, 10))
	confirmedBytes := []byte(strconv.FormatUint(confirmed, 10))
	b.db.PutInBucket(_hisConfirmed, heightBytes, confirmedBytes)
}

func (b *ChainDB) SaveCycleLastHash(cycle uint64, hash arry.Hash) {
	bytes := []byte(strconv.FormatUint(cycle, 10))
	b.db.PutInBucket(_cycleHash, bytes, hash.Bytes())
}
