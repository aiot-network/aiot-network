package codec

import (
	"fmt"
	"testing"
)

func TestUint64toBytes(t *testing.T) {
	fmt.Println(Uint64toBytes(12345678910))
}
