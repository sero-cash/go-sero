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

package txstate

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type OutState struct {
	Index  uint64
	Out_O  *stx.Out_O    `rlp:"nil"`
	Out_Z  *stx.Out_Z    `rlp:"nil"`
	OutCM  *keys.Uint256 `rlp:"nil"`
	RootCM *keys.Uint256 `rlp:"nil"`
}

func (self *OutState) Clone() (ret OutState) {
	utils.DeepCopy(&ret, self)
	return
}

func (out *OutState) IsO() bool {
	if out.Out_Z == nil {
		return true
	} else {
		return false
	}
}

func (self *OutState) ToIndexRsk() (ret keys.Uint256) {
	ret = utils.NewU256(self.Index).ToRef().ToUint256()
	return
}
func (self *OutState) ToOutCM() *keys.Uint256 {
	if self.IsO() {
		if self.OutCM == nil {
			asset := self.Out_O.Asset.ToFlatAsset()
			cm := cpt.GenOutCM(
				asset.Tkn.Currency.NewRef(),
				asset.Tkn.Value.ToUint256().NewRef(),
				asset.Tkt.Category.NewRef(),
				asset.Tkt.Value.NewRef(),
				&self.Out_O.Memo,
				&self.Out_O.Addr,
				self.ToIndexRsk().NewRef(),
			)
			self.OutCM = &cm
		}
		return self.OutCM
	} else {
		return self.Out_Z.OutCM.NewRef()
	}
}

func (self *OutState) ToRootCM() *keys.Uint256 {
	if self.RootCM == nil {
		out_cm := self.ToOutCM()
		cm := cpt.GenRootCM(self.Index, out_cm)
		self.RootCM = &cm
	}
	return self.RootCM
}

func (self *OutState) ToPKr() *keys.PKr {
	if self.IsO() {
		return &self.Out_O.Addr
	} else {
		return &self.Out_Z.PKr
	}
}

func (self *OutState) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type OutState0Get struct {
	out *OutState
}

func (self *OutState0Get) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.out = nil
		return
	} else {
		self.out = &OutState{}
		if err := rlp.DecodeBytes(v, &self.out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
