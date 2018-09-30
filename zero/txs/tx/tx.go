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

package tx

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
)

type In struct {
	Root keys.Uint256
}

type Out struct {
	Addr  keys.Uint512
	Value utils.U256
	Memo  keys.Uint512
	Z     OutType
}

type OutType int

const (
	TYPE_N = OutType(0)
	TYPE_O = OutType(1)
	TYPE_Z = OutType(2)
)

type CTx struct {
	Currency keys.Uint256
	Fee      utils.U256
	Ins      []In
	Outs     []Out
}

func (self *CTx) Cost() (ret utils.U256) {
	if len(self.Outs) > 0 {
		cost := utils.NewU256(0)
		for _, out := range self.Outs {
			cost.AddU(&out.Value)
		}
		cost.AddU(&self.Fee)
		return cost
	} else {
		return self.Fee
	}
}

type T struct {
	Ehash keys.Uint256
	CTxs  []CTx
}
