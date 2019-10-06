package types

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/rlp"
)

func TestRLP(t *testing.T) {
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)

	tx := Header{}
	tx.GasUsed = 10
	tx.CurrentVotes = append(tx.CurrentVotes, HeaderVote{})

	e := rlp.Encode(w, &tx)
	fmt.Println(e)
	w.Flush()

	dtx := Header{}
	stream := rlp.NewStream(&buf, uint64(buf.Len()))
	_, size, _ := stream.Kind()
	fmt.Println(size)
	e = stream.Decode(&dtx)
	fmt.Println(e)
	fmt.Println(dtx)
}
