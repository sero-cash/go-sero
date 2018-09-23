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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

type State1 struct {
	State0 *State0

	G2outs  map[keys.Uint256]*OutState1
	G2wouts []keys.Uint256

	data State1Data

	is_dirty bool
}

func LoadState1(state0 *State0) (state State1) {
	state.State0 = state0
	state.load()
	return
}

func (self *State1) add_out_dirty(k *keys.Uint256, state *OutState1) {
	self.G2outs[*k] = state
	self.is_dirty = true
}
func (self *State1) del_wout_dirty(i uint) {
	self.G2wouts = append(self.G2wouts[:i], self.G2wouts[i+1:]...)
	self.is_dirty = true
}

func (self *State1) append_wout_dirty(k *keys.Uint256) {
	self.G2wouts = append(self.G2wouts, *k)
	self.is_dirty = true
}

func (state *State1) clear_dirty() {
	state.is_dirty = false
}

func (self *State1) load() {
	self.G2outs = make(map[keys.Uint256]*OutState1)
	self.G2wouts = []keys.Uint256{}
	self.clear_dirty()

	if self.State0.num > 0 {
		current_state_name := self.genLoadStateName()
		if current_state_name != "" {
			current_file := zconfig.State1_file(current_state_name)
			if bytes, err := ioutil.ReadFile(current_file); err != nil {
				panic(err)
			} else {
				get := State1DataGet{}
				get.Unserial(bytes)
				self.data = get.out
				self.dataTo()
			}
		}
	} else {
	}
}

func (self *State1) genLoadStateName() (ret string) {
	last_root := self.State0.Block.Tree.RootKey()
	ret = fmt.Sprintf("%010d.%s", self.State0.num-1, hexutil.Encode(last_root[4:16]))
	return
}

func (self *State1) genUpdateStateName() (ret string) {
	last_root := self.State0.Cur.Tree.RootKey()
	ret = fmt.Sprintf("%010d.%s", self.State0.num, hexutil.Encode(last_root[4:16]))
	return
}

func (self *State1) toData() {
	outs := []*OutState1{}
	for _, root := range self.G2wouts {
		if out, ok := self.G2outs[root]; !ok {
			panic("State1 serial but g2outs can not find such root")
		} else {
			outs = append(outs, out)
		}
	}
	self.data.Outs = outs
}

func (self *State1) dataTo() {
	for _, out := range self.data.Outs {
		root := keys.Uint256(out.Witness.Tree.Root())
		self.G2wouts = append(self.G2wouts, root)
		self.G2outs[root] = out
		self.G2outs[out.Nil] = out
	}
}

func (self *State1) Finalize() {
	self.toData()
	self.clear_dirty()
	current_state_name := self.genUpdateStateName()
	if current_state_name != "" {
		current_file := zconfig.State1_file(current_state_name)
		serial := self.data.Serial()
		if err := ioutil.WriteFile(current_file, serial, os.ModePerm); err != nil {
			panic(err)
		} else {
		}
	} else {
		panic("state1 finalize but name is empty")
	}
	zconfig.Remove_State1_dir_files(int(self.State0.num))
	return
}

func (state *State1) GetOut(root *keys.Uint256) (src *OutState1, e error) {
	if out, ok := state.G2outs[*root]; ok {
		if out != nil {
			return out, nil
		} else {
			panic("out from g2outs is nil")
		}
	} else {
		return nil, nil
	}
}

func (self *State1) addOut(tks []keys.Uint512, os *OutState0, os_tree *merkle.Tree) {
	out_hash := os.ToCommitment()
	out_leaf := merkle.Leaf{}
	copy(out_leaf[:], out_hash[:])

	Debug_State1_addout_assert(self, os)

	t := utils.TR_enter(fmt.Sprintf("ADD_OUT..MAX_NUM num=%v", self.State0.num))
	max_num := uint64(0)
	for _, wout := range self.G2wouts {
		if src, err := self.GetOut(&wout); err != nil {
			panic("gen witness wout can not find src")
			return
		} else {
			if max_num < src.Num {
				max_num = src.Num
			}
		}
	}

	t.Renter("RANGE.WOUTS")

	for _, wout := range self.G2wouts {
		if src, err := self.GetOut(&wout); err != nil {
			panic("gen witness wout can not find src")
			return
		} else {
			if src.Num < max_num {
				continue
			}
			if src == nil {
				panic("gen witness can not find wout in G2outs")
			} else {
				if !src.Witness.IsComplete() {
					wit := src.Witness.Clone()
					wit.Append(out_leaf)
					leaf := wit.Root()
					root := leaf.ToUint256()
					if src.Index+1 != os.Index {
						panic("gen witness src.index+1!=os.Index")
					}
					if *root != os_tree.RootKey() {
						panic("gen witness src wit root != out")
					} else {
						src.Witness = wit

					}
				} else {
				}
				src.Index++
				self.add_out_dirty(&wout, src)
			}
		}
	}

	Debug_State1_addout_end_assert(self, os)

	t.Renter("ADD.WOUTS")
	self.addWouts(tks, os, os_tree)
	t.Leave()
	return
}

func (state *State1) addWouts(tks []keys.Uint512, os *OutState0, os_tree *merkle.Tree) {
	for _, tk := range tks {
		if os.IsO() {
			out_o := os.Out_O
			if out_o.Out.Value.Cmp(&utils.U256_0) <= 0 {
				break
			}
			if out_o.Out.Addr == (keys.Uint512{}) {
				break
			}
			info := cpt.Info{}
			info.Currency = out_o.Currency
			info.V = out_o.Out.Value.ToUint256()
			info.Text = out_o.Out.Memo
			if succ, einfo, commitment := cpt.EncodeEInfo(&tk, &out_o.Out.Addr, &info); succ {
				if commitment == *os.ToCommitment() {
					root := os_tree.RootKey()
					state.append_wout_dirty(&root)
					wos := OutState1{}
					wos.Witness.Tree = os_tree.Clone()
					wos.Tk = tk
					wos.Out_O = *os.Out_O
					wos.Index = os.Index
					wos.Desc_Z = &stx.Desc_Z{}
					wos.Desc_Z.Out.Commitment = commitment
					wos.Desc_Z.Out.EInfo = einfo
					wos.Nil = info.Nil
					wos.Z = false
					wos.Num = state.State0.num
					state.add_out_dirty(&root, &wos)
					state.add_out_dirty(&info.Nil, &wos)
					break
				} else {
					panic("add wouts but commitment not match!")
				}
			} else {
			}
		} else {
			desc_z := os.Desc_Z
			if info, e := cpt.DecodeEInfo(
				&tk,
				&desc_z.Out.EInfo,
				&desc_z.S1,
				&desc_z.Out.Commitment,
			); e == nil {
				if bytes.Compare(info.V[:], keys.Empty_Uint256[:]) > 0 {
					root := os_tree.RootKey()
					state.append_wout_dirty(&root)
					wos := OutState1{}
					wos.Out_O.Out.Memo = info.Text
					wos.Out_O.Out.Value = utils.NewU256_ByKey(&info.V)
					//wos.Out_O.Out.R = info.R
					wos.Out_O.Currency = info.Currency
					wos.Tk = tk
					wos.Witness.Tree = os_tree.Clone()
					wos.Desc_Z = os.Desc_Z
					wos.Index = os.Index
					wos.Nil = info.Nil
					wos.Z = true
					wos.Num = state.State0.num
					state.add_out_dirty(&root, &wos)
					state.add_out_dirty(&info.Nil, &wos)
				} else {
				}
				break
			} else {
			}
		}
	}
}

func (state *State1) del(del *keys.Uint256) (e error) {
	if src, err := state.GetOut(del); err != nil {
		e = err
		return
	} else {
		if src == nil {
		} else {
			for i, wout := range state.G2wouts {
				root := src.Witness.Tree.Root()
				if wout == *root.ToUint256() {
					state.del_wout_dirty(uint(i))
					break
				} else {
				}
			}
		}
	}
	return
}

func (state *State1) UpdateWitness(tks []keys.Uint512) {
	trees := state.State0.GenState0Trees()
	for _, del := range state.State0.Block.Dels {
		state.del(&del)
	}
	for i, commitment := range state.State0.Block.Commitments {
		tree := trees.trees[trees.start_index+uint64(i)]
		out := tree.RootKey()
		if os, err := state.State0.GetOut(&out); err != nil {
			panic(err)
		} else {
			if os == nil {
				panic("gen witness out from B2outs can not find in G2outs")
			} else {
			}
			os_commitment := os.ToCommitment()
			if commitment != *os_commitment {
				panic("gen witness out!=os.Root()")
			} else {
			}
			t := utils.TR_enter("UpdateWitness. addOut")
			state.addOut(tks, os, &tree)
			t.Leave()
		}
	}
	return
}
