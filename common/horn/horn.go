package horn

import (
	"github.com/aiot-network/aiot-network/service/gorutinue"
	"github.com/aiot-network/aiot-network/service/peers"
	"github.com/aiot-network/aiot-network/service/request"
	log "github.com/aiot-network/aiot-network/tools/log/log15"
	"github.com/aiot-network/aiot-network/types"
)

const module = "horn"

type Horn struct {
	peers   *peers.Peers
	request request.IRequestHandler
	local   *types.Peer
	gPool   *gorutinue.Pool
}

func NewHorn(peers *peers.Peers, gPool *gorutinue.Pool, request request.IRequestHandler) *Horn {
	return &Horn{
		peers:   peers,
		request: request,
		local:   peers.Local(),
		gPool:   gPool,
	}
}

func (h *Horn) BroadcastMsg(message types.IMessage) {
	peers := h.peers.PeersMap()
	for id, peer := range peers {
		if h.local == nil || id != h.local.Address.ID.String() {
			conn := peer.Conn
			if err := h.gPool.AddTask(gorutinue.NewTask(
				func() error {
					return h.request.SendMsg(conn, message)
				})); err != nil {
				log.Warn("Adding the task to send the message failed", "module", module,
					"hash", message.Hash().String(), "target", peer.Address.String())
			}
		}
	}
}

func (h *Horn) BroadcastBlock(block types.IBlock) {
	peers := h.peers.PeersMap()
	for id, peer := range peers {
		if h.local == nil || id != h.local.Address.ID.String() {
			conn := peer.Conn
			if err := h.gPool.AddTask(gorutinue.NewTask(
				func() error {
					return h.request.SendBlock(conn, block)
				})); err != nil {
				log.Warn("Adding the task to send the block failed", "module", module,
					"height", block.GetHash().String(), "target", peer.Address.String())
			}
		}
	}
}
