package stx

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/rlp"
)

func TestRLP(t *testing.T) {
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)

	tx := T{}
	tx.Fee.Value = utils.NewU256(2)
	tx.Fee.Currency = utils.CurrencyToUint256("SERO")
	tx.Desc_Cmd.RegistPool = &RegistPoolCmd{}
	tx.Desc_Cmd.RegistPool.Value = utils.NewU256(3)
	tx.Desc_Cmd.RegistPool.FeeRate = 10

	e := rlp.Encode(w, &tx)
	fmt.Println(e)
	w.Flush()

	dtx := T{}
	stream := rlp.NewStream(&buf, uint64(buf.Len()))
	_, size, _ := stream.Kind()
	fmt.Println(size)
	e = stream.Decode(&dtx)
	fmt.Println(e)
	fmt.Println(dtx)
}
