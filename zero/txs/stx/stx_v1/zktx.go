package stx_v1

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
)

type Tx struct {
	Ins_P  []In_P
	Ins_P0 []In_P0
	Ins_C  []In_C
	Outs_C []Out_C
	Outs_P []Out_P
}

func (self *Tx) Count() (ret int) {
	ret += len(self.Ins_P0)
	ret += len(self.Ins_P)
	ret += len(self.Ins_C)
	ret += len(self.Outs_C)
	ret += len(self.Outs_P)
	return
}

func (self *Tx) Tx1_Hash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins_P0 {
		d.Write(in.Tx1_Hash().NewRef()[:])
	}
	for _, in := range self.Ins_P {
		d.Write(in.Tx1_Hash().NewRef()[:])
	}
	for _, in := range self.Ins_C {
		d.Write(in.Tx1_Hash().NewRef()[:])
	}
	for _, out := range self.Outs_C {
		d.Write(out.Tx1_Hash().NewRef()[:])
	}
	for _, out := range self.Outs_P {
		d.Write(out.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Tx) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins_P0 {
		d.Write(in.ToHash().NewRef()[:])
	}
	for _, in := range self.Ins_P {
		d.Write(in.ToHash().NewRef()[:])
	}
	for _, in := range self.Ins_C {
		d.Write(in.ToHash().NewRef()[:])
	}
	for _, out := range self.Outs_C {
		d.Write(out.ToHash().NewRef()[:])
	}
	for _, out := range self.Outs_P {
		d.Write(out.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}
