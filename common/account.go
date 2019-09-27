package common

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-czero-import/c_type"
)

type AccountKey [64]byte

func (a AccountKey) String() string {
	return base58.Encode(a[:])
}

func (a AccountKey) Bytes() []byte {
	return a[:]
}

func (a AccountKey) ToUint512() (ret c_type.Uint512) {
	copy(ret[:], a[:])
	return
}

type TkAddress [64]byte

func Base58ToTk(bs string) (ret TkAddress) {
	b := base58.Decode(bs)
	copy(ret[:], b)
	return
}

func (t TkAddress) String() string {
	return base58.Encode(t[:])
}

func (a TkAddress) Bytes() []byte {
	return a[:]
}

func (a TkAddress) ToTk() (ret c_type.Tk) {
	copy(ret[:], a[:])
	return
}
