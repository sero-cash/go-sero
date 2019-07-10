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
	"sync/atomic"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

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

func (self *In_Z) ToHash_for_sign() (ret keys.Uint256) {
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
	RPK     keys.Uint256
	EInfo   cpt.Einfo
	PKr     keys.PKr
	Proof   cpt.Proof
}

func ConfirmOut_Z(deInfo *cpt.InfoDesc, out_z *Out_Z) (e error) {
	desc := cpt.ConfirmOutputDesc{}
	desc.Memo = deInfo.Memo
	desc.Tkn_currency = deInfo.Tkn_currency
	desc.Tkn_value = deInfo.Tkn_value
	desc.Tkt_category = deInfo.Tkt_category
	desc.Tkt_value = deInfo.Tkt_value
	desc.Rsk = deInfo.Rsk
	desc.Pkr = out_z.PKr
	desc.Out_cm = out_z.OutCM
	e = cpt.ConfirmOutput(&desc)
	return
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

func (self *Out_Z) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.PKr[:])
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Out_Z) ToHash_for_sign() (ret keys.Uint256) {
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

func (self *Desc_Z) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	for _, out := range self.Outs {
		d.Write(out.ToHash_for_gen().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Desc_Z) ToHash_for_sign() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash_for_sign().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash_for_sign().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

type T struct {
	Ehash    keys.Uint256
	From     keys.PKr
	Fee      assets.Token
	Sign     keys.Uint512
	Bcr      keys.Uint256
	Bsign    keys.Uint512
	Desc_Z   Desc_Z
	Desc_O   Desc_O
	Desc_Pkg PkgDesc_Z
	Desc_Cmd DescCmd

	//cache
	hash atomic.Value
}

func (self *T) ToHash() (ret keys.Uint256) {
	if h, ok := self.hash.Load().(keys.Uint256); ok {
		ret = h
		return
	}
	v := self._ToHash()
	self.hash.Store(v)
	return v
}

func (self *T) _ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	d.Write(self.Desc_Z.ToHash().NewRef()[:])
	d.Write(self.Desc_O.ToHash().NewRef()[:])
	d.Write(self.Desc_Pkg.ToHash().NewRef()[:])
	d.Write(self.Sign[:])
	d.Write(self.Bcr[:])
	d.Write(self.Bsign[:])
	if self.Desc_Cmd.Count() > 0 {
		d.Write(self.Desc_Cmd.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	d.Write(self.Desc_Z.ToHash_for_gen().NewRef()[:])
	d.Write(self.Desc_O.ToHash_for_gen().NewRef()[:])
	d.Write(self.Desc_Pkg.ToHash_for_gen().NewRef()[:])
	if self.Desc_Cmd.Count() > 0 {
		d.Write(self.Desc_Cmd.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash_for_sign() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	d.Write(self.Desc_Z.ToHash_for_sign().NewRef()[:])
	d.Write(self.Desc_O.ToHash_for_sign().NewRef()[:])
	d.Write(self.Desc_Pkg.ToHash_for_sign().NewRef()[:])
	if self.Desc_Cmd.Count() > 0 {
		d.Write(self.Desc_Cmd.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}
