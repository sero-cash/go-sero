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
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"

	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/zstate/pkgstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate"
	"github.com/sero-cash/go-sero/zero/utils"

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

func CurrentState(tri0 tri.Tri, num uint64) (state *ZState) {
	state = &ZState{}
	state.Tri = tri0
	state.num = num
	state.State = txstate.NewState(tri0, num)
	state.Pkgs = pkgstate.NewPkgState(tri0, num)
	return
}

func NextState(tri0 tri.Tri, num int64) (state *ZState) {
	return CurrentState(tri0, uint64(num+1))
}

func (self *ZState) Copy() *ZState {
	return self
}

func (self *ZState) Update() {
	self.State.Update()
	self.Pkgs.Update()
	return
}

func (self *ZState) PreGenerateRoot(header *types.Header, ch txstate.Chain) {
	self.State.PreGenerateRoot(header, ch)
}

func (self *ZState) RecordBlock(db serodb.Putter, hash *c_type.Uint256) {
	block := localdb.Block{}
	block.Roots = self.State.GetBlockRoots()
	block.Dels = self.State.GetBlockDels()
	block.Pkgs = self.Pkgs.GetPkgHashes()
	localdb.PutBlock(db, self.num, hash, &block)

	for _, hash := range block.Pkgs {
		self.Pkgs.RecordState(db, &hash)
	}

	for _, k := range block.Roots {
		self.State.RecordState(db, &k)
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

func (state *ZState) AddOut_O(out *stx_v0.Out_O, txhash common.Hash) {
	state.State.AddOut_O(out.Clone().ToRef(), txhash.HashToUint256())
}

func (state *ZState) AddOut_P(out *stx_v1.Out_P, txhash common.Hash) {
	state.State.AddOut_P(out.Clone().ToRef(), txhash.HashToUint256())
}

func (state *ZState) AddStx(st *stx.T) (e error) {
	if err := state.State.AddStx(st); err != nil {
		e = err
		return
	} else {
		hash_for_s := st.ToHash_for_sign()
		if st.Desc_Pkg.Create != nil {
			if e = state.Pkgs.Force_add(&st.From, st.Desc_Pkg.Create); e != nil {
				return
			}
		}
		if st.Desc_Pkg.Close != nil {
			if e = state.Pkgs.Force_del(&hash_for_s, st.Desc_Pkg.Close); e != nil {
				return
			}
		}
		if st.Desc_Pkg.Transfer != nil {
			if e = state.Pkgs.Force_transfer(&hash_for_s, st.Desc_Pkg.Transfer); e != nil {
				return
			}
		}
	}
	return
}

func (state *ZState) AddTxOutWithCheck(addr common.Address, asset assets.Asset, txhash common.Hash) (alarm bool) {
	alarm = false
	if state.Num() >= seroparam.VP0() {
		count := state.State.AddTxOut_Log(addr.ToPKr())
		if count > seroparam.MAX_CONTRACT_OUT_COUNT_LENGTH {
			log.Error("[ALARM] ZState AddTxOut Overflow", "MAX_CONTRACT_OUT_COUNT_LENGTH", seroparam.MAX_CONTRACT_OUT_COUNT_LENGTH)
			alarm = true
		}
	}

	state.AddTxOut(addr, asset, txhash)

	return
}

func (state *ZState) AddTxOut(addr common.Address, asset assets.Asset, txhash common.Hash) {
	t := utils.TR_enter("AddTxOut-----")
	need_add := false
	if asset.Tkn != nil {
		if asset.Tkn.Currency != c_type.Empty_Uint256 {
			if asset.Tkn.Value.ToUint256() != c_type.Empty_Uint256 {
				need_add = true
			}
		}
	}
	if asset.Tkt != nil {
		if asset.Tkt.Category != c_type.Empty_Uint256 {
			if asset.Tkt.Value != c_type.Empty_Uint256 {
				need_add = true
			}
		}
	}
	if need_add {
		pkr := addr.ToPKr()
		if c_superzk.IsSzkPKr(pkr) {
			o := stx_v1.Out_P{PKr: *addr.ToPKr(), Asset: asset, Memo: c_type.Uint512{}}
			state.AddOut_P(&o, txhash)
		} else {
			o := stx_v0.Out_O{Addr: *addr.ToPKr(), Asset: asset, Memo: c_type.Uint512{}}
			state.AddOut_O(&o, txhash)
		}
	}
	t.Leave()

	return
}
