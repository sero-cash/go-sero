package utils

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/sero-cash/go-sero/common/hexutil"
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

func TestU256_ToBEBytes(t *testing.T) {
	//a := big.NewInt(339014133842837404430)
	//b := U256(*a)
	b := []byte("\"339014133842837404430\"")
	a := U256{}
	err := json.Unmarshal(b, &a)
	if err != nil {
		t.Error(err)
	}

	r := a.ToBEBytes()
	fmt.Println(hexutil.Encode(r))
}
