package types

import "github.com/aiot-network/aiotchain/chain/types"

type RpcBody struct {
	Messages []*types.RpcMessage `json:"transactions"`
}

func BodyToRpcBody(body *types.Body) (*RpcBody, error) {
	var rpcMsgs []*types.RpcMessage
	for _, msg := range body.Messages {
		rpcMsg, err := types.MsgToRpcMsg(msg.(*types.Message))
		if err != nil {
			return nil, err
		}
		rpcMsgs = append(rpcMsgs, rpcMsg)
	}
	return &RpcBody{rpcMsgs}, nil
}
