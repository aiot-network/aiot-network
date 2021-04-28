package msglist

import (
	chaintypes "github.com/aiot-network/aiotchain/chain/types"
	"github.com/aiot-network/aiotchain/common/db/base"
	"github.com/aiot-network/aiotchain/types"
)

const (
	path   = "msglist"
	bucket = "msglist"
)

type MsgListDB struct {
	base *base.Base
}

func Open(path string) (*MsgListDB, error) {
	var err error
	baseDB, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &MsgListDB{base: baseDB}, nil
}

func (t *MsgListDB) Read() []types.IMessage {
	msgs := t.base.Foreach(bucket)
	rs := make([]types.IMessage, 0)
	for _, bytes := range msgs {
		rlpMsg, _ := chaintypes.DecodeMessage(bytes)
		rs = append(rs, rlpMsg.ToMessage())
	}
	return rs
}

func (t *MsgListDB) Save(msg types.IMessage) {
	key := base.Key(bucket, msg.Hash().Bytes())
	t.base.Put(key, msg.ToRlp().Bytes())
}

func (t *MsgListDB) Delete(tx types.IMessage) {
	key := base.Key(bucket, tx.Hash().Bytes())
	t.base.Delete(key)
}

func (t *MsgListDB) Clear() {
	t.base.Clear(bucket)
}

func (t *MsgListDB) Close() error {
	return t.base.Close()
}
