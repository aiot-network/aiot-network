package types

import "github.com/aiot-network/aiotchain/tools/arry"

type IMessageBody interface {
	MsgTo() IReceiver
	MsgContract() arry.Address
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
