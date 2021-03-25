package types

import (
	"github.com/aiot-network/aiot-network/chain/types"
	"time"
)

type RpcHeader struct {
	Version   uint32              `json:"version"`
	Hash      string              `json:"hash"`
	PreHash   string              `json:"parenthash"`
	MsgRoot   string              `json:"txroot"`
	ActRoot   string              `json:"actroot"`
	TokenRoot string              `json:"tokenroot"`
	DPosRoot  string              `json:"dposroot"`
	Height    uint64              `json:"height"`
	Time      time.Time           `json:"time"`
	Cycle     uint64              `json:"cycle"`
	Signer    string              `json:"signer"`
	Signature *types.RpcSignature `json:"signature"`
}

func HeaderToRpcHeader(header *types.Header) *RpcHeader {
	return &RpcHeader{
		Version:   header.Version,
		Hash:      header.Hash.String(),
		PreHash:   header.PreHash.String(),
		MsgRoot:   header.MsgRoot.String(),
		ActRoot:   header.ActRoot.String(),
		TokenRoot: header.TokenRoot.String(),
		DPosRoot:  header.DPosRoot.String(),
		Height:    header.Height,
		Time:      time.Unix(int64(header.Time), 0),
		Cycle:     header.Cycle,
		Signer:    header.Signer.String(),
		Signature: &types.RpcSignature{
			Signature: header.Signature.SignatureString(),
			PubKey:    header.Signature.PubKeyString(),
		},
	}
}
