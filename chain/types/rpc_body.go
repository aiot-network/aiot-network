package types

type RpcBody struct {
	Messages []*RpcMessage `json:"transactions"`
}

func BodyToRpcBody(body *Body, statusFunc GetContractState) (*RpcBody, error) {
	var rpcMsgs []*RpcMessage
	for _, msg := range body.Messages {
		status := statusFunc(msg.Hash())
		rpcMsg, err := MsgToRpcMsgWithState(msg.(*Message), status.(*ContractStatus))
		if err != nil {
			return nil, err
		}
		rpcMsgs = append(rpcMsgs, rpcMsg)
	}
	return &RpcBody{rpcMsgs}, nil
}
