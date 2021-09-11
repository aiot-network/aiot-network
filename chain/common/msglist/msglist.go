package msglist

import (
	"fmt"
	"github.com/aiot-network/aiotchain/chain/db/msglist"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/validator"
	"github.com/aiot-network/aiotchain/types"
	"sync"
)

const msgList_db = "msg_List_db"
const maxPoolTx = 100000

type lastHeightFunc func() uint64

type MsgManagement struct {
	cache     *Cache
	ready     *Sorted
	validator validator.IValidator
	actStatus types.IActStatus
	mutex     sync.RWMutex
	msgDB     ITxListDB
	lastHeightFunc
}

func NewMsgManagement(validator validator.IValidator, actStatus types.IActStatus, lastHeightFunc lastHeightFunc) (*MsgManagement, error) {
	msgDB, err := msglist.Open(config.Param.Data + "/" + msgList_db)
	if err != nil {
		return nil, err
	}
	return &MsgManagement{
		cache:          NewCache(msgDB),
		ready:          NewSorted(msgDB),
		validator:      validator,
		actStatus:      actStatus,
		msgDB:          msgDB,
		lastHeightFunc: lastHeightFunc,
	}, nil
}

func (t *MsgManagement) Read() error {
	msgs := t.msgDB.Read()
	if msgs != nil {
		for _, msg := range msgs {
			if err := t.Put(msg); err != nil {
				t.msgDB.Delete(msg)
			}
		}
	}
	return nil
}

func (t *MsgManagement) Close() error {
	return t.msgDB.Close()
}

func (t *MsgManagement) Count() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.cache.Len() + t.ready.Len()
}

func (t *MsgManagement) Put(msg types.IMessage) error {
	if t.Exist(msg) {
		return fmt.Errorf("the message %s already exists", msg.Hash().String())
	}
	if err := t.validator.CheckMsg(msg, false, t.lastHeightFunc()); err != nil {
		return err
	}

	if t.cache.Len() >= maxPoolTx {
		t.DeleteEnd(msg)
	}
	return t.put(msg)
}

func (t *MsgManagement) put(msg types.IMessage) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	from := msg.From().String()
	nonce := t.actStatus.Nonce(msg.From())
	if nonce == msg.Nonce()-1 {
		oldTx := t.ready.GetByAddress(from)
		if oldTx != nil {
			if oldTx.Nonce() == msg.Nonce() && oldTx.Fee() < msg.Fee() {
				t.ready.Remove(oldTx)
			} else if oldTx.Nonce() < msg.Nonce() {
				t.ready.Remove(oldTx)
			} else if oldTx.Nonce() == msg.Nonce() {
				return fmt.Errorf("the same nonce %d message already exists, so if you want to replace the nonce message, add a fee", msg.Nonce())
			} else {
				return fmt.Errorf("the nonce value %d is repeated, increase the nonce value", msg.Nonce())
			}
		}
		t.ready.Put(msg)
	} else if nonce >= msg.Nonce() {
		return fmt.Errorf("the nonce value %d is repeated, increase the nonce value", msg.Nonce())
	} else {
		t.cache.Put(msg)
	}
	t.msgDB.Save(msg)
	return nil
}

func (t *MsgManagement) Delete(msg types.IMessage) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.remove(msg)
	t.update()
}

func (t *MsgManagement) DeleteEnd(newTx types.IMessage) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.ready.PopMin(newTx.Fee())
}

func (t *MsgManagement) NeedPackaged(maxSize uint32) []types.IMessage {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.ready.NeedPackaged(maxSize)
}

func (t *MsgManagement) StagnantMsgs() []types.IMessage {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.ready.StagnantMsgs()
}

func (t *MsgManagement) GetAll() ([]types.IMessage, []types.IMessage) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	readyTxs := t.ready.All()
	cacheTxs := t.cache.All()
	return readyTxs, cacheTxs
}

func (t *MsgManagement) Get(msgHash string) (types.IMessage, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	msg, ok := t.ready.Get(msgHash)
	if !ok {
		msg, ok := t.cache.Get(msgHash)
		return msg, ok
	}
	return msg, ok
}

func (t *MsgManagement) Exist(msg types.IMessage) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if !t.ready.Exist(msg.Hash().String()) {
		return t.cache.Exist(msg.From().String())
	}
	return true
}

func (t *MsgManagement) Update() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.update()
}

func (t *MsgManagement) update() {
	t.ready.RemoveExecuted(t.validator, t.lastHeightFunc())
	for _, msg := range t.cache.msgs {
		nonce := t.actStatus.Nonce(msg.From())
		if nonce < msg.Nonce()-1 {
			continue
		}
		if nonce == msg.Nonce()-1 {
			t.ready.Put(msg)
		}
		t.cache.Remove(msg)
	}
}

func (t *MsgManagement) DeleteExpired(timeThreshold int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.ready.RemoveExpiredTx(timeThreshold)

	for _, msg := range t.cache.msgs {
		if msg.Time() <= uint64(timeThreshold) {
			t.cache.Remove(msg)
		}
	}
}

func (t *MsgManagement) remove(msg types.IMessage) {
	t.cache.Remove(msg)
	t.ready.Remove(msg)
}
