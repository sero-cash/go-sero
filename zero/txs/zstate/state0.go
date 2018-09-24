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
	"fmt"
	"sort"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

type State0 struct {
	tri    tri.Tri
	num    uint64
	Cur    Current
	Block  State0Block
	G2ins  map[keys.Uint256]bool
	G2outs map[keys.Uint256]*OutState0

	last_out_dirty bool
	g2ins_dirty    map[keys.Uint256]bool
	g2outs_dirty   map[keys.Uint256]bool
}

func (self *State0) Tri() tri.Tri {
	return self.tri
}

func (self *State0) Num() uint64 {
	return self.num
}

func NewState0(tri tri.Tri, num uint64) (state State0) {
	state = State0{tri: tri, num: num}
	state.clear()
	state.load()
	return
}

func (state *State0) clear_dirty() {
	state.last_out_dirty = false
	state.g2ins_dirty = make(map[keys.Uint256]bool)
	state.g2outs_dirty = make(map[keys.Uint256]bool)
}
func (state *State0) clear() {
	state.Cur = NewCur()
	state.G2ins = make(map[keys.Uint256]bool)
	state.G2outs = make(map[keys.Uint256]*OutState0)
	state.Block = State0Block{}
	state.clear_dirty()
}

func (state *State0) append_del_dirty(del *keys.Uint256) {
	if del == nil {
		panic("set_last_out but del is nil")
	}
	state.Block.Dels = append(state.Block.Dels, *del)
	state.last_out_dirty = true
}
func (state *State0) append_commitment_dirty(commitment *keys.Uint256) {
	if commitment == nil {
		panic("set_last_out but out is nil")
	}

	state.Cur.Index = state.Cur.Index + int64(1)

	if !state.Cur.Tree.IsComplete() {
		state.Cur.Tree = state.Cur.Tree.Clone()
	} else {
		state.Cur.Tree = merkle.Tree{}
	}
	state.Cur.Tree.Append(merkle.Leaf(*commitment))
	state.Block.Commitments = append(state.Block.Commitments, *commitment)
	state.last_out_dirty = true
}

func (state *State0) add_in_dirty(in *keys.Uint256) {
	state.G2ins[*in] = true
	state.g2ins_dirty[*in] = true
}

func (state *State0) add_out_dirty(k *keys.Uint256, out *OutState0) {
	state.G2outs[*k] = out
	state.g2outs_dirty[*k] = true
}

const LAST_OUTSTATE0_NAME = tri.KEY_NAME("ZState0_Cur")
const BLOCK_NAME = "ZState0_BLOCK"

func (self *State0) Name2BKey(name string, num uint64) (ret []byte) {
	key := fmt.Sprintf("%s_%d", name, num)
	ret = []byte(key)
	return
}

func (self *State0) load() {
	get := CurrentGet{}
	tri.GetObj(
		self.tri,
		LAST_OUTSTATE0_NAME.Bytes(),
		&get,
	)
	self.Cur = get.out

	blockget := State0BlockGet{}
	tri.GetObj(
		self.tri,
		self.Name2BKey(BLOCK_NAME, self.num),
		&blockget,
	)
	self.Block = blockget.out
	if self.Block.Tree == nil {
		tree := self.Cur.Tree.Clone()
		self.Block.Tree = &tree
		self.last_out_dirty = true
	}
}

func inName(k *keys.Uint256) (ret []byte) {
	ret = []byte("ZState0_InName")
	ret = append(ret, k[:]...)
	return
}
func outName0(k *keys.Uint256) (ret []byte) {
	ret = []byte("ZState0_OutName")
	ret = append(ret, k[:]...)
	return
}
func (self *State0) Update() {
	if self.last_out_dirty {
		tri.UpdateObj(self.tri, LAST_OUTSTATE0_NAME.Bytes(), &self.Cur)
		tri.UpdateObj(
			self.tri,
			self.Name2BKey(BLOCK_NAME, self.num),
			&self.Block,
		)

		blockget := State0BlockGet{}
		tri.GetObj(
			self.tri,
			self.Name2BKey(BLOCK_NAME, self.num),
			&blockget,
		)
		i := 0
		i++
	}

	g2ins_dirty := utils.Uint256s{}
	for k := range self.g2ins_dirty {
		g2ins_dirty = append(g2ins_dirty, k)
	}
	sort.Sort(g2ins_dirty)

	for _, k := range g2ins_dirty {
		v := []byte{1}
		if err := self.tri.TryUpdate(inName(&k), v); err != nil {
			panic(err)
			return
		}
	}

	g2outs_dirty := utils.Uint256s{}
	for k := range self.g2outs_dirty {
		g2outs_dirty = append(g2outs_dirty, k)
	}
	sort.Sort(g2outs_dirty)

	for _, k := range g2outs_dirty {
		if v := self.G2outs[k]; v != nil {
			tri.UpdateObj(self.tri, outName0(&k), v)
		} else {
			panic("state0 update g2outs can not find dirty out")
		}
	}
	self.clear_dirty()
	return
}

func (self *State0) Revert() {
	self.clear()
	self.load()
	return
}

func (state *State0) AddOut(out_o *Out0, desc_z *stx.Desc_Z) (root keys.Uint256) {
	os := OutState0{}
	if out_o != nil {
		o := *out_o
		os.Out_O = &o
	}
	if desc_z != nil {
		o := *desc_z
		os.Desc_Z = &o
	}
	commitment := os.ToCommitment()
	state.append_commitment_dirty(commitment)

	if state.Cur.Index < 0 {
		panic("add out but cur.index < 0")
	}
	root = state.Cur.Tree.RootKey()
	os.Index = uint64(state.Cur.Index)

	Debug_State0_addout_assert(state, &os)

	state.add_out_dirty(&root, &os)
	return
}

func (self *State0) HasIn(hash *keys.Uint256) (exists bool) {
	if v, ok := self.G2ins[*hash]; ok {
		exists = v
		return
	} else {
		if v, err := self.tri.TryGet(inName(hash)); err != nil {
			panic(err)
			return
		} else {
			if v != nil && v[0] == 1 {
				exists = true
			} else {
				exists = false
			}
			self.G2ins[*hash] = exists
		}
	}
	return
}

func (state *State0) addIn(root *keys.Uint256) (e error) {
	if exists := state.HasIn(root); exists {
		e = fmt.Errorf("add in but exists")
		return
	} else {
		state.add_in_dirty(root)
		return
	}
}

func (state *State0) AddStx(st *stx.T) (e error) {
	for _, desc_o := range st.Desc_Os {
		for _, in := range desc_o.Ins {
			if err := state.addIn(&in.Root); err != nil {
				e = err
				return
			} else {
				state.append_del_dirty(&in.Root)
			}
		}
		for _, out := range desc_o.Outs {
			out0 := Out0{
				desc_o.Currency,
				out,
			}
			state.AddOut(&out0, nil)
		}
	}
	for _, desc_z := range st.Desc_Zs {
		if err := state.addIn(&desc_z.In.Trace); err != nil {
			e = err
			return
		} else {
			state.append_del_dirty(&desc_z.In.Trace)
		}
		state.AddOut(nil, &desc_z)
	}
	return
}

func (state *State0) GetOut(root *keys.Uint256) (src *OutState0, e error) {
	if out := state.G2outs[*root]; out != nil {
		return out, nil
	} else {
		get := OutState0Get{}
		tri.GetObj(state.tri, outName0(root), &get)
		if get.out != nil {
			state.G2outs[*root] = get.out
			return get.out, nil
		} else {
			return nil, nil
		}
	}
}

type State0Trees struct {
	Trees       map[uint64]merkle.Tree
	Roots       []keys.Uint256
	Start_index uint64
}

func (state *State0) GenState0Trees() (ret State0Trees) {
	if state.Cur.Index >= 0 {
		ret.Trees = make(map[uint64]merkle.Tree)
		tree := state.Block.Tree.Clone()
		ret.Start_index = uint64(state.Cur.Index - int64(len(state.Block.Commitments)) + 1)
		for i, commitment := range state.Block.Commitments {
			tree.Append(merkle.Leaf(commitment))
			ret.Trees[ret.Start_index+uint64(i)] = tree.Clone()
			ret.Roots = append(ret.Roots, tree.RootKey())
		}
	}
	return

}
