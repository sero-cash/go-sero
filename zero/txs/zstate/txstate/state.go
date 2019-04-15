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
	"sync"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data_v1"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/core/types"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

type State struct {
	tri   tri.Tri
	rw    *sync.RWMutex
	num   uint64
	MTree MerkleTree

	data      data.IData
	snapshots utils.Snapshots
}

func (self *State) Tri() tri.Tri {
	return self.tri
}

func (self *State) Num() uint64 {
	return self.num
}

func NewState(tri tri.Tri, num uint64) (state State) {
	state = State{tri: tri, num: num}
	state.rw = new(sync.RWMutex)
	state.MTree = NewMerkleTree(tri)
	if num >= cpt.SIP2 {
		state.data = data_v1.NewData(num)
	} else {
		state.data = data.NewData(num)
	}
	state.data.Clear()
	state.load()
	return
}

func (self *State) RecordState(putter serodb.Putter, root *keys.Uint256) {
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

func (self *State) Snapshot(revid int) {
	self.snapshots.Push(revid, self.data)
}

func (self *State) Revert(revid int) {
	self.data.Clear()
	self.data = self.snapshots.Revert(revid).(data.IData)
}

func (state *State) AddOut(out_o *stx.Out_O, out_z *stx.Out_Z, txhash *keys.Uint256) (root keys.Uint256) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.addOut(out_o, out_z, txhash)
}

func (state *State) addOut(out_o *stx.Out_O, out_z *stx.Out_Z, txhash *keys.Uint256) (root keys.Uint256) {
	os := localdb.OutState{}
	if out_o != nil {
		o := *out_o
		os.Out_O = &o
	}
	if out_z != nil {
		o := out_z.Clone()
		os.Out_Z = &o
	}

	//index := state.MTree.GetCurrentIndex()

	//os.Index = uint64(state.data.GetIndex() + 1)
	os.Index = state.MTree.GetLeafSize()

	commitment := os.ToRootCM()

	root = state.MTree.AppendLeaf(*commitment)

	state.data.AddOut(&root, &os, txhash)
	return
}

func (self *State) HasIn(hash *keys.Uint256) (exists bool) {
	self.rw.Lock()
	defer self.rw.Unlock()
	return self.data.HasIn(self.tri, hash)
}

func (state *State) AddStx(st *stx.T) (e error) {
	state.rw.Lock()
	defer state.rw.Unlock()
	t := utils.TR_enter("AddStx---ins")
	for _, in := range st.Desc_O.Ins {
		if state.num >= cpt.SIP2 {
			if state.data.HasIn(state.tri, &in.Nil) {
				e = errors.New("desc_o.in.nil already be used !")
				return
			} else {
				state.data.AddNil(&in.Nil)
				state.data.AddDel(&in.Root)
			}
		} else {
			if state.data.HasIn(state.tri, &in.Root) {
				e = errors.New("desc_o.in.root already be used !")
				return
			} else {
				state.data.AddNil(&in.Root)
			}
		}
	}

	t.Renter("AddStx---z_ins")
	for _, in := range st.Desc_Z.Ins {
		if state.data.HasIn(state.tri, &in.Nil) {
			e = errors.New("desc_o.nil already be used !")
			return
		} else {
			state.data.AddNil(&in.Nil)
			state.data.AddDel(&in.Trace)
		}
	}

	t.Renter("AddStx---z_outs")
	txhash := st.ToHash()
	for _, out := range st.Desc_Z.Outs {
		state.addOut(nil, &out, &txhash)
	}

	t.Leave()

	return
}

func (state *State) GetOut(root *keys.Uint256) (src *localdb.OutState) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.data.GetOut(state.tri, root)
}

func (self *State) GetBlockRoots() (roots []keys.Uint256) {
	return self.data.GetRoots()
}

func (self *State) GetBlockDels() (dels []keys.Uint256) {
	return self.data.GetDels()
}

type Chain interface {
	GetBlock(hash common.Hash, number uint64) *types.Block
}

func AnalyzeNils(header *types.Header, ch Chain) {
	hash := header.ParentHash
	number := header.Number.Uint64() - 1
	m := make(map[keys.Uint256]int)
	for {
		b := ch.GetBlock(hash, number)
		for _, tx := range b.Transactions() {
			for _, in := range tx.Stxt().Desc_O.Ins {
				if v, ok := m[in.Root]; ok {
					fmt.Printf("num=%v,block=%v,tx=%v,current=%v,hit=%v\n", number, hexutil.Encode(hash[:]), hexutil.Encode(in.ToHash().NewRef()[:]), 1, v)
				} else {
					m[in.Root] = 1
				}
			}
			for _, in := range tx.Stxt().Desc_O.Ins {
				if v, ok := m[in.Nil]; ok {
					fmt.Printf("num=%v,block=%v,tx=%v,current=%v,hit=%v\n", number, hexutil.Encode(hash[:]), hexutil.Encode(in.ToHash().NewRef()[:]), 2, v)
				} else {
					m[in.Nil] = 2
				}
			}
			for _, in := range tx.Stxt().Desc_Z.Ins {
				if v, ok := m[in.Nil]; ok {
					fmt.Printf("num=%v,block=%v,tx=%v,current=%v,hit=%v\n", number, hexutil.Encode(hash[:]), hexutil.Encode(in.ToHash().NewRef()[:]), 3, v)
				} else {
					m[in.Nil] = 3
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
	if header.Number.Uint64() == (cpt.SIP2) {
		hash := header.ParentHash
		number := header.Number.Uint64() - 1
		size := number
		progress := utils.NewProgress("PRE GEN ROOTS: ", size)
		count := 0
		for {
			b := ch.GetBlock(hash, number)
			for _, tx := range b.Transactions() {
				for _, in := range tx.Stxt().Desc_O.Ins {
					self.data.AddNil(&in.Nil)
					count++
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

type State0Trees struct {
	Trees       map[uint64]merkle.Tree
	Roots       []keys.Uint256
	Start_index uint64
}
