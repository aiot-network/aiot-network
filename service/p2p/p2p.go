package p2p

import (
	"context"
	"fmt"
	"github.com/aiot-network/aiotchain/common/config"
	"github.com/aiot-network/aiotchain/common/private"
	"github.com/aiot-network/aiotchain/service/peers"
	"github.com/aiot-network/aiotchain/service/request"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/aiot-network/aiotchain/tools/utils"
	"github.com/aiot-network/aiotchain/types"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	crypto2 "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	p2pcfg "github.com/libp2p/go-libp2p/config"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

const module = "p2p"

type P2p struct {
	host       core.Host
	local      *types.Peer
	dht        *dht.IpfsDHT
	peers      *peers.Peers
	reqHandler request.IRequestHandler
	close      chan bool
	closed     chan bool
}

func NewP2p(ps *peers.Peers, reqHandler request.IRequestHandler) (*P2p, error) {
	var err error
	ser := &P2p{
		peers:      ps,
		reqHandler: reqHandler,
		close:      make(chan bool),
		closed:     make(chan bool),
	}
	if config.Param.CustomBoot != "" {
		ma, err := multiaddr.NewMultiaddr(config.Param.CustomBoot)
		if err != nil {
			return nil, fmt.Errorf("incorrect bootstrap node address， %s", err)
		}
		CustomBootPeers = append(CustomBootPeers, ma)
	}

	host, err := NewP2PHost(config.Param.IPrivate.PrivateKey(),
		config.Param.P2pPort, config.Param.ExternalIp)
	if err != nil {
		return nil, err
	}
	ser.host = host
	ser.local = types.NewPeer(config.Param.IPrivate.PrivateKey(),
		&peer.AddrInfo{
			ID:    host.ID(),
			Addrs: host.Addrs()}, nil, nil)
	ser.initP2pHandle()
	ps.SetLocal(ser.local)
	log.Info("P2p host created", "module", module, "id", host.ID(), "address", host.Addrs())
	return ser, nil
}

func NewP2PHost(private *secp256k1.PrivateKey, port, external string) (core.Host, error) {
	ips := utils.GetLocalIp()
	ips = append(ips, external)
	f := newFactory(ips, port)
	p2pKey, err := crypto2.UnmarshalSecp256k1PrivateKey(private.Serialize())
	if err != nil {
		return nil, err
	}
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.Identity(p2pKey),
		libp2p.DefaultMuxers,
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
		libp2p.AddrsFactory(f),
	}
	return libp2p.New(context.Background(), opts...)
}

func newFactory(ips []string, port string) p2pcfg.AddrsFactory {
	return func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		addrs = []multiaddr.Multiaddr{}
		for _, ip := range ips {
			extMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, port))
			if extMultiAddr != nil {
				addrs = append(addrs, extMultiAddr)
			}
		}
		return addrs
	}
}

func (p *P2p) Name() string {
	return module
}

func (p *P2p) Start() error {
	if err := p.connectBootNode(); err != nil {
		log.Error("Failed to connect the boot node!", "module", module, "error", err)
		return err
	}

	go p.peerDiscovery()
	log.Info("P2P started successfully", "module", module)
	return nil
}

func (p *P2p) Stop() error {
	close(p.close)
	if err := p.host.Close(); err != nil {
		return err
	}
	log.Info("Stop P2P server")
	return nil
}

func (p *P2p) Info() map[string]interface{} {
	return map[string]interface{}{
		"peer":    p.local.Conn.PeerId.String(),
		"address": p.local.Address.String(),
	}
}

func (p *P2p) Addr() string {
	addrs := p.host.Addrs()
	var rs string
	for _, addr := range addrs {
		rs += "[" + addr.String() + "]"
	}
	return rs
}

func (p *P2p) ID() string {
	return p.host.ID().String()
}

func (p *P2p) newStream(id peer.ID) (network.Stream, error) {
	return p.host.NewStream(context.Background(), id, protocol.ID(config.Param.P2pParam.NetWork))
}

func (p *P2p) initP2pHandle() {
	p.host.SetStreamHandler(protocol.ID(config.Param.P2pParam.NetWork), p.reqHandler.SendToReady)
}

func (p *P2p) connectBootNode() error {
	var err error
	p.dht, err = dht.New(context.Background(), p.host)
	if err != nil {
		return err
	}

	log.Info("Initializing node DHT", "module", module)
	if err = p.dht.Bootstrap(context.Background()); err != nil {
		return err
	}

	boots := DefaultBootPeers
	if len(CustomBootPeers) > 0 {
		boots = CustomBootPeers
	}
	var wg sync.WaitGroup
	for _, address := range boots {
		addrInfo, _ := peer.AddrInfoFromP2pAddr(address)
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
			defer cancel()
			if err := p.host.Connect(ctx, *addrInfo); err != nil {
				log.Warn("Failed to connection established with boot node", "module", module, "error", err)
			} else {
				log.Info("Connection established with boot node", "module", module, "peer", *addrInfo)
			}
		}()
	}
	wg.Wait()
	return nil
}

// peerDiscovery new nodes every 8s
func (p *P2p) peerDiscovery() {
	rouDis := discovery.NewRoutingDiscovery(p.dht)
	discovery.Advertise(context.Background(), rouDis, config.Param.P2pParam.NetWork)

	for {
		select {
		case _, _ = <-p.close:
			return
		default:
			//log.Info("Look for other peers...", "module", module)
			ch, err := rouDis.FindPeers(context.Background(), config.Param.P2pParam.NetWork)
			if err != nil {
				log.Error("Peer search failed", "module", module, "error", err)
				time.Sleep(time.Second * 10)
				continue
			}
			p.readAddrInfo(ch)
		}
		time.Sleep(time.Second * 8)
	}
}

func (p *P2p) readAddrInfo(addrCh <-chan peer.AddrInfo) {
	for {
		select {
		case addrInfo, ok := <-addrCh:
			if ok {
				if addrInfo.ID == p.local.Address.ID || IsBootPeers(addrInfo.ID) {
					continue
				}
				if !p.peers.AddressExist(&addrInfo) {
					if stream, err := p.connect(addrInfo.ID); err != nil {
						//log.Warn("connect failed", "addr", addrInfo.String())
						p.peers.RemovePeer(addrInfo.ID.String())
						continue
					} else {
						p.peers.AddPeer(types.NewPeer(nil, cpAddrInfo(&addrInfo), p.newStream, stream))
					}
				}
			} else {
				return
			}
		}
	}
}

func (p *P2p) connect(id peer.ID) (network.Stream, error) {
	stream, err := p.newStream(id)
	if err != nil {
		return nil, err
	}
	stream.Reset()
	stream.Close()
	return stream, nil
}

func PrivateToP2pId(key private.IPrivate) (peer.ID, error) {
	p2pPriKey, err := crypto2.UnmarshalSecp256k1PrivateKey(key.Serialize())
	if err != nil {
		return "", err
	}
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", "65535")),
		libp2p.Identity(p2pPriKey),
	}
	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return "", err
	}
	defer host.Close()
	return host.ID(), nil
}

func cpAddrInfo(addr *peer.AddrInfo) *peer.AddrInfo {
	bytes, _ := addr.MarshalJSON()
	destAddr := new(peer.AddrInfo)
	destAddr.UnmarshalJSON(bytes)
	return destAddr
}
