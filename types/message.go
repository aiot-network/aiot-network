package types

type IMessage interface {
	IMessageHeader
	MsgBody() IMessageBody
	ToRlp() IRlpMessage
	Check() error
}

type IMessageIndex interface {
	GetHeight() uint64
}
