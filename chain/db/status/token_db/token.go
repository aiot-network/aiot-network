package token_db

import (
	"github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/chain/types/status"
	"github.com/aiot-network/aiotchain/common/db/base"
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/trie"
)

type TokenDB struct {
	base *base.Base
	trie *trie.Trie
}

func Open(path string) (*TokenDB, error) {
	var err error
	baseDB, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &TokenDB{base: baseDB}, nil
}

func (t *TokenDB) SetRoot(hash arry.Hash) error {
	tri, err := trie.New(hash, t.base)
	if err != nil {
		return err
	}
	t.trie = tri
	return nil
}

func (t *TokenDB) Commit() (arry.Hash, error) {
	return t.trie.Commit()
}

func (t *TokenDB) Root() arry.Hash {
	return t.trie.Hash()
}

func (t *TokenDB) Close() error {
	return t.base.Close()
}

func (t *TokenDB) Token(address arry.Address) *types.TokenRecord {
	bytes := t.trie.Get(address.Bytes())
	token, err := types.DecodeToken(bytes)
	if err != nil {
		return nil
	}
	return token
}

func (t *TokenDB) SetToken(token *types.TokenRecord) {
	t.trie.Update(token.Address.Bytes(), token.Bytes())
}

func (t *TokenDB) Contract(address arry.Address) *status.Contract {
	bytes := t.trie.Get(address.Bytes())
	contract, err := status.DecodeContract(bytes)
	if err != nil {
		return nil
	}
	return contract
}

func (t *TokenDB) SetContract(contract *status.Contract) {
	t.trie.Update(contract.Address.Bytes(), contract.Bytes())
}

func (t *TokenDB) SetContractState(msgHash arry.Hash, state *types.ContractStatus) {
	t.trie.Update(msgHash.Bytes(), state.Bytes())
}

func (t *TokenDB) ContractState(msgHash arry.Hash) *types.ContractStatus {
	bytes := t.trie.Get(msgHash.Bytes())
	cs, _ := types.DecodeContractState(bytes)
	return cs
}

const symbolBucket = "s_"

func (t *TokenDB) TokenList() []map[string]string {
	rs := t.base.Foreach(symbolBucket)
	var tokens []map[string]string
	for key, value := range rs {
		tokens = append(tokens, map[string]string{
			key: arry.BytesToAddress(value).String(),
		})
	}
	return nil
}

func (t *TokenDB) SymbolContract(symbol string) (arry.Address, bool) {
	bytes, _ := t.base.GetFromBucket(symbolBucket, []byte(symbol))
	if bytes == nil {
		return arry.Address{}, false
	}
	return arry.BytesToAddress(bytes), true
}

func (t *TokenDB) SetSymbolContract(symbol string, address arry.Address) {
	t.base.PutInBucket(symbolBucket, []byte(symbol), address.Bytes())
}
