package kit

import (
	"fmt"
	"github.com/aiot-network/aiotchain/common/param"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	"testing"
)

func TestGenerateAddress(t *testing.T) {
	key, _ := secp256k1.ParseStringToPrivate("cbb838da3e01d02946afdf6d6394ca79cb07068048503ce5b0ff1c1e65de3eac")

	addr, _ := GenerateAddress(param.TestNet, key.PubKey().SerializeCompressedString())
	fmt.Println(addr)
	if !CheckAddress(param.TestNet, addr) {
		t.Fatalf("failed")
	}
	fmt.Println(GenerateTokenAddress(param.TestNet, "SA"))
}

func TestCheckAddress(t *testing.T) {
	fmt.Println(CheckAddress("testnet", "AicoinEaterAddressDontSend000000000"))
}
