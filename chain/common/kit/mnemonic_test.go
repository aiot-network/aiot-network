package kit

import (
	"fmt"
	"testing"
)

func TestHdDerive(t *testing.T) {
	e, _ := Entropy()
	m, _ := Mnemonic(e)
	fmt.Println(m)
	e1, err := MnemonicToSeed(m)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if e != e1 {
		t.Fatalf("wrong seed")
	}
	hdPri, err := HdDerive("mainnet", e1, 1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(hdPri)
	hdpub, err := HdPrivateToPublic(hdPri, "mainnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(hdpub)
	pub, err := HdToEc(hdpub, "mainnet")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(pub)
	address, err := GenerateAddress("mainnet", pub)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(address)
}
