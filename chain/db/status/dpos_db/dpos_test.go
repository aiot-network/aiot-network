package dpos_db

import (
	"github.com/aiot-network/aiot-network/tools/arry"
	"testing"
)

func TestDPosDB_SaveVoter(t *testing.T) {
	dposDB, err := Open("test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	dposDB.SetRoot(arry.Hash{})
	dposDB.SaveVoter(arry.StringToAddress("A"), arry.StringToAddress("B"))
	dposDB.SaveVoter(arry.StringToAddress("C"), arry.StringToAddress("B"))
	dposDB.SaveVoter(arry.StringToAddress("E"), arry.StringToAddress("F"))
	dposDB.SaveVoter(arry.StringToAddress("G"), arry.StringToAddress("H"))

	rs := dposDB.Voters()
	count := 0
	other := 0
	for _, addr := range rs {
		if addr.IsEqual(arry.StringToAddress("A")) || addr.IsEqual(arry.StringToAddress("C")) {
			count++
		} else {
			other++
		}
	}

	if count == 2 && other == 0 {
		t.Log("pass")
	} else {
		t.Fatalf("failed")
	}
}
