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
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type In struct {
	Root keys.Uint256
}

type Out struct {
	Addr keys.Uint512
	Pkg  assets.Package
	Memo keys.Uint512
	Z    OutType
}

type OutType int

const (
	TYPE_N = OutType(0)
	TYPE_O = OutType(1)
	TYPE_Z = OutType(2)
)

type T struct {
	Ehash keys.Uint256
	Fee   utils.U256
	Ins   []In
	Outs  []Out
}

func stringToUint256(str string) keys.Uint256{
	var ret keys.Uint256
	b :=[]byte(str)
	if len(b) > len(ret) {
		b = b[len(b)-len(ret):]
	}
	copy(ret[len(ret)-len(b):], b)
	return ret


}

func (self *T) TokenCost() (ret map[keys.Uint256]utils.U256){
	ret = make(map[keys.Uint256]utils.U256)
	seroCy := stringToUint256("sero")
	ret[seroCy] = self.Fee
	if len(self.Outs) > 0 {
		for _, out := range self.Outs {
			if out.Pkg.Tkn != nil {
				if cost,ok:=ret[out.Pkg.Tkn.Currency];ok {
					cost.AddU(&out.Pkg.Tkn.Value)
					ret[out.Pkg.Tkn.Currency] = cost
				}else {
					ret[out.Pkg.Tkn.Currency] = out.Pkg.Tkn.Value
				}
			}
		}
	}
	return
}

func (self *T) Cost() (ret utils.U256) {
	if len(self.Outs) > 0 {
		cost := utils.NewU256(0)
		for _, out := range self.Outs {
			cost.AddU(&out.Pkg.Tkn.Value)
		}
		cost.AddU(&self.Fee)
		return cost
	} else {
		return self.Fee
	}
}
