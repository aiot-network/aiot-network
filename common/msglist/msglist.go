package msglist

import "github.com/aiot-network/aiot-network/types"

type IMsgList interface {
	DeleteExpired(int64)
	DeleteEnd(types.IMessage)
	Delete(types.IMessage)
	Read() error
	Close() error
	Update()
	Exist(types.IMessage) bool
	Put(types.IMessage) error
	NeedPackaged(count int) []types.IMessage
	StagnantMsgs() []types.IMessage
	GetAll() ([]types.IMessage, []types.IMessage)
	Count() int
}
