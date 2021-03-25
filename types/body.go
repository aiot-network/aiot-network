package types

type IBody interface {
	MsgList() []IMessage
	Count() int
}
