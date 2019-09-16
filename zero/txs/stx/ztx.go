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

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

type Tx2 struct {
}

type T struct {
	Ehash    c_type.Uint256
	From     c_type.PKr
	Fee      assets.Token
	Sign     c_type.Uint512
	Bcr      c_type.Uint256
	Bsign    c_type.Uint512
	Desc_O   stx_v0.Desc_O
	Desc_Z   stx_v0.Desc_Z
	Desc_Pkg PkgDesc_Z
	Desc_Cmd DescCmd
	Tx1      stx_v1.Tx

	//cache
	hash        atomic.Value
	feeCC_Szk   atomic.Value
	feeCC_Czero atomic.Value
}

func (self *T) Tx0() (ret *stx_v0.Tx) {
	if self.Desc_O.HasContent() || self.Desc_Z.HasContent() {
		ret = &stx_v0.Tx{}
		ret.Desc_Z = self.Desc_Z
		ret.Desc_O = self.Desc_O
		return
	} else {
		return
	}
}

func (self *T) ContractAsset() *assets.Asset {
	if self.Desc_Cmd.Contract != nil {
		return &self.Desc_Cmd.Contract.Asset
	} else {
		for _, desc_o := range self.Desc_O.Outs {
			return &desc_o.Asset
		}
	}
	return nil
}

func (self *T) ContractAddress() *c_type.PKr {
	if self.Desc_Cmd.Contract != nil {
		return self.Desc_Cmd.Contract.To
	} else {
		for _, out := range self.Desc_O.Outs {
			if out.Addr != (c_type.PKr{}) {
				return &out.Addr
			}
			return nil
		}
	}
	return nil
}

func (self *T) IsOpContract() bool {
	if self.Desc_Cmd.Contract != nil {
		return true
	} else {
		if len(self.Desc_O.Outs) > 0 {
			return true
		}
	}
	return false
}

func (self *T) ToFeeCC_Czero() c_type.Uint256 {
	if cc, ok := self.feeCC_Czero.Load().(c_type.Uint256); ok {
		return cc
	}
	asset_desc := c_czero.AssetDesc{
		Asset: self.Fee.ToTypeAsset(),
	}
	c_czero.GenAssetCC(&asset_desc)
	v := asset_desc.Asset_cc
	self.feeCC_Czero.Store(v)
	return v
}

func (self *T) ToFeeCC_Szk() c_type.Uint256 {
	if cc, ok := self.feeCC_Szk.Load().(c_type.Uint256); ok {
		return cc
	}
	v, _ := c_superzk.GenAssetCC(self.Fee.ToTypeAsset().NewRef())
	self.feeCC_Szk.Store(v)
	return v
}

func (self *T) ToHash() (ret c_type.Uint256) {
	if h, ok := self.hash.Load().(c_type.Uint256); ok {
		ret = h
		return
	}
	v := self._ToHash()
	self.hash.Store(v)
	return v
}

func (self *T) _ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	d.Write(self.Desc_Z.ToHash().NewRef()[:])
	d.Write(self.Desc_O.ToHash().NewRef()[:])
	if self.Tx1.Count() > 0 {
		d.Write(self.Tx1.ToHash().NewRef()[:])
	}
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

func (self *T) ToHash_for_gen() (ret c_type.Uint256) {
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

func (self *T) ToHash_for_sign() (ret c_type.Uint256) {
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
