package utils

import (
	"fmt"
	"math/big"
	"testing"
)

func TestU256_MarshalText(t *testing.T) {
	a := big.NewInt(75)
	a.Mul(a, big.NewInt(100000000000))
	a.Mul(a, big.NewInt(100000000000000))
	value := a
	b := U256(*value)
	m, _ := b.MarshalText()
	fmt.Print(string(m))
}
func TestU256_UnmarshalText(t *testing.T) {
	b := []byte("750000000000000000000000000")
	a := U256{}
	a.UnmarshalJSON(b)
	fmt.Printf("%v", a)
}
