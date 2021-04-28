package types

import (
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	hash2 "github.com/aiot-network/aiotchain/tools/crypto/hash"
	"github.com/aiot-network/aiotchain/tools/rlp"
	"github.com/aiot-network/aiotchain/types"
)

type Header struct {
	Version   uint32
	Hash      arry.Hash
	PreHash   arry.Hash
	MsgRoot   arry.Hash
	ActRoot   arry.Hash
	DPosRoot  arry.Hash
	TokenRoot arry.Hash
	Height    uint64
	Time      uint64
	Cycle     uint64
	Signer    arry.Address
	Signature *Signature
}

func NewHeader(preHash, msgRoot, actRoot, dPosRoot, tokenRoot arry.Hash, height uint64,
	blockTime uint64, signer arry.Address) *Header {
	return &Header{
		PreHash:   preHash,
		MsgRoot:   msgRoot,
		ActRoot:   actRoot,
		DPosRoot:  dPosRoot,
		TokenRoot: tokenRoot,
		Height:    height,
		Time:      blockTime,
		Cycle:     blockTime / uint64(param.CycleInterval),
		Signer:    signer,
		Signature: &Signature{},
	}
}

func DecodeHeader(bytes []byte) (*Header, error) {
	var h = new(Header)
	if err := rlp.DecodeBytes(bytes, h); err != nil {
		return h, err
	}
	return h, nil
}

func (h *Header) GetSigner() arry.Address {
	return h.Signer
}

func (h *Header) GetHash() arry.Hash {
	return h.Hash
}

func (h *Header) GetPreHash() arry.Hash {
	return h.PreHash
}

func (h *Header) Bytes() []byte {
	bytes, _ := rlp.EncodeToBytes(h)
	return bytes
}

func (h *Header) GetHeight() uint64 {
	return h.Height
}

func (h *Header) GetMsgRoot() arry.Hash {
	return h.MsgRoot
}

func (h *Header) GetActRoot() arry.Hash {
	return h.ActRoot
}

func (h *Header) GetDPosRoot() arry.Hash {
	return h.DPosRoot
}

func (h *Header) GetTokenRoot() arry.Hash {
	return h.TokenRoot
}

func (h *Header) GetSignature() types.ISignature {
	return h.Signature
}

func (h *Header) GetTime() uint64 {
	return h.Time
}

func (h *Header) GetCycle() uint64 {
	return h.Cycle
}

func (h *Header) SetHash() {
	h.Hash = hash2.Hash(h.Bytes())
}

func (h *Header) Sign(key *secp256k1.PrivateKey) error {
	sig, err := Sign(key, h.Hash)
	if err != nil {
		return err
	}
	h.Signature = sig
	return nil
}

func (h *Header) ToRlpHeader() types.IRlpHeader {
	return h
}
