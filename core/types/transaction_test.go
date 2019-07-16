package types

import (
	"bufio"
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/sero-cash/go-sero/rlp"
)

type Uint256 [32]byte

func TestRLPEhash(t *testing.T) {

	var testByte [32]byte
	var testByte1 Uint256
	bt := big.NewInt(34).Bytes()
	copy(testByte[:], bt[:])
	copy(testByte1[:], bt[:])

	h1 := rlpHash([]interface{}{
		testByte,
	})

	h2 := rlpHash([]interface{}{
		testByte1,
	})

	fmt.Printf("%v\n", h1)
	fmt.Printf("%v\n", h2)
}

func TestEhash(t *testing.T) {
	price := big.NewInt(40)
	gasLimit := uint64(22)
	h1 := rlpHash([]interface{}{
		*price,
		gasLimit,
	})
	h2 := rlpHash([]interface{}{
		&price,
		gasLimit,
	})
	h3 := rlpHash([]interface{}{
		price,
		gasLimit,
	})
	if h3 != h2 && h2 != h1 {
		t.Errorf("Ehash must be the right type")
	}

	fmt.Printf("%v\n", h1)
	fmt.Printf("%v\n", h2)
	fmt.Printf("%v\n", h3)
}

func TestHeader(t *testing.T) {
	hs := []Header{}
	hs = append(hs, Header{})

	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)
	e := rlp.Encode(w, &hs)
	fmt.Println(e)
	w.Flush()

	hs_d := []Header{}
	stream := rlp.NewStream(&buf, uint64(buf.Len()))
	_, size, _ := stream.Kind()
	fmt.Println(size)
	e = stream.Decode(&hs_d)
	fmt.Println(e)
}
