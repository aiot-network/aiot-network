package hash

import (
	"crypto/sha256"
	"github.com/aiot-network/aiot-network/tools/arry"
	"golang.org/x/crypto/ripemd160"
	"io"
)

const HashLength = 32

func Hash(bytes []byte) arry.Hash {
	sum256 := sha256.Sum256(bytes)
	h := sum256[:]
	return arry.BytesToHash(h)
}

func Hash160(bytes []byte) ([]byte, error) {
	hasher := ripemd160.New()
	_, err := io.WriteString(hasher, string(bytes))
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}
