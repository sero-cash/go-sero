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
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
)

type State struct {
	State0
}

func NewState(tri0 tri.Tri, num uint64) (state *State) {
	state = &State{}
	state.State0 = NewState0(tri0, num)
	return
}

func (self *State) Copy() *State {
	return nil
}

func (self *State) Update() {
	self.State0.Update()
	return
}

func (self *State) Revert() {
	self.State0.Revert()
	return
}

func (state *State) AddOut_O(out *stx.Out_O, currency *keys.Uint256) {
	out0 := Out0{
		*currency,
		*out,
	}
	state.State0.AddOut(&out0, nil)
}

func (state *State) AddStx(st *stx.T) (e error) {
	if err := state.State0.AddStx(st); err != nil {
		e = err
		return
	} else {
	}
	return
}

func (state *State) AddTxOut(addr common.Address, value *big.Int, currency *keys.Uint256) {
	o := stx.Out_O{*addr.ToUint512(), utils.U256(*value), keys.Uint512{}}
	state.AddOut_O(&o, currency)
}
