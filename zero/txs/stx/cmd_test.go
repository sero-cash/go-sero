package stx

import (
	"bytes"
	"testing"
)

func TestContractData_UnmarshalText(t *testing.T) {
	str1 := "0x3c87d061b8bfc306c0882e8b8a43627b0000fa30b251000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000046464646400000000000000000000000000000000000000000000000000000000"
	var r1 ContractData
	input := []byte(str1)
	err := r1.UnmarshalText(input)
	if err != nil {
		t.Error(err)
	}

	str2 := "PIfQYbi/wwbAiC6LikNiewAA+jCyUQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAARkZGRkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
	var r2 ContractData
	err = r2.UnmarshalText([]byte(str2))
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(r1[:], r2[:]) != 0 {
		t.Error("not equals")
	}
}
