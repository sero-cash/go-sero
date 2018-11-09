// copyright 2018 The sero.cash Authors
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

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
)

type Proof [cpt.PROOF_WIDTH]byte

func (b Proof) MarshalText() ([]byte, error) {
	result := make([]byte, len(b)*2+2)
	copy(result, `0x`)
	hex.Encode(result[2:], b[:])
	return result, nil
}

func (self *Proof) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self[:])
	copy(ret[:], d.Sum(nil))
	return
}

type In_Z struct {
	Anchor  keys.Uint256
	Nil     keys.Uint256
	Trace   keys.Uint256
	AssetCM keys.Uint256
	Proof   cpt.Proof
}

func (self *In_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Anchor[:])
	d.Write(self.Nil[:])
	d.Write(self.Trace[:])
	d.Write(self.AssetCM[:])
	d.Write(self.Proof.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

type Out_Z struct {
	AssetCM keys.Uint256
	OutCM   keys.Uint256
	EInfo   [cpt.INFO_WIDTH]byte `json:"-"`
	PKr     keys.Uint512
	Proof   cpt.Proof
}

func (self *Out_Z) Clone() (ret Out_Z) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Out_Z) ToRef() (ret *Out_Z) {
	ret = &Out_Z{}
	*ret = this
	return
}

func (self *Out_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.AssetCM[:])
	d.Write(self.OutCM[:])
	d.Write(self.EInfo[:])
	d.Write(self.PKr[:])
	d.Write(self.Proof.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

type Desc_Z struct {
	Ins  []In_Z
	Outs []Out_Z
}

func (self *Desc_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

type T struct {
	Ehash  keys.Uint256
	From   keys.Uint512
	Fee    utils.U256
	Sign   keys.Uint512
	Bcr    keys.Uint256
	Bsign  keys.Uint512
	Desc_Z Desc_Z
	Desc_O Desc_O
}

func (self *T) ToHash_for_z() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToUint256().NewRef()[:])
	d.Write(self.Desc_O.ToHash_for_z().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash_for_o() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.ToHash_for_z().NewRef()[:])
	d.Write(self.Desc_Z.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.ToHash_for_o().NewRef()[:])
	d.Write(self.Sign[:])
	for _, in := range self.Desc_O.Ins {
		d.Write(in.Sign[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}
