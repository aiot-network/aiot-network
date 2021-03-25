package kit

import (
	"fmt"
	"testing"
)

func TestCalCoinBase(t *testing.T) {
	x := CalCoinBase("testnet", 10000, 100)
	fmt.Println(x)
}
