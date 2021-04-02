package blockchain

import (
	"errors"
	"fmt"
	"github.com/aiot-network/aiot-network/chain/common/kit"
	"github.com/aiot-network/aiot-network/chain/db/chain_db"
	chaintypes "github.com/aiot-network/aiot-network/chain/types"
	"github.com/aiot-network/aiot-network/common/config"
	"github.com/aiot-network/aiot-network/common/dpos"
	"github.com/aiot-network/aiot-network/common/param"
	"github.com/aiot-network/aiot-network/common/status"
	servicesync "github.com/aiot-network/aiot-network/service/sync"
	"github.com/aiot-network/aiot-network/tools/arry"
	log "github.com/aiot-network/aiot-network/tools/log/log15"
	"github.com/aiot-network/aiot-network/types"
	"sync"
)

const chainDB = "chain_db"
const module = "module"

type Chain struct {
	mutex         sync.RWMutex
	status        status.IStatus
	db            IChainDB
	dPos          dpos.IDPos
	actRoot       arry.Hash
	dPosRoot      arry.Hash
	tokenRoot     arry.Hash
	lastHeight    uint64
	confirmed     uint64
	poolDeleteMsg func(message types.IMessage)
}

func NewChain(status status.IStatus, dPos dpos.IDPos) (*Chain, error) {
	var err error
	c := &Chain{status: status, dPos: dPos}
	c.db, err = chain_db.Open(config.Param.Data + "/" + chainDB)
	if err != nil {
		return nil, fmt.Errorf("failed to open chain db, %s", err.Error())
	}
	// Read the status tree root hash
	c.actRoot, _ = c.db.ActRoot()
	c.dPosRoot, _ = c.db.DPosRoot()
	c.tokenRoot, _ = c.db.TokenRoot()
	// Initializes the state root hash
	if err := c.status.InitRoots(c.actRoot, c.dPosRoot, c.tokenRoot); err != nil {
		return nil, fmt.Errorf("failed to init status root, %s", err.Error())
	}

	// Initialize chain height
	if c.lastHeight, err = c.db.LastHeight(); err != nil {
		c.saveGenesisBlock(c.dPos.GenesisBlock())
	}
	c.UpdateConfirmed(c.dPos.Confirmed())

	return c, nil
}

func (c *Chain) LastHeight() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.lastHeight
}

func (c *Chain) NextHeader(time uint64) (types.IHeader, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	preHeader, err := c.GetHeaderHeight(c.lastHeight)
	if err != nil {
		return nil, err
	}
	// Build block header
	header := chaintypes.NewHeader(
		preHeader.GetHash(),
		arry.Hash{},
		c.actRoot,
		c.dPosRoot,
		c.tokenRoot,
		c.lastHeight+1,
		time,
		config.Param.IPrivate.Address(),
	)

	return header, nil
}

func (c *Chain) NextBlock(msgs []types.IMessage, blockTime uint64) (types.IBlock, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	height := c.lastHeight + 1

	cycle := blockTime / uint64(param.CycleInterval)

	allWorks := c.getAllWorks(cycle)
	works := c.getWorks(cycle, config.Param.IPrivate.Address())

	coinbase := kit.CalCoinBase(config.Param.Name, allWorks, works)
	to := chaintypes.NewReceivers()
	to.Add(config.Param.IPrivate.Address(), coinbase+chaintypes.CalculateFee(msgs))
	coinBase := &chaintypes.Message{
		Header: &chaintypes.MsgHeader{
			Type: chaintypes.Transaction,
			From: chaintypes.CoinBase,
			Time: blockTime,
		},
		Body: &chaintypes.TransactionBody{
			TokenAddress: config.Param.TokenParam.MainToken,
			Receivers:    to,
		},
	}
	coinBase.SetHash()
	chainMsgs := msgs
	chainMsgs = append(chainMsgs, coinBase)
	lastHeader, err := c.GetHeaderHeight(c.lastHeight)
	if err != nil {
		return nil, err
	}
	// Build block header
	header := chaintypes.NewHeader(
		lastHeader.GetHash(),
		chaintypes.MsgRoot(chainMsgs),
		c.actRoot,
		c.dPosRoot,
		c.tokenRoot,
		height,
		blockTime,
		config.Param.IPrivate.Address(),
	)
	body := &chaintypes.Body{chainMsgs}
	newBlock := &chaintypes.Block{
		Header: header,
		Body:   body,
	}
	newBlock.SetHash()
	if err := newBlock.Sign(config.Param.IPrivate.PrivateKey()); err != nil {
		return nil, err
	}
	return newBlock, nil
}

func (c *Chain) LastConfirmed() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.confirmed
}

func (c *Chain) SetConfirmed(confirmed uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.confirmed = confirmed
	c.status.SetConfirmed(confirmed)
}

func (c *Chain) LastHeader() (types.IHeader, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.db.GetHeaderHeight(c.lastHeight)
}

func (c *Chain) GetBlockHeight(height uint64) (types.IBlock, error) {
	header, err := c.getHeaderHeight(height)
	if err != nil {
		return nil, err
	}
	txs, err := c.db.GetMessages(header.MsgRoot)
	if err != nil {
		return nil, err
	}
	rlpBody := &chaintypes.RlpBody{txs}
	block := &chaintypes.Block{header, rlpBody.ToBody()}
	return block, nil
}

func (c *Chain) GetBlockHash(hash arry.Hash) (types.IBlock, error) {
	header, err := c.getHeaderHash(hash)
	if err != nil {
		return nil, err
	}
	txs, err := c.db.GetMessages(header.MsgRoot)
	if err != nil {
		return nil, err
	}
	rlpBody := &chaintypes.RlpBody{txs}
	block := &chaintypes.Block{header, rlpBody.ToBody()}
	return block, nil
}

func (c *Chain) GetHeaderHeight(height uint64) (types.IHeader, error) {
	return c.getHeaderHeight(height)
}

func (c *Chain) getHeaderHeight(height uint64) (*chaintypes.Header, error) {
	if height > c.LastHeight() {
		return nil, fmt.Errorf("%d block header is not exist", height)
	}
	return c.db.GetHeaderHeight(height)
}

func (c *Chain) GetHeaderHash(hash arry.Hash) (types.IHeader, error) {
	return c.getHeaderHash(hash)
}

func (c *Chain) getHeaderHash(hash arry.Hash) (*chaintypes.Header, error) {
	return c.db.GetHeaderHash(hash)
}

func (c *Chain) CycleLastHash(cycle uint64) (arry.Hash, error) {
	return c.db.CycleLastHash(cycle)
}

func (c *Chain) GetRlpBlockHeight(height uint64) (types.IRlpBlock, error) {
	header, err := c.db.GetHeaderHeight(height)
	if err != nil {
		return nil, err
	}
	txs, err := c.db.GetMessages(header.MsgRoot)
	if err != nil {
		return nil, err
	}
	rlpBody := &chaintypes.RlpBody{txs}
	rlpHeader := header.ToRlpHeader().(*chaintypes.Header)
	block := &chaintypes.RlpBlock{rlpHeader, rlpBody}
	return block, nil
}

func (c *Chain) GetRlpBlockHash(hash arry.Hash) (types.IRlpBlock, error) {
	header, err := c.db.GetHeaderHash(hash)
	if err != nil {
		return nil, err
	}
	txs, err := c.db.GetMessages(header.MsgRoot)
	if err != nil {
		return nil, err
	}
	rlpBody := &chaintypes.RlpBody{txs}
	rlpHeader := header.ToRlpHeader().(*chaintypes.Header)
	block := &chaintypes.RlpBlock{rlpHeader, rlpBody}
	return block, nil
}

func (c *Chain) GetMessage(hash arry.Hash) (types.IMessage, error) {
	rlpTx, err := c.db.GetMessage(hash)
	if err != nil {
		return nil, err
	}
	return rlpTx.ToMessage(), nil
}

func (c *Chain) GetMessageIndex(hash arry.Hash) (types.IMessageIndex, error) {
	msgIndex, err := c.db.GetMsgIndex(hash)
	if err != nil {
		return nil, err
	}
	if msgIndex.Height > c.LastHeight() {
		return nil, errors.New("not exist")
	}
	return msgIndex, nil
}

func (c *Chain) Insert(block types.IBlock) error {
	if err := c.checkBlock(block); err != nil {
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lastHeight >= block.GetHeight() {
		return errors.New("wrong block height")
	}
	if err := c.status.Change(block.BlockBody().MsgList(), block); err != nil {
		return err
	}
	msgs := block.BlockBody().MsgList()
	for _, msg := range msgs {
		if c.poolDeleteMsg != nil {
			c.poolDeleteMsg(msg)
		} else {
			log.Error("Need to register message pool delete function", "module", module)
		}
	}
	c.saveBlock(block)
	return nil
}

func (c *Chain) saveBlock(block types.IBlock) {
	bk := block.(*chaintypes.Block)
	rlpBlock := bk.ToRlpBlock().(*chaintypes.RlpBlock)
	c.db.SaveHeader(bk.Header)
	c.db.SaveMessages(block.GetMsgRoot(), rlpBlock.RlpBody.MsgList())
	c.db.SaveMsgIndex(bk.GetMsgIndexs())
	c.db.SaveHeightHash(block.GetHeight(), block.GetHash())
	c.db.SaveConfirmedHeight(block.GetHeight(), c.confirmed)
	c.db.SaveCycleLastHash(block.GetCycle(), block.GetHash())
	c.actRoot, c.tokenRoot, c.dPosRoot, _ = c.status.Commit()
	c.db.SaveActRoot(c.actRoot)
	c.db.SaveDPosRoot(c.dPosRoot)
	c.db.SaveTokenRoot(c.tokenRoot)

	c.lastHeight = block.GetHeight()
	c.db.SaveLastHeight(c.lastHeight)
	/*log.Info("Save block", "module", "module",
	"height", block.GetHeight(),
	"hash", block.GetHash().String(),
	"actroot", block.GetActRoot().String(),
	"tokenroot", block.GetTokenRoot().String(),
	"dposroot", block.GetDPosRoot().String(),
	"signer", block.GetSigner().String(),
	"msgcount", len(block.BlockBody().MsgList()),
	"time", block.GetTime(),
	"cycle", block.GetCycle())*/
}

func (c *Chain) saveGenesisBlock(block types.IBlock) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.verifyGenesis(block); err != nil {
		return err
	}

	c.status.Change(block.BlockBody().MsgList(), block)
	bk := block.(*chaintypes.Block)
	rlpBlock := bk.ToRlpBlock().(*chaintypes.RlpBlock)
	c.db.SaveHeader(bk.Header)
	c.db.SaveMessages(block.GetMsgRoot(), rlpBlock.RlpBody.MsgList())
	c.db.SaveMsgIndex(bk.GetMsgIndexs())
	c.db.SaveHeightHash(block.GetHeight(), block.GetHash())
	c.lastHeight = block.GetHeight()
	c.db.SaveConfirmedHeight(block.GetHeight(), c.confirmed)
	c.status.SetConfirmed(0)
	c.actRoot, c.tokenRoot, c.dPosRoot, _ = c.status.Commit()
	c.db.SaveActRoot(c.actRoot)
	c.db.SaveDPosRoot(c.dPosRoot)
	c.db.SaveTokenRoot(c.tokenRoot)
	c.db.SaveLastHeight(c.lastHeight)

	log.Info("Save block", "module", "module",
		"height", block.GetHeight(),
		"hash", block.GetHash().String(),
		"actroot", block.GetActRoot().String(),
		"tokenroot", block.GetTokenRoot().String(),
		"dposroot", block.GetDPosRoot().String(),
		"signer", block.GetSigner().String(),
		"msgcount", len(block.BlockBody().MsgList()),
		"time", block.GetTime(),
		"cycle", block.GetCycle())
	return nil
}

func (c *Chain) verifyGenesis(block types.IBlock) error {
	var sumCoins uint64
	for _, tx := range block.BlockBody().MsgList() {
		sumCoins += tx.MsgBody().MsgAmount()
	}
	if sumCoins != config.Param.PreCirculation {
		return fmt.Errorf("wrong genesis coins")
	}
	return nil
}

func (c *Chain) checkBlock(block types.IBlock) error {
	lastHeight := c.LastHeight()

	if block.GetHeight() == lastHeight {
		lastHeader, err := c.GetHeaderHeight(lastHeight)
		if err == nil && lastHeader.GetHash().IsEqual(block.GetHash()) {
			return servicesync.Err_RepeatBlock
		}
	}

	if lastHeight != block.GetHeight()-1 {
		return fmt.Errorf("last height is %d, the current block height is %d", lastHeight, block.GetHeight())
	}

	if !block.CheckMsgRoot() {
		log.Warn("the message root hash verification failed", "module", module,
			"height", block.GetHeight(), "msgroot", block.GetMsgRoot().String())
		return errors.New("the message root hash verification failed")
	}
	if !block.GetActRoot().IsEqual(c.actRoot) {
		log.Warn("the account status root hash verification failed", "module", module,
			"height", block.GetHeight(), "actroot", block.GetActRoot().String())
		return errors.New("the account status root hash verification failed")
	}
	if !block.GetDPosRoot().IsEqual(c.dPosRoot) {
		log.Warn("the dpos status root hash verification failed", "module", module,
			"height", block.GetHeight(), "dposroot", block.GetDPosRoot().String())
		return errors.New("wrong dpos root")
	}
	if !block.GetTokenRoot().IsEqual(c.tokenRoot) {
		log.Warn("the token status root hash verification failed", "module", module,
			"height", block.GetHeight(), "tokenroot", block.GetTokenRoot().String())
		return errors.New("wrong token root")
	}
	preHeader, err := c.GetHeaderHash(block.GetPreHash())
	if err != nil {
		return fmt.Errorf("no previous block %s found", block.GetPreHash().String())
	}

	if err := c.dPos.CheckHeader(block.BlockHeader(), preHeader, c); err != nil {
		return err
	}
	if err := c.dPos.CheckSeal(block.BlockHeader(), preHeader, c); err != nil {
		return err
	}
	if err := c.checkMsgs(block.BlockBody().MsgList(), block.GetHeight()); err != nil {
		return err
	}
	return nil
}

func (c *Chain) checkMsgs(msgs []types.IMessage, height uint64) error {
	address := make(map[string]int)
	for i, msg := range msgs {
		if msg.IsCoinBase() {
			if err := c.checkCoinBase(msg, chaintypes.CalculateFee(msgs), height); err != nil {
				return err
			}
		} else {
			if err := c.checkMsg(msg); err != nil {
				return err
			}
		}
		from := msg.From().String()
		if lastIndex, ok := address[from]; !ok {
			address[from] = i
		} else {
			log.Warn("Repeat address block", "module", module,
				"preMsg", msgs[lastIndex],
				"curMsg", msg)
			return errors.New("one address in a block can only send one transaction")
		}
	}
	return nil
}

func (c *Chain) checkCoinBase(coinBase types.IMessage, fee, height uint64) error {
	msg, ok := coinBase.(*chaintypes.Message)
	if !ok {
		return errors.New("wrong message type")
	}
	cycle := msg.Time() / uint64(param.CycleInterval)
	rei := coinBase.MsgBody().MsgTo().ReceiverList()

	allWorks := c.getAllWorks(cycle)
	works := c.getWorks(cycle, rei[0].Address)
	coinbase := kit.CalCoinBase(config.Param.Name, allWorks, works)
	if err := msg.CheckCoinBase(fee, coinbase); err != nil {
		return err
	}
	return nil
}

func (c *Chain) checkMsg(msg types.IMessage) error {
	msg, ok := msg.(*chaintypes.Message)
	if !ok {
		return errors.New("wrong message type")
	}

	if err := msg.Check(); err != nil {
		return err
	}

	if err := c.status.CheckMsg(msg, true); err != nil {
		return err
	}
	return nil
}

func (c *Chain) Confirmed() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.confirmed
}

func (c *Chain) Roll() error {
	var curHeight uint64
	confirmed := c.Confirmed()
	if confirmed != 0 {
		curHeight = confirmed
	}
	return c.RollbackTo(curHeight)
}

func (c *Chain) RollbackTo(height uint64) error {
	confirmedHeight := c.confirmed
	if height > confirmedHeight && height != 0 {
		err := fmt.Sprintf("the height of the roolback must be less than or equal to %d and greater than %d", confirmedHeight, 0)
		log.Error("Roll back to block height", "height", height, "error", err)
		return errors.New(err)
	}

	var curBlockHeight, nextBlockHeight uint64
	curActRoot := arry.Hash{}
	curTokenRoot := arry.Hash{}
	curDPosRoot := arry.Hash{}

	nextBlockHeight = height + 1
	curBlockHeight = height

	// set new confirmed height and header
	hisConfirmedHeight, err := c.db.GetConfirmedHeight(curBlockHeight)
	if err != nil {
		log.Error("Fall back to block height", "height", height, "error", "can not find history confirmed height")
		return fmt.Errorf("fall back to block height %d failed! Can not find history confirmed height", height)
	}
	c.dPos.SetConfirmed(hisConfirmedHeight)

	log.Warn("Fall back to block height", "height", height)
	header, err := c.GetHeaderHeight(nextBlockHeight)
	if err != nil {
		log.Error("Fall back to block height", "height", height, "error", "can not find block")
		return fmt.Errorf("fall back to block height %d failed! Can not find block %d", height, nextBlockHeight)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.confirmed = hisConfirmedHeight
	c.status.SetConfirmed(hisConfirmedHeight)

	// fall back to pre state root
	curActRoot = header.GetActRoot()
	curTokenRoot = header.GetTokenRoot()
	curDPosRoot = header.GetDPosRoot()
	err = c.status.InitRoots(curActRoot, curDPosRoot, curTokenRoot)
	if err != nil {
		log.Error("Fall back to block height", "height", height, "error", "init state trie failed")
		return fmt.Errorf("fall back to block height %d failed! nit state trie failed", height)
	}
	c.actRoot = curActRoot
	c.tokenRoot = curTokenRoot
	c.dPosRoot = curDPosRoot
	c.db.SaveActRoot(c.actRoot)
	c.db.SaveTokenRoot(c.tokenRoot)
	c.db.SaveDPosRoot(c.dPosRoot)

	c.lastHeight = curBlockHeight
	c.db.SaveLastHeight(curBlockHeight)
	return nil
}

func (c *Chain) UpdateConfirmed(height uint64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.confirmed = height
	c.status.SetConfirmed(height)
}

func (c *Chain) Vote(address arry.Address) uint64 {
	var vote uint64
	act := c.status.Account(address)
	vote += act.GetBalance(config.Param.MainToken)
	return vote
}

func (c *Chain) RegisterMsgPoolDeleteFunc(fun func(message types.IMessage)) {
	c.poolDeleteMsg = fun
}

func (c *Chain) getAllWorks(cycle uint64) uint64 {
	var allWorks uint64
	supers := c.status.CycleSupers(cycle)
	if supers != nil {
		list := supers.List()
		for _, s := range list {
			act := c.status.Account(s.GetSinger())
			work := act.GetWorks()
			actCycle, actWorks := work.GetCycle(), work.GetWorkLoad()
			if actCycle == cycle-1 {
				allWorks += actWorks
			}
		}
	}
	return allWorks
}

func (c *Chain) getWorks(cycle uint64, address arry.Address) uint64 {
	var works uint64
	act := c.status.Account(address)
	work := act.GetWorks()
	actCycle, actWorks := work.GetCycle(), work.GetWorkLoad()
	if actCycle == cycle-1 {
		works = actWorks
	}
	return works
}
