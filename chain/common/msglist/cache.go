package msglist

import (
	"fmt"
	"github.com/aiot-network/aiot-network/types"
	"strconv"
)

type Cache struct {
	msgs     map[string]types.IMessage
	nonceTxs map[string]string
	db       ITxListDB
}

func NewCache(db ITxListDB) *Cache {
	return &Cache{
		msgs:     make(map[string]types.IMessage),
		nonceTxs: make(map[string]string),
		db:       db,
	}
}

func (c *Cache) Put(msg types.IMessage) error {
	if c.Exist(msg.Hash().String()) {
		return fmt.Errorf("transation hash %s exsit", msg.Hash())
	}
	nonceKey := nonceKey(msg)
	if oldTxHash := c.getHash(nonceKey); oldTxHash != "" {
		oldTx := c.msgs[oldTxHash]
		if oldTx.Fee() > msg.Fee() {
			return fmt.Errorf("transation nonce %d exist, the fees must biger than before %d", msg.Nonce(), oldTx.Fee())
		}
		c.Remove(oldTx)
	}
	c.msgs[msg.Hash().String()] = msg
	c.nonceTxs[nonceKey] = msg.Hash().String()
	c.db.Save(msg)
	return nil
}

func (c *Cache) Get(msgHash string) (types.IMessage, bool) {
	msg, ok := c.msgs[msgHash]
	return msg, ok
}

func (c *Cache) Remove(msg types.IMessage) {
	delete(c.msgs, msg.Hash().String())
	delete(c.nonceTxs, nonceKey(msg))
	c.db.Delete(msg)
}

func (c *Cache) Exist(msgHash string) bool {
	_, ok := c.msgs[msgHash]
	return ok
}

func (c *Cache) Len() int {
	return len(c.msgs)
}

func (c *Cache) All() []types.IMessage {
	var all = make([]types.IMessage, 0)
	for _, msg := range c.msgs {
		all = append(all, msg)
	}
	return all
}

func (c *Cache) getHash(nonceKey string) string {
	return c.nonceTxs[nonceKey]
}

func nonceKey(msg types.IMessage) string {
	return msg.From().String() + "_" + strconv.FormatUint(msg.Nonce(), 10)
}
