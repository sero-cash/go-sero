package types

import (
	"fmt"
	"math/big"
	"testing"
)

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
