package types

import (
	"github.com/aiot-network/aiot-network/tools/rlp"
	"github.com/aiot-network/aiot-network/types"
)

type RlpBlock struct {
	RlpHeader *Header
	RlpBody   *RlpBody
}

func (r *RlpBlock) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(r)
	return bytes
}

func (r *RlpBlock) ToBlock() types.IBlock {
	return &Block{
		Header: r.RlpHeader,
		Body:   r.RlpBody.ToBody(),
	}
	return nil
}

func DecodeRlpBlock(bytes []byte) (*RlpBlock, error) {
	var rlpBlock *RlpBlock
	if err := rlp.DecodeBytes(bytes, &rlpBlock); err != nil {
		return nil, err
	}
	return rlpBlock, nil
}

func RlpBlocksToBlocks(rlpBlocks []*RlpBlock) []types.IBlock {
	rs := make([]types.IBlock, len(rlpBlocks))
	for i, rlpBlock := range rlpBlocks {
		rs[i] = rlpBlock.ToBlock()
	}
	return rs
}

func EncodeRlpBlocks(rlpBlocks []*RlpBlock) ([]byte, error) {
	return rlp.EncodeToBytes(rlpBlocks)
}

func DecodeRlpBlocks(bytes []byte) ([]*RlpBlock, error) {
	var rlpBlocks []*RlpBlock
	err := rlp.DecodeBytes(bytes, &rlpBlocks)
	if err != nil {
		return nil, err
	}
	return rlpBlocks, nil
}
