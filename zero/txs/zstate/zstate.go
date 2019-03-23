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
	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/zstate/pkgstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type ZState struct {
	Tri   tri.Tri
	num   uint64
	State txstate.State
	Pkgs  pkgstate.PkgState
}

func (self *ZState) Num() uint64 {
	return self.num
}

func NewState(tri0 tri.Tri, num uint64) (state *ZState) {
	state = &ZState{}
	state.Tri = tri0
	state.num = num
	state.State = txstate.NewState(tri0, num)
	state.Pkgs = pkgstate.NewPkgState(tri0, num)
	return
}

func (self *ZState) Copy() *ZState {
	return nil
}

func (self *ZState) Update() {
	self.State.Update()
	self.Pkgs.Update()
	return
}

func (self *ZState) RecordBlock(db serodb.Putter, hash *keys.Uint256) {
	block := localdb.Block{}
	block.Pkgs = self.Pkgs.GetBlockDetails()
	block.Roots = self.State.Block.Roots
	block.Dels = self.State.Block.Dels
	localdb.PutBlock(db, self.num, hash, &block)
	for _, k := range block.Roots {
		os := self.State.G2outs[k]
		localdb.PutOut(db, &k, os)
	}
}

func (self *ZState) Snapshot(revid int) {
	t := utils.TR_enter("Snapshot")
	self.State.Snapshot(revid)
	self.Pkgs.Snapshot(revid)
	t.Leave()
}

func (self *ZState) Revert(revid int) {
	self.State.Revert(revid)
	self.Pkgs.Revert(revid)
	return
}

func (state *ZState) AddOut_O(out *stx.Out_O) {
	state.State.AddOut(out.Clone().ToRef(), nil)
}

func (state *ZState) AddStx(st *stx.T) (e error) {
	if err := state.State.AddStx(st); err != nil {
		e = err
		return
	} else {
		if st.Desc_Pkg.Create != nil {
			state.Pkgs.Force_add(&st.From, st.Desc_Pkg.Create)
		}
		if st.Desc_Pkg.Close != nil {
			state.Pkgs.Force_del(&st.Desc_Pkg.Close.Id)
		}
		if st.Desc_Pkg.Transfer != nil {
			state.Pkgs.Force_transfer(&st.Desc_Pkg.Transfer.Id, &st.Desc_Pkg.Transfer.PKr)
		}
	}
	return
}

func (state *ZState) AddTxOut(addr common.Address, asset assets.Asset) {
	t := utils.TR_enter("AddTxOut-----")
	need_add := false
	if asset.Tkn != nil {
		if asset.Tkn.Currency != keys.Empty_Uint256 {
			if asset.Tkn.Value.ToUint256() != keys.Empty_Uint256 {
				need_add = true
			}
		}
	}
	if asset.Tkt != nil {
		if asset.Tkt.Category != keys.Empty_Uint256 {
			if asset.Tkt.Value != keys.Empty_Uint256 {
				need_add = true
			}
		}
	}
	if need_add {
		o := stx.Out_O{*addr.ToPKr(), asset, keys.Uint512{}}
		state.AddOut_O(&o)
	}
	t.Leave()
}
