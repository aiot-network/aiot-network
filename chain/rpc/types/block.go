package types

import (
	chaintypes "github.com/aiot-network/aiot-network/chain/types"
)

type RpcBlock struct {
	RpcHeader *RpcHeader `json:"header"`
	RpcBody   *RpcBody   `json:"body"`
	Confirmed bool       `json:"confirmed"`
}

func BlockToRpcBlock(block *chaintypes.Block, confirmed uint64) (*RpcBlock, error) {
	rpcHeader := HeaderToRpcHeader(block.Header)
	rpcBody, err := BodyToRpcBody(block.Body)
	if err != nil {
		return nil, err
	}
	return &RpcBlock{
		RpcHeader: rpcHeader,
		RpcBody:   rpcBody,
		Confirmed: confirmed >= rpcHeader.Height,
	}, nil
}
