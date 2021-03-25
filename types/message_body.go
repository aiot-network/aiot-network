package types

import "github.com/aiot-network/aiot-network/tools/arry"

type IMessageBody interface {
	MsgTo() IReceiver
	MsgToken() arry.Address
	MsgAmount() uint64
	CheckBody(from arry.Address) error
}

type IReceiver interface {
	ReceiverList() []*Receiver
}

type Receiver struct {
	Address arry.Address
	Amount  uint64
}
