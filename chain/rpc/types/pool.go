package types

import (
	chaintypes "github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/types"
)

type TxPool struct {
	MsgsCount  int                      `json:"msgs"`
	ReadyCount int                      `json:"ready"`
	CacheCount int                      `json:"cache"`
	ReadyMsgs  []*chaintypes.RpcMessage `json:"readymsgs"`
	CacheMsgs  []*chaintypes.RpcMessage `json:"cachemsgs"`
}

func MsgsToRpcMsgsPool(readyMsgs []types.IMessage, cacheMsgs []types.IMessage) *TxPool {
	var readyRpcMsgs, cacheRpcMsgs []*chaintypes.RpcMessage
	for _, msg := range readyMsgs {
		rpcMsg, _ := chaintypes.MsgToRpcMsg(msg.(*chaintypes.Message))
		readyRpcMsgs = append(readyRpcMsgs, rpcMsg)
	}

	for _, msg := range cacheMsgs {
		rpcMsg, _ := chaintypes.MsgToRpcMsg(msg.(*chaintypes.Message))
		cacheRpcMsgs = append(cacheRpcMsgs, rpcMsg)
	}

	readyCount := len(readyRpcMsgs)
	cacheCount := len(cacheRpcMsgs)

	return &TxPool{
		MsgsCount:  readyCount + cacheCount,
		ReadyCount: readyCount,
		CacheCount: cacheCount,
		ReadyMsgs:  readyRpcMsgs,
		CacheMsgs:  cacheRpcMsgs,
	}
}
