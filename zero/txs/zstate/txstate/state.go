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

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/common/hexutil"
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

	data      data.Data
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
	state.data = data.NewData(num)
	state.data.Clear()
	state.load()
	return
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
	self.snapshots.Push(revid, &self.data)
}

func (self *State) Revert(revid int) {
	self.data.Clear()
	self.data = *self.snapshots.Revert(revid).(*data.Data)
}

func (state *State) AddOut(out_o *stx.Out_O, out_z *stx.Out_Z) (root keys.Uint256) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.addOut(out_o, out_z)
}

func (state *State) addOut(out_o *stx.Out_O, out_z *stx.Out_Z) (root keys.Uint256) {
	os := localdb.OutState{}
	if out_o != nil {
		o := *out_o
		os.Out_O = &o
	}
	if out_z != nil {
		o := out_z.Clone()
		os.Out_Z = &o
	}

	os.Index = uint64(state.data.Cur.Index + 1)

	commitment := os.ToRootCM()

	root = state.MTree.AppendLeaf(*commitment)

	state.data.AddOut(&root, &os)
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
		if state.data.HasIn(state.tri, &in.Root) {
			e = errors.New("desc_o.root already be used !")
			return
		} else {
			state.data.AddNil(&in.Root)
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
	for _, out := range st.Desc_Z.Outs {
		state.addOut(nil, &out)
	}

	t.Leave()

	return
}

func (state *State) GetOut(root *keys.Uint256) (src *localdb.OutState, e error) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.data.GetOut(state.tri, root), nil
}

func (self *State) GetBlockRoots() (roots []keys.Uint256) {
	return self.data.Block.Roots
}

func (self *State) GetBlockDels() (dels []keys.Uint256) {
	return self.data.Block.Dels
}

type Chain interface {
	GetBlock(hash common.Hash, number uint64) *types.Block
}

func (self *State) PreGenerateRoot(header *types.Header, ch Chain) {
	return
	hash := header.ParentHash
	number := header.Number.Uint64() - 1
	nils := make(map[keys.Uint256]bool)
	for {
		b := ch.GetBlock(hash, number)
		for _, tx := range b.Transactions() {
			for _, in := range tx.Stxt().Desc_Z.Ins {
				if _, ok := nils[in.Nil]; ok {
					fmt.Printf("block=%v,hash=%v,tx:=%v\n", number, hexutil.Encode(hash[:]), tx.Hash())
				} else {
					nils[in.Nil] = true
				}
			}
			for _, in := range tx.Stxt().Desc_O.Ins {
				if _, ok := nils[in.Nil]; ok {
					fmt.Printf("block=%v,hash=%v,tx:=%v\n", number, hexutil.Encode(hash[:]), tx.Hash())
				} else {
					nils[in.Nil] = true
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

type State0Trees struct {
	Trees       map[uint64]merkle.Tree
	Roots       []keys.Uint256
	Start_index uint64
}
