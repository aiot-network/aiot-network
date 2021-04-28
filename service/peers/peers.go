package peers

import (
	request2 "github.com/aiot-network/aiotchain/service/request"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"math/rand"
	"sync"
	"time"
)

const (
	module             = "peers"
	maxPeers           = 1000000
	monitoringInterval = 60 * 30
)

type Peers struct {
	local      *types.Peer
	cache      map[string]*types.Peer
	remove     map[string]*types.Peer
	idList     []string
	rwm        sync.RWMutex
	close      chan bool
	peerInfo   map[string]*types.Local
	infoWm     sync.RWMutex
	reqHandler request2.IRequestHandler
}

func NewPeers(reqHandler request2.IRequestHandler) *Peers {
	return &Peers{
		cache:      make(map[string]*types.Peer, maxPeers),
		remove:     make(map[string]*types.Peer, maxPeers),
		close:      make(chan bool),
		peerInfo:   make(map[string]*types.Local),
		reqHandler: reqHandler,
	}
}

func (p *Peers) Name() string {
	return module
}

func (p *Peers) Start() error {
	log.Info("Peers started successfully", "module", module)
	go p.monitoring()
	go p.peerLocal()
	return nil
}

func (p *Peers) Stop() error {
	log.Info("Peers was stopped", "module", module)
	return nil
}

func (p *Peers) Info() map[string]interface{} {
	return map[string]interface{}{
		"connections": p.Count(),
	}
}

func (p *Peers) AddressExist(address *peer.AddrInfo) bool {
	p.rwm.RLock()
	defer p.rwm.RUnlock()

	if _, ok := p.cache[address.ID.String()]; !ok {
		return false
	}
	return true
}

func (p *Peers) AddPeer(peer *types.Peer) {
	p.rwm.Lock()
	defer p.rwm.Unlock()

	if len(p.cache) >= maxPeers {
		return
	}
	if peer.Speed == 0 {
		peer.Speed = 500
	}
	p.cache[peer.Address.ID.String()] = peer
	p.idList = append(p.idList, peer.Address.ID.String())
	log.Info("Add a peer", "module", module, "id", peer.Address.ID.String(), "address", peer.Address.String())
}

func (p *Peers) RemovePeer(reId string) {
	p.rwm.Lock()
	defer p.rwm.Unlock()

	for index, id := range p.idList {
		if id == reId {
			p.idList = append(p.idList[0:index], p.idList[index+1:]...)
			if peer, ok := p.cache[reId]; ok {
				delete(p.cache, reId)
				p.remove[reId] = peer
				log.Info("Delete a peer", "id", reId)
			}
			break
		}
	}
}

func (p *Peers) SetSpeed(id string, speed uint64) {
	p.rwm.Lock()
	defer p.rwm.Unlock()

	if peer, ok := p.cache[id]; ok {
		p.cache[peer.Address.ID.String()].Speed = speed
	}
}

func (p *Peers) monitoring() {
	t := time.NewTicker(time.Second * monitoringInterval)
	defer t.Stop()
	for {
		select {
		case _ = <-t.C:
			for id, peer := range p.remove {
				if id != p.local.Address.ID.String() {
					if p.isAlive(peer) && !p.AddressExist(peer.Address) {
						p.AddPeer(peer)
					}
				}
			}
			for id, peer := range p.cache {
				if id != p.local.Address.ID.String() {
					if !p.isAlive(peer) {
						p.RemovePeer(id)
					}
				}
			}
		}
	}
}

func (p *Peers) peerLocal() {
	t := time.NewTicker(time.Second * 60)
	defer t.Stop()

	for {
		select {
		case _ = <-t.C:
			p.requestPeerLocal()
		}
	}
}

func (p *Peers) requestPeerLocal() {
	peers := p.PeersMap()
	for id, peer := range peers {
		if id != p.local.Address.ID.String() {
			local, err := p.reqHandler.LocalInfo(peer.Conn)
			if err == nil {
				p.infoWm.Lock()
				p.peerInfo[id] = local
				p.infoWm.Unlock()
			}
		}
	}
}

func (p *Peers) isAlive(peer *types.Peer) bool {
	stream, err := peer.Conn.Create(peer.Address.ID)
	if err != nil {
		return false
	}
	stream.Reset()
	stream.Close()
	return true
}

func (p *Peers) RandomPeer() *types.Peer {
	p.rwm.Lock()
	defer p.rwm.Unlock()

	if len(p.idList) == 0 {
		return nil
	}
	index := rand.New(rand.NewSource(time.Now().Unix())).Int31n(int32(len(p.idList)))
	peerId := p.idList[index]
	return p.cache[peerId]
}

func (p *Peers) Local() *types.Peer {
	return p.local
}

func (p *Peers) SetLocal(local *types.Peer) {
	p.local = local
}

func (p *Peers) PeersMap() map[string]*types.Peer {
	p.rwm.RLock()
	defer p.rwm.RUnlock()

	re := make(map[string]*types.Peer)
	for key, value := range p.cache {
		re[key] = value
	}
	return re
}

func (p *Peers) PeersInfo() []*types.Local {
	p.infoWm.RLock()
	defer p.infoWm.RUnlock()

	re := make([]*types.Local, len(p.peerInfo))
	i := 0
	for _, value := range p.peerInfo {
		re[i] = value
		i++
	}
	return re
}

func (p *Peers) Count() uint32 {
	p.rwm.RLock()
	defer p.rwm.RUnlock()

	count := uint32(len(p.cache))
	return count
}

func (p *Peers) Peer(id string) *types.Peer {
	p.rwm.RLock()
	defer p.rwm.RUnlock()

	peer := p.cache[id]
	return peer
}
