package types

import (
	"github.com/aiot-network/aiot-network/tools/rlp"
)

func DecodeRlpMessages(bytes []byte) ([]*RlpMessage, error) {
	var msgs []*RlpMessage
	err := rlp.DecodeBytes(bytes, &msgs)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func EncodeRlpMessages(msgs []*RlpMessage) []byte {
	bytes, _ := rlp.EncodeToBytes(msgs)
	return bytes
}
