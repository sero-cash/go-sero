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
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/sero-cash/go-sero/zero/txs/pkg"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

type State struct {
	tri    tri.Tri
	num    uint64
	Cur    Current
	Block  StateBlock
	G2ins  map[keys.Uint256]bool
	G2outs map[keys.Uint256]*OutState
	G2pkgs map[uint64]*pkg.Pkg_Z

	last_out_dirty bool
	g2ins_dirty    map[keys.Uint256]bool
	g2outs_dirty   map[keys.Uint256]bool
	g2pkgs_dirty   map[uint64]bool

	rw *sync.RWMutex
}

func (self *State) Tri() tri.Tri {
	return self.tri
}

func (self *State) Num() uint64 {
	return self.num
}

func NewState0(tri tri.Tri, num uint64) (state State) {
	state = State{tri: tri, num: num}
	state.rw = new(sync.RWMutex)
	state.clear()
	state.load()
	return
}

func (state *State) clear_dirty() {
	state.last_out_dirty = false
	state.g2ins_dirty = make(map[keys.Uint256]bool)
	state.g2outs_dirty = make(map[keys.Uint256]bool)
	state.g2pkgs_dirty = make(map[uint64]bool)
}
func (state *State) clear() {
	state.Cur = NewCur()
	state.G2ins = make(map[keys.Uint256]bool)
	state.G2outs = make(map[keys.Uint256]*OutState)
	state.G2pkgs = make(map[uint64]*pkg.Pkg_Z)
	state.Block = StateBlock{}
	state.clear_dirty()
}

func (state *State) add_pkg_dirty(id uint64, pkg *pkg.Pkg_Z) {
	state.G2pkgs[id] = pkg
	state.g2pkgs_dirty[id] = true
}

func (state *State) del_pkg_dirty(id uint64) {
	state.G2pkgs[id] = nil
	state.g2pkgs_dirty[id] = true
}

func (state *State) append_del_dirty(del *keys.Uint256) {
	if del == nil {
		panic("set_last_out but del is nil")
	}
	state.Block.Dels = append(state.Block.Dels, *del)
	state.last_out_dirty = true
}
func (state *State) append_commitment_dirty(commitment *keys.Uint256) {
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

func (state *State) add_in_dirty(in *keys.Uint256) {
	state.G2ins[*in] = true
	state.g2ins_dirty[*in] = true
}

func (state *State) add_out_dirty(k *keys.Uint256, out *OutState) {
	state.G2outs[*k] = out
	state.g2outs_dirty[*k] = true
}

const LAST_OUTSTATE0_NAME = tri.KEY_NAME("ZState0_Cur")
const BLOCK_NAME = "ZState0_BLOCK"

func (self *State) Name2BKey(name string, num uint64) (ret []byte) {
	key := fmt.Sprintf("%s_%d", name, num)
	ret = []byte(key)
	return
}

func (self *State) load() {
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
func pkgName0(k uint64) (ret []byte) {
	ret = []byte("ZState0_PkgName")
	ret = append(ret, big.NewInt(int64(k)).Bytes()...)
	return
}
func (self *State) Update() {
	self.rw.Lock()
	defer self.rw.Unlock()
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

	g2pkgs_dirty := sort.IntSlice{}
	for k := range self.g2pkgs_dirty {
		g2pkgs_dirty = append(g2pkgs_dirty, int(k))
	}
	sort.Sort(g2pkgs_dirty)

	for _, k := range g2pkgs_dirty {
		v := self.G2pkgs[uint64(k)]
		tri.UpdateObj(self.tri, pkgName0(uint64(k)), v)
	}
	self.clear_dirty()
	return
}

func (self *State) Revert() {
	self.clear()
	self.load()
	return
}

func (state *State) addPkg(pkg *pkg.Pkg_Z) (ret uint64) {
	id := state.Cur.Index
	state.add_pkg_dirty(uint64(id), pkg.Clone().ToRef())
	return uint64(id)
}

func (state *State) DelPkg(id uint64) {
	state.rw.Lock()
	defer state.rw.Unlock()
	state.del_pkg_dirty(id)
}

func (state *State) OpenPkg(id uint64, key *pkg.Key) (ret pkg.Pkg_O, e error) {
	pg := state.GetPkg(id)
	ret, e = pkg.Check(&key.K0, pg)
	if e != nil {
		return
	} else {
		if e = pkg.Verify(&key.K1, &ret, pg); e != nil {
			return
		} else {
			state.DelPkg(id)
			return
		}
	}
}

func (state *State) AddOut(out_o *stx.Out_O, out_z *stx.Out_Z) (root keys.Uint256) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.addOut(out_o, out_z)
}

func (state *State) addOut(out_o *stx.Out_O, out_z *stx.Out_Z) (root keys.Uint256) {
	os := OutState{}
	if out_o != nil {
		o := *out_o
		os.Out_O = &o
	}
	if out_z != nil {
		o := out_z.Clone()
		os.Out_Z = &o
	}

	os.Index = uint64(state.Cur.Index + 1)

	commitment := os.ToRootCM()
	state.append_commitment_dirty(commitment)

	if state.Cur.Index != int64(os.Index) {
		panic("add out but cur.index != current_index")
	}

	if state.Cur.Index < 0 {
		panic("add out but cur.index < 0")
	}

	Debug_State0_addout_assert(state, &os)

	root = state.Cur.Tree.RootKey()
	state.add_out_dirty(&root, &os)
	return
}

func (self *State) HasIn(hash *keys.Uint256) (exists bool) {
	self.rw.Lock()
	defer self.rw.Unlock()
	return self.hasIn(hash)
}
func (self *State) hasIn(hash *keys.Uint256) (exists bool) {
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

func (state *State) addIn(root *keys.Uint256) (e error) {
	if exists := state.hasIn(root); exists {
		e = fmt.Errorf("add in but exists")
		return
	} else {
		state.add_in_dirty(root)
		return
	}
}

func (state *State) AddStx(st *stx.T) (e error) {
	state.rw.Lock()
	defer state.rw.Unlock()
	for _, in := range st.Desc_O.Ins {
		if err := state.addIn(&in.Root); err != nil {
			e = err
			return
		} else {
			state.append_del_dirty(&in.Root)
		}
	}
	//for _, out := range st.Desc_O.Outs {
	//	state.AddOut(out.Clone().ToRef(), nil)
	//}

	for _, in := range st.Desc_Z.Ins {
		if err := state.addIn(&in.Nil); err != nil {
			e = err
			return
		} else {
			state.append_del_dirty(&in.Nil)
			state.append_del_dirty(&in.Trace)
		}
	}

	for _, out := range st.Desc_Z.Outs {
		state.addOut(nil, &out)
	}

	if st.Desc_Pkg.Pack != nil {
		state.addPkg(st.Desc_Pkg.Pack)
	}

	return
}

func (state *State) getPkg(id uint64) (pg *pkg.Pkg_Z) {
	if pg = state.G2pkgs[id]; pg != nil {
		return
	} else {
		get := pkg.Pkg_ZGet{}
		tri.GetObj(state.tri, pkgName0(id), &get)
		pg = get.Out()
		return
	}
}
func (state *State) GetPkg(id uint64) (pg *pkg.Pkg_Z) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.getPkg(id)
}

func (state *State) GetOut(root *keys.Uint256) (src *OutState, e error) {
	state.rw.Lock()
	defer state.rw.Unlock()
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

func (state *State) GenState0Trees() (ret State0Trees) {
	state.rw.RLock()
	defer state.rw.RUnlock()
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
