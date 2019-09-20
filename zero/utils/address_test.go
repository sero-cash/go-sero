package utils

import (
	"fmt"
	"testing"
)

func TestAddress(t *testing.T) {
	bytes := []byte("hello address")
	addr := NewAddressByBytes(bytes)
	fmt.Println(addr.ToCode())

	hexs := "0x12346889093334"
	addr1, e := NewAddressByHex(hexs)
	fmt.Println(addr1.ToCode(), e)
	fmt.Println(addr1.ToHex())

	addr2, e := NewAddressByString(addr.ToCode())
	fmt.Println(addr2.ToCode())
}
