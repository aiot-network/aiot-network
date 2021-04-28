package types

import (
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	"github.com/aiot-network/aiotchain/types"
)

type Block struct {
	*Header
	*Body
}

func (b *Block) Hash() arry.Hash {
	return b.Header.Hash
}

func (b *Block) MsgRoot() arry.Hash {
	return b.Header.MsgRoot
}

func (b *Block) Signer() arry.Address {
	return b.Header.Signer
}

func (b *Block) Height() uint64 {
	return b.Header.Height
}

func (b *Block) Time() int64 {
	return int64(b.Header.Time)
}

func (b *Block) Add(message types.IMessage) {
	b.Body.Messages = append(b.Body.Messages, message)
}

func (b *Block) Msgs() []types.IMessage {
	return b.Body.Msgs()
}

func (b *Block) ToRlpHeader() types.IRlpHeader {
	panic("implement me")
}

func (b *Block) BlockHeader() types.IHeader {
	return b.Header
}

func (b *Block) BlockBody() types.IBody {
	return b.Body
}

func (b *Block) ToRlpBlock() types.IRlpBlock {
	return &RlpBlock{
		RlpHeader: b.Header,
		RlpBody:   b.Body.ToRlpBody(),
	}
}

func (b *Block) SetHash() {
	b.Header.SetHash()
}

func (b *Block) Sign(key *secp256k1.PrivateKey) error {
	return b.Header.Sign(key)
}

func (b *Block) CheckMsgRoot() bool {
	msgRoot := MsgRoot(b.Body.Messages)
	return b.Header.MsgRoot.IsEqual(msgRoot)
}

func (b *Block) GetMsgIndexs() map[arry.Hash]*MsgIndex {
	mapLocation := make(map[arry.Hash]*MsgIndex)
	for index, tx := range b.MsgList() {
		mapLocation[tx.Hash()] = &MsgIndex{
			MsgRoot: b.MsgRoot(),
			Index:   uint32(index),
			Height:  b.Height(),
		}
	}
	return mapLocation
}

type Blocks []*Block

func (b Blocks) Blocks() []*Block {
	return b
}
