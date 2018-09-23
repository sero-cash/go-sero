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
	"github.com/sero-cash/go-sero/zero/witness"
)

type OutState1 struct {
	Witness witness.Witness
	Index   uint64
	Num     uint64
	Tk      keys.Uint512
	Out_O   Out0
	Desc_Z  *stx.Desc_Z `rlp:"nil"`
	Nil     keys.Uint256
	Z       bool
}

func (self *OutState1) IsMine(tk *keys.Uint512) bool {
	if self.Tk == *tk {
		return true
	} else {
		return false
	}
}

func (self *OutState1) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}
func (self *OutState1) Unserial(v []byte) (*OutState1, error) {
	out := OutState1{}
	if err := rlp.DecodeBytes(v, &out); err != nil {
		return nil, err
	} else {
		return &out, nil
	}
}
func (self *OutState1) ToWitness() (commitment keys.Uint256, index uint32, path [cpt.DEPTH * 32]byte, anchor keys.Uint256) {
	el := self.Witness.Element()
	commitment = *el.ToUint256()
	path_temp, index_temp := self.Witness.Path()
	index = uint32(index_temp)
	for i, p := range path_temp {
		copy(path[i*32:], p[:])
	}
	anchor_temp := self.Witness.Root()
	copy(anchor[:], anchor_temp[:])
	return
}
