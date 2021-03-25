package request

import (
	"errors"
	"github.com/aiot-network/aiot-network/server"
	"github.com/aiot-network/aiot-network/types"
	"github.com/libp2p/go-libp2p-core/network"
)

var (
	Err_BlockNotFound = errors.New("block not exist")
	Err_PeerClosed    = errors.New("peer has closed")
)

type IRequestHandler interface {
	server.IService
	ISend
	IRegister
	IResponse
}

type ISend interface {
	LastHeight(conn *types.Conn) (uint64, error)
	SendMsg(conn *types.Conn, msg types.IMessage) error
	SendBlock(conn *types.Conn, block types.IBlock) error
	GetBlocks(conn *types.Conn, height, count uint64) ([]types.IBlock, error)
	GetBlock(conn *types.Conn, height uint64) (types.IBlock, error)
	IsEqual(conn *types.Conn, header types.IHeader) (bool, error)
	LocalInfo(conn *types.Conn) (*types.Local, error)
}

type IRegister interface {
	RegisterReceiveBlock(func(types.IBlock) error)
	RegisterReceiveMessage(func(types.IMessage) error)
}

type IResponse interface {
	SendToReady(stream network.Stream)
}
