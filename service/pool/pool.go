package pool

import (
	"fmt"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/horn"
	"github.com/aiot-network/aiotchain/common/msglist"
	hasharry "github.com/aiot-network/aiotchain/tools/arry"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/tools/utils"
	"github.com/aiot-network/aiotchain/types"
	"time"
)

const module = "pool"

type Pool struct {
	msgMgt      msglist.IMsgList
	horn        *horn.Horn
	broadcastCh chan types.IMessage
	deleteMsg   chan types.IMessage
	close       chan bool
}

func NewPool(horn *horn.Horn, msgMgt msglist.IMsgList) *Pool {
	pool := &Pool{
		msgMgt:      msgMgt,
		horn:        horn,
		broadcastCh: make(chan types.IMessage, 100),
		deleteMsg:   make(chan types.IMessage, 10000),
		close:       make(chan bool),
	}
	return pool
}

func (p *Pool) Name() string {
	return module
}

func (p *Pool) Start() error {
	if err := p.msgMgt.Read(); err != nil {
		log.Error("The message pool failed to read the message", "module", module, "error", err)
		return err
	}
	go p.monitorExpired()
	go p.startChan()
	log.Info("Pool started successfully", "module", module)
	return nil
}

func (p *Pool) Stop() error {
	if err := p.msgMgt.Close(); err != nil {
		p.close <- true
		return err
	}
	p.close <- true
	log.Info("Message pool was stopped", "module", module)
	return nil
}

func (p *Pool) Info() map[string]interface{} {
	return map[string]interface{}{
		"messages": p.msgMgt.Count(),
	}
}

// Verify adding messages to the message pool
func (p *Pool) Put(msg types.IMessage, isPeer bool) error {
	if err := p.msgMgt.Put(msg); err != nil {
		return utils.Error(fmt.Sprintf("add message failed, %s", err.Error()), module)
	}
	log.Info("Received the message", "module", module, "hash", msg.Hash().String())
	if !isPeer {
		p.broadcastCh <- msg
	}
	return nil
}

func (p *Pool) NeedPackaged(count int) []types.IMessage {
	msgs := p.msgMgt.NeedPackaged(count)
	return msgs
}

func (p *Pool) startChan() {
	for {
		select {
		case _ = <-p.close:
			return
		case msg := <-p.broadcastCh:
			p.horn.BroadcastMsg(msg)
		case msg := <-p.deleteMsg:
			p.msgMgt.Delete(msg)
		}
	}
}

func (p *Pool) GetMessage(hash hasharry.Hash) (types.IMessage, bool) {
	return p.msgMgt.Get(hash.String())
}

func (p *Pool) ReceiveMsgFromPeer(msg types.IMessage) error {
	return p.Put(msg, true)
}

func (p *Pool) monitorExpired() {
	t := time.NewTicker(time.Second * config.Param.MonitorMsgInterval)
	defer t.Stop()

	for {
		select {
		case _ = <-t.C:
			p.removeExpired()
			p.dealStagnant()
		}
	}
}

func (p *Pool) removeExpired() {
	threshold := utils.NowUnix() - config.Param.MsgExpiredTime
	p.msgMgt.DeleteExpired(threshold)
}

func (p *Pool) dealStagnant() {
	msgs := p.msgMgt.StagnantMsgs()
	if msgs != nil && len(msgs) > 0 {
		for _, msg := range msgs {
			p.broadcastCh <- msg
		}
	}
}

func (p *Pool) Delete(msg types.IMessage) {
	p.deleteMsg <- msg
}

// Get all transactions in the trading pool
func (p *Pool) All() ([]types.IMessage, []types.IMessage) {
	prepareTxs, futureTxs := p.msgMgt.GetAll()
	return prepareTxs, futureTxs
}
