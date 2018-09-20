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

package zstate

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Out0 struct {
	Currency keys.Uint256
	Out      stx.Out_O
}

func (self *Out0) ToCommitment() (ret keys.Uint256) {
	return cpt.GenCommitment(&self.Currency, &self.Out.Addr, self.Out.Value.ToUint256().NewRef(), &self.Out.Memo)
}

type OutState0 struct {
	Index  uint64
	Out_O  *Out0       `rlp:"nil"`
	Desc_Z *stx.Desc_Z `rlp:"nil"`
}

func (self *OutState0) Clone() (ret OutState0) {
	utils.DeepCopy(&ret, self)
	return
}

func (out *OutState0) IsO() bool {
	if out.Desc_Z == nil {
		return true
	} else {
		return false
	}
}

func (self *OutState0) ToCommitment() *keys.Uint256 {
	if self.IsO() {
		return self.Out_O.ToCommitment().NewRef()
	} else {
		return self.Desc_Z.Out.Commitment.NewRef()
	}
}

func (self *OutState0) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type OutState0Get struct {
	out *OutState0
}

func (self *OutState0Get) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.out = nil
		return
	} else {
		self.out = &OutState0{}
		if err := rlp.DecodeBytes(v, &self.out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
