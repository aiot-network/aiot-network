package validator

import "github.com/aiot-network/aiotchain/types"

type IValidator interface {
	CheckMsg(types.IMessage, bool, uint64) error
}
