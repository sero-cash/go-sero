package account

import (
	"fmt"
	"regexp"
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

var reg1, _ = regexp.Compile(`^(.*)0(.*)0(.*)$`)

func TestRex(t *testing.T) {
	a := "SP103ATBFhrrwJhbLGd7PEGZeTg3uFjjknyegqvDyN5pvM4gvp4jyfyGTTLvwEwJhwqBjS2ceaQjN7uHC5Q5biifbRZf0NF"
	strs := reg1.FindStringSubmatch(a)
	fmt.Println(strs)
}
