package types

import (
	"fmt"
	"github.com/aiot-network/aiotchain/tools/amount"
	"testing"
)

func TestTokenRecord_Check(t1 *testing.T) {
	amount1 := 4699999999999999999
	amount2 := 4699999999999999999
	amount3 := amount.Amount(amount1 + amount2)
	fmt.Println(amount3)
	fmt.Println(amount3.ToCoin())
}
