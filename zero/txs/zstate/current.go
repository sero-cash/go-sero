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
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

type Current struct {
	Index int64
	Tree  merkle.Tree
}

func NewCur() (ret Current) {
	ret.Index = -1
	return
}

func (self *Current) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type CurrentGet struct {
	out Current
}

func (self *CurrentGet) Unserial(v []byte) (e error) {
	if v == nil || len(v) == 0 {
		self.out = NewCur()
		return
	} else {
		if err := rlp.DecodeBytes(v, &self.out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
