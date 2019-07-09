package types

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/rlp"
)

func TestBlockRLP(t *testing.T) {
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)

	tx := extblock{}
	tx.Header = &Header{}

	e := rlp.Encode(w, &tx)
	fmt.Println(e)
	w.Flush()

	dtx := extblock{}
	stream := rlp.NewStream(&buf, uint64(buf.Len()))
	_, size, _ := stream.Kind()
	fmt.Println(size)
	e = stream.Decode(&dtx)
	fmt.Println(e)
	fmt.Println(dtx)
}
