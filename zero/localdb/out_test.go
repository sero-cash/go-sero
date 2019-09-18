package localdb

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/rlp"
)

type Test struct {
	I0 int
	I1 int
}

type Test1 struct {
	I0 int
	I1 int
	I2 int
}

func TestRlp(t *testing.T) {
	test := Test{1, 2}
	bs, _ := rlp.EncodeToBytes(&test)

	var test1 Test1
	e := rlp.DecodeBytes(bs, &test1)
	fmt.Println(e)
	fmt.Println(test1.I2)
}
