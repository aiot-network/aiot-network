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

func (t *TokenDB) TokenList() []map[string]string {
	/*iter := t.trie.PrefixIterator([]byte{})
	var tokens []*types.TokenRecord
	for iter.Next(true) {
		if iter.Leaf() {
			key := iter.LeafKey()
			tokens = append(tokens, &types.Token{
				Symbol:   string(database.LeafKeyToKey(symbolBucket, key)),
				Contract: hasharry.BytesToAddress(iter.LeafBlob()).String(),
			})
		}
	}
	return tokens*/
	return nil
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

func (t *TokenDB) SymbolContract(symbol string) (arry.Address, bool) {
	bytes := t.trie.Get([]byte(symbol))
	if bytes == nil {
		return arry.Address{}, false
	}
	return arry.BytesToAddress(bytes), true
}

func (t *TokenDB) SetSymbolContract(symbol string, address arry.Address) {
	t.trie.Update([]byte(symbol), address.Bytes())
}
