package utils

import "C"
import (
	"errors"

	"github.com/btcsuite/btcutil/base58"
)

func Base58Encode(bytes []byte) (ret *string) {
	str := base58.Encode(bytes)
	if len(str) > len(bytes) {
		ret = &str
		return
	} else {
		return
	}
}

func Base58Decode(str string, bytes []byte) (e error) {
	bs := base58.Decode(str)
	if len(bs) <= len(bytes) {
		copy(bytes, bs)
		return
	} else {
		e = errors.New("base58 can not decode string")
		return
	}
}
