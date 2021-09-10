package kit

import (
	"fmt"
	"testing"
)

func TestCalCoinBase(t *testing.T) {
	coinbase := CalCoinBase("testnet", 10000, 100)
	if coinbase != 42796875{
		t.Fatalf("error")
	}
}

func TestGenerateTokenAddress(t *testing.T) {
	addr, err := GenerateTokenAddress("mainnet", "ABC")
	if err != nil{
		t.Fatalf(err.Error())
	}
	fmt.Println(addr)
	if !CheckContractAddress("mainnet", addr){
		t.Fatalf(err.Error())
	}
}
