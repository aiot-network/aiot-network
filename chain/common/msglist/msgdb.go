package msglist

import "github.com/aiot-network/aiotchain/types"

type ITxListDB interface {
	Read() []types.IMessage
	Save(message types.IMessage)
	Delete(msg types.IMessage)
	Clear()
	Close() error
}
