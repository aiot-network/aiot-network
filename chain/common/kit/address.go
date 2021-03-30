package kit

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aiot-network/aiot-network/common/param"
	"github.com/aiot-network/aiot-network/tools/arry"
	"github.com/aiot-network/aiot-network/tools/crypto/base58"
	"github.com/aiot-network/aiot-network/tools/crypto/ecc/secp256k1"
	"github.com/aiot-network/aiot-network/tools/crypto/hash"
)

const addressLength = 35
const addressBytesLength = 26

func GenerateAddress(net string, pubKey string) (string, error) {
	ver := []byte{}
	switch net {
	case param.MainNet:
		ver = append(ver, param.MainNetParam.PubKeyHashAddrID[0:]...)
	case param.TestNet:
		ver = append(ver, param.TestNetParam.PubKeyHashAddrID[0:]...)
	default:
		return "", errors.New("wrong network")
	}

	pubBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return "", fmt.Errorf("wrong public key, error:%s", err.Error())
	}
	key, err := secp256k1.ParsePubKey(pubBytes)
	if err != nil {
		return "", fmt.Errorf("wrong public key, error:%s", err.Error())
	}
	hashed256 := hash.Hash(key.SerializeCompressed())
	hashed160, err := hash.Hash160(hashed256.Bytes())
	if err != nil {
		return "", err
	}

	addNet := append(ver, hashed160...)
	hashed1 := hash.Hash(addNet)
	hashed2 := hash.Hash(hashed1.Bytes())
	checkSum := hashed2[0:4]
	hashedCheck1 := append(addNet, checkSum...)
	return arry.StringToAddress(base58.Encode(hashedCheck1)).String(), nil
}

func generateAddress(ver []byte, pubKey string) (string, error) {
	pubBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return "", fmt.Errorf("wrong public key, error:%s", err.Error())
	}
	key, err := secp256k1.ParsePubKey(pubBytes)
	if err != nil {
		return "", fmt.Errorf("wrong public key, error:%s", err.Error())
	}
	hashed256 := hash.Hash(key.SerializeCompressed())
	hashed160, err := hash.Hash160(hashed256.Bytes())
	if err != nil {
		return "", err
	}

	addNet := append(ver, hashed160...)
	hashed1 := hash.Hash(addNet)
	hashed2 := hash.Hash(hashed1.Bytes())
	checkSum := hashed2[0:4]
	hashedCheck1 := append(addNet, checkSum...)

	return arry.StringToAddress(base58.Encode(hashedCheck1)).String(), nil
}

func CheckAddress(net string, address string) bool {
	ver := []byte{}
	switch net {
	case param.MainNet:
		if address == param.MainNetParam.EaterAddress.String() {
			return true
		}
		ver = append(ver, param.MainNetParam.PubKeyHashAddrID[0:]...)
	case param.TestNet:
		if address == param.TestNetParam.EaterAddress.String() {
			return true
		}
		ver = append(ver, param.TestNetParam.PubKeyHashAddrID[0:]...)
	default:
		return false
	}
	if len(address) != addressLength {
		return false
	}
	addrBytes := base58.Decode(address)
	if len(addrBytes) != addressBytesLength {
		return false
	}
	checkSum := addrBytes[len(addrBytes)-4:]
	checkBytes := addrBytes[0 : len(addrBytes)-4]
	checkBytesHashed1 := hash.Hash(checkBytes)
	checkBytesHashed2 := hash.Hash(checkBytesHashed1.Bytes())
	netBytes := checkBytes[0:2]
	if bytes.Compare(ver, netBytes) != 0 {
		return false
	}
	return bytes.Compare(checkSum, checkBytesHashed2[0:4]) == 0
}
