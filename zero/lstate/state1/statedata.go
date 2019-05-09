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

package state1

import (
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/lstate"
)

type StateData struct {
	Outs      []*lstate.OutState
	Pkgs_from []*lstate.Pkg
	Pkgs_to   []*lstate.Pkg
	MaxNum    uint64
}

func (self *StateData) Serial() (ret []byte) {
	if self != nil {
		if bytes, err := rlp.EncodeToBytes(self); err != nil {
			panic(err)
		} else {
			ret = bytes
			return
		}
	} else {
		return
	}
}

type StateDataGet struct {
	out StateData
}

func (self *StateDataGet) Unserial(v []byte) {
	if len(v) == 0 {
		return
	} else {
		if err := rlp.DecodeBytes(v, &self.out); err != nil {
			panic(err)
			return
		} else {
			return
		}
	}
}
