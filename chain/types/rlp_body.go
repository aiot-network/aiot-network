package types

import "github.com/aiot-network/aiotchain/types"

type RlpBody struct {
	Msgs []*RlpMessage
}

func (r *RlpBody) ToBody() *Body {
	msgs := make([]types.IMessage, len(r.Msgs))
	for i, msg := range r.Msgs {
		msgs[i] = msg.ToMessage()
	}
	return &Body{msgs}
}

func (r *RlpBody) MsgList() []*RlpMessage {
	return r.Msgs
}
