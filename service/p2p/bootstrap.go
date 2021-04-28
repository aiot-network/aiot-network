package p2p

import (
	"context"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	log "github.com/aiot-network/aiotchain/tools/log/log15"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"strings"
)

// Default boot node list
var DefaultBootPeers []multiaddr.Multiaddr

// Custom boot node list
var CustomBootPeers []multiaddr.Multiaddr

func init() {
	for _, s := range []string{
		//103.68.63.164
		"/ip4/127.0.0.1/tcp/19563/ipfs/16Uiu2HAkwVbA7r9wEA5BnHwe6XpdAbLj5GUffwggg9TQfsPWfK59",
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		DefaultBootPeers = append(DefaultBootPeers, ma)
	}
}

func IsBootPeers(id peer.ID) bool {
	bootstrap := DefaultBootPeers
	if len(CustomBootPeers) > 0 {
		bootstrap = CustomBootPeers
	}
	for _, bootstrap := range bootstrap {
		if id.String() == strings.Split(bootstrap.String(), "/")[6] {
			return true
		}
	}
	return false
}

func NewBoot(port, external string, private *secp256k1.PrivateKey) (*P2p, error) {
	host, err := NewP2PHost(private, port, external)
	if err != nil {
		return nil, err
	}
	p2p := &P2p{host: host}
	log.Info("Host created", "id", p2p.host.ID(), "address", p2p.host.Addrs())
	return p2p, nil
}

func (p *P2p) StartBoot() error {
	var err error
	p.dht, err = dht.New(context.Background(), p.host)
	if err != nil {
		return err
	}
	log.Info("Start the boot node", "module", module)
	if err = p.dht.Bootstrap(context.Background()); err != nil {
		return err
	}
	return nil
}
