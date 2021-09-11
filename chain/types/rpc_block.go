package types

type RpcBlock struct {
	RpcHeader *RpcHeader `json:"header"`
	RpcBody   *RpcBody   `json:"body"`
	Confirmed bool       `json:"confirmed"`
}

func BlockToRpcBlock(block *Block, confirmed uint64, stateFunc GetContractState) (*RpcBlock, error) {
	rpcHeader := HeaderToRpcHeader(block.Header)
	rpcBody, err := BodyToRpcBody(block.Body, stateFunc)
	if err != nil {
		return nil, err
	}
	return &RpcBlock{
		RpcHeader: rpcHeader,
		RpcBody:   rpcBody,
		Confirmed: confirmed >= rpcHeader.Height,
	}, nil
}
