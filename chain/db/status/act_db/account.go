package act_db

import (
	"github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/common/db/base"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/trie"
	types2 "github.com/aiot-network/aiot-network/types"
)

type ActDB struct {
	base *base.Base
	trie *trie.Trie
}

func Open(path string) (*ActDB, error) {
	var err error
	baseDB, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &ActDB{base: baseDB}, nil
}

func (a *ActDB) SetRoot(hash arry.Hash) error {
	t, err := trie.New(hash, a.base)
	if err != nil {
		return err
	}
	a.trie = t
	return nil
}

func (a *ActDB) Root() arry.Hash {
	return a.trie.Hash()
}

func (a *ActDB) Commit() (arry.Hash, error) {
	return a.trie.Commit()
}

func (a *ActDB) Account(address arry.Address) types2.IAccount {
	bytes := a.trie.Get(address.Bytes())
	if account, err := types.DecodeAccount(bytes); err != nil {
		return types.NewAccount()
	} else {
		return account
	}
}

func (a *ActDB) SetAccount(account types2.IAccount) {
	a.trie.Update(account.GetAddress().Bytes(), account.Bytes())
}

func (a *ActDB) Nonce(address arry.Address) uint64 {
	bytes := a.trie.Get(address.Bytes())
	account, err := types.DecodeAccount(bytes)
	if err != nil {
		return 0
	}
	return account.Nonce
}

func (a *ActDB) Close() error {
	return a.base.Close()
}
