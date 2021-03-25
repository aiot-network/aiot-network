package blockchain

import (
	"github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/tools/arry"
)

type IChainDB interface {
	ActRoot() (arry.Hash, error)
	DPosRoot() (arry.Hash, error)
	TokenRoot() (arry.Hash, error)
	LastHeight() (uint64, error)
	GetMessage(hash arry.Hash) (*types.RlpMessage, error)
	GetMessages(txRoot arry.Hash) ([]*types.RlpMessage, error)
	GetMsgIndex(hash arry.Hash) (*types.MsgIndex, error)
	GetHeaderHeight(height uint64) (*types.Header, error)
	GetHeaderHash(hash arry.Hash) (*types.Header, error)
	GetConfirmedHeight(height uint64) (uint64, error)
	CycleLastHash(cycle uint64) (arry.Hash, error)

	SaveHeader(header *types.Header)
	SaveLastHeight(height uint64)
	SaveMessages(msgRoot arry.Hash, iTxs []*types.RlpMessage)
	SaveMsgIndex(msgIndexs map[arry.Hash]*types.MsgIndex)
	SaveHeightHash(height uint64, hash arry.Hash)
	SaveActRoot(hash arry.Hash)
	SaveTokenRoot(hash arry.Hash)
	SaveDPosRoot(hash arry.Hash)
	SaveConfirmedHeight(height uint64, confirmed uint64)
	SaveCycleLastHash(cycle uint64, hash arry.Hash)
}
