package token_status

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/tools/arry"
)

type ITokenDB interface {
	SetRoot(hash arry.Hash) error
	Root() arry.Hash
	Commit() (arry.Hash, error)
	Token(addr arry.Address) *types.TokenRecord
	SetToken(token *types.TokenRecord)
}
