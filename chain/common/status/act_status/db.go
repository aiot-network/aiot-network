package act_status

import (
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/types"
)

type IActDB interface {
	SetRoot(hash arry.Hash) error
	Root() arry.Hash
	Commit() (arry.Hash, error)
	Close() error
	Account(address arry.Address) types.IAccount
	SetAccount(account types.IAccount)
	Nonce(address arry.Address) uint64
}
