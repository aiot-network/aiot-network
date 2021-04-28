package private

import (
	"github.com/aiot-network/aiotchain/tools/arry"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
)

type IPrivate interface {
	Create(network string, file string, key string) error
	Load(file string, key string) error
	Serialize() []byte
	PrivateKey() *secp256k1.PrivateKey
	Address() arry.Address
}
