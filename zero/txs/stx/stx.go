// Copyright 2015 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package stx

import (
	"encoding/hex"
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
)

type Proof struct {
	G [cpt.PROOF_WIDTH]byte
}

func (b Proof) MarshalText() ([]byte, error) {
	result := make([]byte, len(b.G)*2+2)
	copy(result, `0x`)
	hex.Encode(result[2:], b.G[:])
	return result, nil
}

func (self *Proof) ToHash_for_o() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.G[:])
	copy(ret[:], d.Sum(nil))
	return
}

type In_Z struct {
	Anchor keys.Uint256
	Nil    keys.Uint256
	Trace  keys.Uint256
}

func (self *In_Z) ToHash_for_o() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Anchor[:])
	d.Write(self.Nil[:])
	d.Write(self.Trace[:])
	copy(ret[:], d.Sum(nil))
	return
}

type Out_Z struct {
	Commitment keys.Uint256
	EInfo      [cpt.ETEXT_WIDTH]byte `json:"-"`
}

func (self *Out_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Commitment[:])
	d.Write(self.EInfo[:])
	copy(ret[:], d.Sum(nil))
	return
}

type Desc_Z struct {
	In    In_Z
	Out   Out_Z
	R     keys.Uint256
	S1    keys.Uint256
	Proof Proof
}

func (self *Desc_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.In.ToHash_for_o().NewRef()[:])
	d.Write(self.Out.ToHash().NewRef()[:])
	d.Write(self.R[:])
	d.Write(self.S1[:])
	d.Write(self.Proof.ToHash_for_o().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

type T struct {
	Ehash   keys.Uint256
	From    keys.Uint512
	Desc_Zs []Desc_Z
	Desc_Os []Desc_O
}

func (self *T) ToHash_for_o() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	for _, desc_z := range self.Desc_Zs {
		d.Write(desc_z.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash_for_z() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	for _, desc := range self.Desc_Os {
		d.Write(desc.ToHash_for_z().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.ToHash_for_o().NewRef()[:])
	d.Write(self.ToHash_for_z().NewRef()[:])

	for _, desc_o := range self.Desc_Os {
		for _, in := range desc_o.Ins {
			d.Write(in.Sign[:])
		}
	}
	copy(ret[:], d.Sum(nil))
	return
}
