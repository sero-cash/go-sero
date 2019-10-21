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

package txstate

import (
	"fmt"
	"sort"
	"sync"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"

	"github.com/sero-cash/go-sero/zero/txs/zstate/merkle"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data_v1"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/core/types"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
)

type State struct {
	tri       tri.Tri
	rw        *sync.RWMutex
	num       uint64
	CzeroTree merkle.MerkleTree
	SzkTree   merkle.MerkleTree

	data data.IData

	logs      []data.Log
	revisions []data.Revision
}

func (self *State) Num() uint64 {
	return self.num
}

func (self *State) Tri() tri.Tri {
	return self.tri
}

func NewState(tri tri.Tri, num uint64) (state State) {
	state = State{tri: tri, num: num}
	state.rw = new(sync.RWMutex)
	state.CzeroTree = CzeroMerkleParam.NewMerkleTree(tri)
	state.SzkTree = SzkMerkleParam.NewMerkleTree(tri)
	if state.num >= seroparam.SIP2() {
		state.data = data_v1.NewData(num)
	} else {
		state.data = data.NewData(num)
	}
	state.data.Clear()
	state.load()
	return
}

func (self *State) RecordState(putter serodb.Putter, root *c_type.Uint256) {
	self.data.RecordState(putter, root)
}

func (self *State) load() {
	self.data.LoadState(self.tri)
}

func (self *State) Update() {
	self.rw.Lock()
	defer self.rw.Unlock()
	self.data.SaveState(self.tri)
	return
}

func (state *State) Snapshot(revid int) {
	state.revisions = append(state.revisions, data.Revision{revid, len(state.logs)})
}

func (state *State) Revert(revid int) {

	idx := sort.Search(len(state.revisions), func(i int) bool {
		return state.revisions[i].Id >= revid
	})
	if idx == len(state.revisions) || state.revisions[idx].Id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}

	index := state.revisions[idx].JournalIndex

	state.revisions = state.revisions[:idx]
	state.logs = state.logs[:index]

	state.data.Clear()
	for _, log := range state.logs {
		log.Op(state.data)
	}
}

func (self *State) addOut_Log(root *c_type.Uint256, out *localdb.OutState, txhash *c_type.Uint256) {
	clone := out.Clone()
	if txhash != nil {
		self.logs = append(self.logs, data.AddOutLog{root.NewRef(), &clone, txhash.NewRef()})
	} else {
		self.logs = append(self.logs, data.AddOutLog{root.NewRef(), &clone, nil})
	}

	self.data.AddOut(root, out, txhash)
	return
}
func (self *State) addNil_Log(in *c_type.Uint256) {
	self.logs = append(self.logs, data.AddNilLog{in.NewRef()})
	self.data.AddNil(in)
}
func (self *State) addDel_Log(in *c_type.Uint256) {
	self.logs = append(self.logs, data.AddDelLog{in.NewRef()})
	self.data.AddDel(in)
}

func (self *State) AddTxOut_Log(pkr *c_type.PKr) int {
	self.logs = append(self.logs, data.AddTxOutLog{pkr.NewRef()})
	return self.data.AddTxOut(pkr)
}

func (state *State) AddOut_O(out_o *stx_v0.Out_O, txhash *c_type.Uint256) (root c_type.Uint256) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.addOut_O(out_o, txhash)
}

func (state *State) AddOut_P(out_p *stx_v1.Out_P, txhash *c_type.Uint256) (root c_type.Uint256) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.addOut_P(out_p, txhash)
}

func (state *State) insertOS(os *localdb.OutState, txhash *c_type.Uint256) (root c_type.Uint256) {
	if os.Out_O != nil || os.Out_Z != nil {
		os.Index = state.CzeroTree.GetLeafSize()
		os.GenRootCM()
		root = state.CzeroTree.AppendLeaf(*os.RootCM)
		state.addOut_Log(&root, os, txhash)
	} else {
		os.Index = state.SzkTree.GetLeafSize()
		os.GenRootCM()
		root = state.SzkTree.AppendLeaf(*os.RootCM)
		state.addOut_Log(&root, os, txhash)
	}
	return
}

func (state *State) addOut_O(out_o *stx_v0.Out_O, txhash *c_type.Uint256) (root c_type.Uint256) {
	os := localdb.OutState{}
	if out_o != nil {
		o := *out_o
		os.Out_O = &o
	}
	return state.insertOS(&os, txhash)
}

func (state *State) addOut_Z(out_z *stx_v0.Out_Z, txhash *c_type.Uint256) (root c_type.Uint256) {
	os := localdb.OutState{}
	if out_z != nil {
		o := out_z.Clone()
		os.Out_Z = &o
	}
	return state.insertOS(&os, txhash)
}

func (state *State) addOut_C(out_c *stx_v1.Out_C, txhash *c_type.Uint256) (root c_type.Uint256) {
	os := localdb.OutState{}
	if out_c != nil {
		o := out_c.Clone()
		os.Out_C = &o
	}
	return state.insertOS(&os, txhash)
}

func (state *State) addOut_P(out_p *stx_v1.Out_P, txhash *c_type.Uint256) (root c_type.Uint256) {
	os := localdb.OutState{}
	if out_p != nil {
		o := out_p.Clone()
		os.Out_P = &o
	}
	return state.insertOS(&os, txhash)
}

func (self *State) HasIn(hash *c_type.Uint256) (exists bool) {
	self.rw.Lock()
	defer self.rw.Unlock()
	return self.data.HasIn(self.tri, hash)
}

func (state *State) addTx0(tx *stx_v0.Tx, txhash *c_type.Uint256) (e error) {
	t := utils.TR_enter("AddStx---ins")
	for _, in := range tx.Desc_O.Ins {
		if state.num >= seroparam.SIP2() {
			if state.data.HasIn(state.tri, &in.Nil) {
				e = errors.New("desc_o.in.nil already be used !")
				return
			} else {
				state.addNil_Log(&in.Nil)
				state.addDel_Log(&in.Root)
			}
		} else {
			if state.data.HasIn(state.tri, &in.Root) {
				e = errors.New("desc_o.in.root already be used !")
				return
			} else {
				state.addNil_Log(&in.Root)
			}
		}
	}

	t.Renter("AddStx---z_ins")
	for _, in := range tx.Desc_Z.Ins {
		if state.data.HasIn(state.tri, &in.Nil) {
			e = errors.New("desc_o.nil already be used !")
			return
		} else {
			state.addNil_Log(&in.Nil)
			state.addDel_Log(&in.Trace)
		}
	}

	t.Renter("AddStx---z_outs")
	for _, out := range tx.Desc_Z.Outs {
		state.addOut_Z(&out, txhash)
	}

	t.Leave()
	return
}

func (state *State) addTx1(tx *stx_v1.Tx, txhash *c_type.Uint256) (e error) {
	for _, in := range tx.Ins_P0 {
		if !state.data.HasIn(state.tri, &in.Nil) {
			if !state.data.HasIn(state.tri, &in.Root) {
				state.addNil_Log(&in.Nil)
				state.addNil_Log(&in.Root)
				state.addDel_Log(&in.Trace)
			} else {
				e = errors.New("tx1.in_p0.root already be used !")
				return
			}
		} else {
			e = errors.New("tx1.in_p0.nil already be used !")
			return
		}
	}
	for _, in := range tx.Ins_P {
		if !state.data.HasIn(state.tri, &in.Nil) {
			if !state.data.HasIn(state.tri, &in.Root) {
				state.addNil_Log(&in.Nil)
				state.addNil_Log(&in.Root)
			} else {
				e = errors.New("tx1.in_p.root already be used !")
				return
			}
		} else {
			e = errors.New("tx1.in_p.nil already be used !")
			return
		}
	}
	for _, in := range tx.Ins_C {
		if !state.data.HasIn(state.tri, &in.Nil) {
			state.addNil_Log(&in.Nil)
		} else {
			e = errors.New("tx1.in_c.nil already be used !")
			return
		}
	}
	for _, out := range tx.Outs_C {
		state.addOut_C(&out, txhash)
	}
	for _, out := range tx.Outs_P {
		if c_superzk.IsSzkPKr(&out.PKr) {
			state.addOut_P(&out, txhash)
		} else {
			state.addOut_O(
				&stx_v0.Out_O{
					Addr:  out.PKr,
					Asset: out.Asset,
					Memo:  out.Memo,
				},
				txhash,
			)
		}
	}
	return
}

func (state *State) AddStx(st *stx.T) (e error) {
	state.rw.Lock()
	defer state.rw.Unlock()
	txhash := st.ToHash()
	if st.Tx0() != nil && st.Tx1.Count() > 0 {
		return fmt.Errorf("add stx, tx0 & tx1 only one can has value ")
	}

	if st.Tx0() != nil {
		return state.addTx0(st.Tx0(), &txhash)
	}

	if st.Tx1.Count() > 0 {
		return state.addTx1(&st.Tx1, &txhash)
	}

	return
}

func (state *State) GetOut(root *c_type.Uint256) (src *localdb.OutState) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.data.GetOut(state.tri, root)
}

func (state *State) FindAnchorInSzk(root *c_type.Uint256) bool {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.data.HashRoot(state.tri, root)
}

func (self *State) GetBlockRoots() (roots []c_type.Uint256) {
	return self.data.GetRoots()
}

func (self *State) GetBlockDels() (dels []c_type.Uint256) {
	return self.data.GetDels()
}

type Chain interface {
	GetBlock(hash common.Hash, number uint64) *types.Block
}

func AnalyzeNils(header *types.Header, ch Chain) {
	hash := header.ParentHash
	number := header.Number.Uint64() - 1
	m := make(map[c_type.Uint256]int)
	for {
		b := ch.GetBlock(hash, number)
		for _, tx := range b.Transactions() {
			if tx.Stxt().Tx0() != nil {
				for _, in := range tx.Stxt().Tx0().Desc_O.Ins {
					if v, ok := m[in.Root]; ok {
						fmt.Printf("num=%v,block=%v,tx=%v,current=%v,hit=%v\n", number, hexutil.Encode(hash[:]), hexutil.Encode(in.ToHash().NewRef()[:]), 1, v)
					} else {
						m[in.Root] = 1
					}
				}
				for _, in := range tx.Stxt().Tx0().Desc_O.Ins {
					if v, ok := m[in.Nil]; ok {
						fmt.Printf("num=%v,block=%v,tx=%v,current=%v,hit=%v\n", number, hexutil.Encode(hash[:]), hexutil.Encode(in.ToHash().NewRef()[:]), 2, v)
					} else {
						m[in.Nil] = 2
					}
				}
				for _, in := range tx.Stxt().Tx0().Desc_Z.Ins {
					if v, ok := m[in.Nil]; ok {
						fmt.Printf("num=%v,block=%v,tx=%v,current=%v,hit=%v\n", number, hexutil.Encode(hash[:]), hexutil.Encode(in.ToHash().NewRef()[:]), 3, v)
					} else {
						m[in.Nil] = 3
					}
				}
			}
		}
		if number == 0 {
			break
		}
		hash = b.ParentHash()
		number = number - 1
	}
}

func (self *State) PreGenerateRoot(header *types.Header, ch Chain) {
	if header.Number.Uint64() == (seroparam.SIP2()) {
		hash := header.ParentHash
		number := header.Number.Uint64() - 1
		size := number
		progress := utils.NewProgress("PRE GEN ROOTS: ", size)
		count := 0
		for {
			b := ch.GetBlock(hash, number)
			for _, tx := range b.Transactions() {
				if tx.Stxt().Tx0() != nil {
					for _, in := range tx.Stxt().Tx0().Desc_O.Ins {
						self.addNil_Log(&in.Nil)
						count++
					}
				}
			}
			progress.Tick(size-number, "count", count)
			if number == 0 {
				break
			}
			hash = b.ParentHash()
			number = number - 1
		}
	}
}
