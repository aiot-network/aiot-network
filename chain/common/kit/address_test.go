package kit

import (
	"fmt"
	"github.com/aiot-network/aiot-network/common/param"
	"os"
	"testing"
)

func TestGenerateAddress(t *testing.T) {
	e, _ := Entropy()
	m, _ := Mnemonic(e)
	key, _ := MnemonicToEc(m)
	addr, _ := GenerateAddress(param.TestNet, key.PubKey().SerializeCompressedString())
	fmt.Println(addr)
	if !CheckAddress(param.MainNet, addr) {
		t.Fatalf("failed")
	}
	fmt.Println(GenerateTokenAddress(param.MainNet, addr, "SA"))
}

func TestCheckAddress(t *testing.T) {
	fmt.Println(CheckAddress("testnet", "FmcoinEaterAddressDontSend000000000"))
}
