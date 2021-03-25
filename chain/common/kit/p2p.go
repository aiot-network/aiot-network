package kit

import (
	"context"
	"fmt"
	"github.com/aiot-network/aiot-network/tools/crypto/ecc/secp256k1"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

func GenerateP2PID(key *secp256k1.PrivateKey) (peer.ID, error) {
	p2pPriKey, err := crypto.UnmarshalSecp256k1PrivateKey(key.Serialize())
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
