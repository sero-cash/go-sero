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

package lstate_types

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type OutState struct {
	Root     keys.Uint256
	RootCM   keys.Uint256
	OutIndex uint64
	Num      uint64
	Tk       keys.Uint512
	Out_O    stx.Out_O
	Out_Z    *stx.Out_Z `rlp:"nil"`
	Trace    keys.Uint256
	Z        bool
}

func (self *OutState) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}
func (self *OutState) Unserial(v []byte) (*OutState, error) {
	out := OutState{}
	if err := rlp.DecodeBytes(v, &out); err != nil {
		return nil, err
	} else {
		return &out, nil
	}
}
