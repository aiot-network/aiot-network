package token_status

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/tools/arry"
)

type ITokenDB interface {
	SetRoot(hash arry.Hash) error
	Root() arry.Hash
	Commit() (arry.Hash, error)
	Token(addr arry.Address) *types.TokenRecord
	SetToken(token *types.TokenRecord)
	Contract(addr arry.Address) *status.Contract
	SetContract(contract *status.Contract)
	ContractState(msgHash arry.Hash) *types.ContractStatus
	SetContractState(msgHash arry.Hash, state *types.ContractStatus)
	SymbolContract(symbol string) (arry.Address, bool)
	SetSymbolContract(symbol string, address arry.Address)
}
