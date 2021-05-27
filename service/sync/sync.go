package sync

import (
	"errors"
	"github.com/aiot-network/aiotchain/common/blockchain"
	"github.com/aiot-network/aiotchain/common/dpos"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/service/peers"
	"github.com/aiot-network/aiotchain/service/request"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/types"
	"sync"
	"time"
)

const module = "sync"

var (
	Err_RepeatBlock = errors.New("repeat the block")
)

type Sync struct {
	chain   blockchain.IChain
	request request.IRequestHandler
	peers   *peers.Peers
	curPeer *types.Peer
	dPos    dpos.IDPosStatus
	stop    chan bool
	stopped chan bool
	mutex sync.RWMutex
}

func NewSync(peers *peers.Peers, dPos dpos.IDPosStatus, request request.IRequestHandler, chain blockchain.IChain) *Sync {
	s := &Sync{
		chain:   chain,
		peers:   peers,
		dPos:    dPos,
		request: request,
		stop:    make(chan bool),
		stopped: make(chan bool),
	}
	return s
}

func (s *Sync) Name() string {
	return module
}

func (s *Sync) Start() error {
	go s.syncBlocks()
	log.Info("Sync started successfully", "module", module)
	return nil
}

func (s *Sync) Stop() error {
	close(s.stop)
	<-s.stopped
	log.Info("Stop sync block", "module", module)
	return nil
}

func (s *Sync) Info() map[string]interface{} {
	return map[string]interface{}{
		"height":    s.chain.LastHeight(),
		"confirmed": s.chain.LastConfirmed(),
	}
}

func (s *Sync)getCurPeer()*types.Peer{
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.curPeer
}

func (s *Sync)setCurPeer(peer *types.Peer){
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.curPeer = peer
}

// Start sync block
func (s *Sync) syncBlocks() {
	for {
		select {
		case _, _ = <-s.stop:
			s.stopped <- true
			return
		default:
			s.createSyncStream()
			s.syncFromConn()

		}
		time.Sleep(time.Millisecond * 1000)
	}
}

// Create a network channel of the synchronization block, and randomly
// select a new peer node for synchronization every 1s.
func (s *Sync) createSyncStream() {
	for {
		select {
		case _, _ = <-s.stop:
			return
		default:
			s.findSyncPeer()
			return
		}
	}
}

// Replace the new peer node
func (s *Sync) findSyncPeer() {
	t := time.NewTicker(time.Microsecond)
	defer t.Stop()

	for {
		select {
		case _, _ = <-s.stop:
			return
		case _ = <-t.C:
			peer := s.peers.RandomPeer()
			if peer != nil{
				s.setCurPeer(peer)
				return
			}
		}
	}
}

// Synchronize blocks from the stream and verify storage
func (s *Sync) syncFromConn() error {
	for {
		select {
		case _, _ = <-s.stop:
			return nil
		default:
			curPeer := s.getCurPeer()
			if curPeer == nil {
				return errors.New("no current peer")
			}
			localHeight := s.chain.LastHeight()

			// Get the block of the remote node from the next block height，
			// If the error is that the peer has stopped, delete the peer.
			// If the storage fails locally, the remote block verification
			// is performed, the verification proves that the local block
			// is wrong, and the local chain is rolled back to the valid block.

			blocks, err := s.request.GetBlocks(curPeer.Conn, localHeight+1, curPeer.Speed)
			if err != nil {
				if err == request.Err_PeerClosed {
					s.reducePeerSpeed(curPeer)
				}
				return err
			}
			if err := s.insert(blocks, curPeer); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Sync) reducePeerSpeed(peer *types.Peer) {
	if peer.Speed == 1 {
		s.peers.RemovePeer(peer.Address.ID.String())
		return
	}
	speed := peer.Speed - 1
	s.peers.SetSpeed(peer.Address.ID.String(), speed)
}

func (s *Sync) insert(blocks []types.IBlock, peer *types.Peer) error {
	for _, block := range blocks {
		select {
		case _, _ = <-s.stop:
			return nil
		default:
			if err := s.chain.Insert(block); err != nil {
				log.Warn("Insert chain failed!", "module", module,
					"error", err, "height",
					block.GetHeight(),
					"signer", block.GetSigner())
				if s.NeedValidation(err) {
					if roll, peerId := s.isRollBack(block.BlockHeader());roll {
						log.Info("Start roll back")
						s.fallBack()
						if peerInfo := s.peers.Peer(peerId); peerInfo == nil {
							return err
						} else {
							s.setCurPeer(peerInfo)
							return nil
						}
					}
				}
				return err
			}
		}
	}
	if len(blocks) > 0 {
		log.Info("Sync blocks complete", "module", module, "start", blocks[0].GetHeight(),
			"end", blocks[len(blocks)-1].GetHeight(), "peer", peer.Address.String(), "speed", peer.Speed)
	}

	return nil
}

// Remotely verify the block, if the block height is less than
// the effective block height, then discard the block. If the
// block occupies the majority of the currently started super
// nodes, it means that the block is more likely to be correct,
// and the block verification is successful.
func (s *Sync) isRollBack(header types.IHeader) (bool, string) {
	if header.GetHeight() <= s.chain.LastConfirmed() {
		return false, ""
	}
	localHeight := s.chain.LastHeight()
	return s.isRoll(header, localHeight)

}

func (s *Sync) validation(header types.IHeader, localEqual bool) bool {
	count := 0
	supers, err := s.dPos.CycleSupers(header.GetCycle())
	if err != nil {
		return false
	}
	for _, candidate := range supers.List() {
		if candidate.GetPeerId() != s.peers.Local().Address.ID.String() {
			peer := s.peers.Peer(candidate.GetPeerId())
			if peer != nil {
				rs, err := s.request.IsEqual(peer.Conn, header)
				if err == nil && rs {
					count++
				}
			}
		} else if localEqual {
			count++
		}
	}
	if count > param.SuperSize/2 {
		return true
	}
	return false
}

func (s *Sync) isRoll(header types.IHeader, localHeight uint64) (bool, string) {
	log.Info("Is need to roll back")
	supers, err := s.dPos.CycleSupers(header.GetCycle())
	if err != nil {
		return false, ""
	}
	var maxHeight uint64
	var maxHeightPeer string
	for _, candidate := range supers.List() {
		if candidate.GetPeerId() != s.peers.Local().Address.ID.String() {
			peer := s.peers.Peer(candidate.GetPeerId())
			if peer != nil {
				height, err := s.request.LastHeight(peer.Conn)
				if err == nil {
					log.Info("Get peer last height", "peer", peer.Address.String(), "height", height)
					if height > maxHeight{
						maxHeight = height
						maxHeightPeer = candidate.GetPeerId()
					}
				}
			}
		}
	}
	if maxHeight > localHeight{
		log.Info("Find the highest node", "remote height", maxHeight, "local height", localHeight, "remote peer", maxHeightPeer)
		peer := s.peers.Peer(maxHeightPeer)
		if peer != nil{
			ok, err := s.request.IsEqual(peer.Conn, header)
			if ok {
				return true, maxHeightPeer
			} else if err != nil {
				log.Error("Failed to validation block hash!", "hash", header.GetHash(), "err", err.Error(), "remote peer", maxHeightPeer)
				return false, ""
			} else {
				return false, ""
			}
		}
	}
	return false, ""
}

// Block chain rolls back to a valid block
func (s *Sync) fallBack() {
	s.chain.Roll()
}

// Process blocks received from other super nodes.If the height
// of the block is greater than the local height, the storage is
// directly verified. If the height is less than the local height,
// the remote verification is performed, and the verification is
// passed back to the local block.
func (s *Sync) ReceivedBlockFromPeer(block types.IBlock) error {
	localHeight := s.chain.LastHeight()
	if localHeight == block.GetHeight()-1 {
		if err := s.chain.Insert(block); err != nil {
			log.Warn("Failed to insert received block", "err", err, "height", block.GetHeight(), "singer", block.GetSigner().String())
			return err
		}
		log.Info("Received block insert success", "module", module, "height", block.GetHeight(), "signer", block.GetSigner())
	} /*else if block.GetHeight() <= localHeight {
		if localHeader, err := s.chain.GetBlockHeight(block.GetHeight()); err == nil {
			if !localHeader.GetHash().IsEqual(block.GetHash()) {
				if roll, peerId := s.isRollBack(block.BlockHeader());roll {
					log.Info("Start roll back")
					s.fallBack()
					if peerInfo := s.peers.Peer(peerId); peerInfo != nil {
						s.setCurPeer(peerInfo)
						return nil
					}
				}
			}
		} else {
			if err := s.chain.Insert(block); err != nil {
				log.Warn("Failed to insert received block", "err", err, "height", block.GetHeight(), "singer", block.GetSigner().String())
				return err
			}
		}
	}*/
	return nil
}

func (s *Sync) NeedValidation(err error) bool {
	return err != Err_RepeatBlock
}
