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

package zstate

import (
	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
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

func (state *State) AddOut_O(out *stx.Out_O) {
	state.State0.AddOut(out.Clone().ToRef(), nil)
}

func (state *State) AddStx(st *stx.T) (e error) {
	if err := state.State0.AddStx(st); err != nil {
		e = err
		return
	} else {
	}
	return
}

func (state *State) AddTxOut(addr common.Address, pkg assets.Package) {
	o := stx.Out_O{*addr.ToUint512(), pkg, keys.Uint512{}}
	state.AddOut_O(&o)
}
