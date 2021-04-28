package msglist

import (
	"container/heap"
	"github.com/aiot-network/aiotchain/common/validator"
	"github.com/aiot-network/aiotchain/tools/utils"
	"github.com/aiot-network/aiotchain/types"
)

const maxStagnantTime uint64 = 60

type Sorted struct {
	msgs  map[string]types.IMessage
	cache map[string]types.IMessage
	index *msgInfos
	db    ITxListDB
}

func NewSorted(db ITxListDB) *Sorted {
	return &Sorted{
		msgs:  make(map[string]types.IMessage),
		cache: make(map[string]types.IMessage),
		index: new(msgInfos),
		db:    db,
	}
}

func (t *Sorted) Put(msg types.IMessage) {
	t.msgs[msg.From().String()] = msg
	t.cache[msg.Hash().String()] = msg
	heap.Push(t.index, &msgInfo{
		address: msg.From().String(),
		msgHash: msg.Hash().String(),
		fees:    msg.Fee(),
		nonce:   msg.Nonce(),
		time:    msg.Time(),
	})
	t.db.Save(msg)
}

func (t *Sorted) Get(msgHash string) (types.IMessage, bool) {
	msg, ok := t.cache[msgHash]
	return msg, ok
}

func (t *Sorted) All() []types.IMessage {
	var all []types.IMessage
	for _, msg := range t.cache {
		all = append(all, msg)
	}
	return all
}

func (t *Sorted) NeedPackaged(count int) []types.IMessage {
	msgs := make([]types.IMessage, 0)
	rIndex := t.index.CopySelf()

	for rIndex.Len() > 0 && count > 0 {
		ti := heap.Pop(rIndex).(*msgInfo)
		msg := t.msgs[ti.address]
		msgs = append(msgs, msg)
		count--
	}
	return msgs
}

func (t *Sorted) StagnantMsgs() []types.IMessage {
	msgs := make([]types.IMessage, 0)
	rIndex := t.index.CopySelf()

	for rIndex.Len() > 0 {
		ti := heap.Pop(rIndex).(*msgInfo)
		msg := t.msgs[ti.address]
		if msg != nil && msg.Time()+maxStagnantTime > uint64(utils.NowUnix()) {
			msgs = append(msgs, msg)
		}

	}
	return msgs
}

func (t *Sorted) GetByAddress(addr string) types.IMessage {
	return t.msgs[addr]
}

// If the message pool is full, delete the message with a small fee
func (t *Sorted) PopMin(fees uint64) types.IMessage {
	if t.Len() > 0 {
		if (*t.index)[0].fees <= fees {
			ti := heap.Remove(t.index, 0).(*msgInfo)
			msg := t.msgs[ti.address]
			delete(t.msgs, ti.address)
			delete(t.cache, ti.msgHash)
			t.db.Delete(msg)
			return msg
		}
	}
	return nil
}

func (t *Sorted) Len() int { return len(t.msgs) }

func (t *Sorted) Exist(msgHash string) bool {
	_, ok := t.cache[msgHash]
	return ok
}

func (t *Sorted) Remove(msg types.IMessage) {
	for i, ti := range *(t.index) {
		if ti.msgHash == msg.Hash().String() {
			heap.Remove(t.index, i)
			delete(t.msgs, msg.From().String())
			delete(t.cache, msg.Hash().String())
			t.db.Delete(msg)
			return
		}
	}
}

// Delete already packed messages
func (t *Sorted) RemoveExecuted(v validator.IValidator) {
	for _, msg := range t.cache {
		if err := v.CheckMsg(msg, false); err != nil {
			t.Remove(msg)
		}
	}
}

// Delete expired messages
func (t *Sorted) RemoveExpiredTx(timeThreshold int64) {
	for _, msg := range t.cache {
		if msg.Time() <= uint64(timeThreshold) {
			t.Remove(msg)
		}
	}
}

type msgInfos []*msgInfo

type msgInfo struct {
	address string
	msgHash string
	fees    uint64
	nonce   uint64
	time    uint64
}

func (t msgInfos) Len() int           { return len(t) }
func (t msgInfos) Less(i, j int) bool { return t[i].fees > t[j].fees }
func (t msgInfos) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func (t *msgInfos) Push(x interface{}) {
	*t = append(*t, x.(*msgInfo))
}

func (t *msgInfos) Pop() interface{} {
	old := *t
	n := len(old)
	x := old[n-1]
	*t = old[0 : n-1]
	return x
}

func (t *msgInfos) CopySelf() *msgInfos {
	reReelList := new(msgInfos)
	for _, nonce := range *t {
		*reReelList = append(*reReelList, nonce)
	}
	return reReelList
}

func (t *msgInfos) FindIndex(addr string, nonce uint64) int {
	for index, ti := range *t {
		if ti.address == addr && ti.nonce <= nonce {
			return index
		}
	}
	return -1
}
