package types

import (
	"github.com/aiot-network/aiotchain/tools/arry"
)

type IToken interface {
	Symbol() string
}

type IContract interface {
	Bytes() []byte
}

type IStatus interface {
}

type ITokenStatus interface {
	SetTrieRoot(hash arry.Hash) error
	TrieRoot() arry.Hash
	CheckMessage(msg IMessage) error
	UpdateToken(msg IMessage, height uint64) error
	Token(address arry.Address) (IToken, error)
	Contract(address arry.Address) (IContract, error)
	ContractState(address arry.Hash) IStatus
	TokenList() []map[string]string
	SymbolContract(symbol string) (arry.Address, bool)
	Commit() (arry.Hash, error)
}
