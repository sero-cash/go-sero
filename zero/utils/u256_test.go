package utils

import (
	"encoding/json"
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
	m, _ := json.Marshal(b)
	fmt.Print(string(m))
}

func TestU256_M(t *testing.T) {
	u := U256_0
	b, e := u.MarshalJSON()
	fmt.Println((e))
	fmt.Println(string(b))
}

func TestU256_UnmarshalText(t *testing.T) {
	b := []byte("\"750000000000000000000000000\"")
	a := U256{}
	//err := a.UnmarshalJSON(b)
	//if err != nil {
	//	t.Error(err)
	//}

	//err = a.UnmarshalText(b)
	err := json.Unmarshal(b, &a)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v", a)
}
