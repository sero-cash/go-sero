package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/sero-cash/go-sero/common/hexutil"
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
		[]byte{},
	})
	if h3 != h2 && h2 != h1 {
		t.Errorf("Ehash must be the right type")
	}

	fmt.Printf("%v\n", h1)
	fmt.Printf("%v\n", hexutil.Encode(h2[:]))
	fmt.Printf("%v\n", hexutil.Encode(h3[:]))
}

type XX struct {
	A *big.Int
}

type st struct {
	Cmd XX
}

func TestJson(t *testing.T) {
	jsonStr := `{"cmd":null}`
	var s st
	err := json.Unmarshal([]byte(jsonStr), &s)
	if err != nil {
		fmt.Println(err)
	}
	b, _ := json.Marshal(s)
	fmt.Println(string(b))

}
