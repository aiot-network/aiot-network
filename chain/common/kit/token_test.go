package kit

import (
	"testing"
)

func TestCalCoinBase(t *testing.T) {
	coinbase := CalCoinBase("testnet", 10000, 100)
	if coinbase != 42796875{
		t.Fatalf("error")
	}
}
