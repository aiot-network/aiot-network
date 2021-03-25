package validator

import "github.com/aiot-network/aiot-network/types"

type IValidator interface {
	CheckMsg(types.IMessage, bool) error
}
