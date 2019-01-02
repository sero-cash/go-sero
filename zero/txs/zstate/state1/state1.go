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

package state1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-sero/zero/txs/stx"

	"github.com/sero-cash/go-sero/zero/witness"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

type State1 struct {
	State0 *zstate.State0

	mu      sync.RWMutex
	G2outs  map[keys.Uint256]*OutState1
	G2wouts []keys.Uint256

	data State1Data

	is_dirty bool
}

func LoadState1(state0 *zstate.State0, loadName string) (state State1) {
	state.State0 = state0
	state.load(loadName)
	return
}

func (self *State1) add_out_dirty(k *keys.Uint256, state *OutState1) {
	self.mu.Lock()
	defer self.mu.Unlock()
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

func (self *State1) load(loadName string) {
	defer func() {
		if r := recover(); r != nil {
			if loadName != "" {
				current_file := zconfig.State1_file(loadName)
				os.Remove(current_file)
				panic(r)
			}
		}
	}()

	self.G2outs = make(map[keys.Uint256]*OutState1)
	self.G2wouts = []keys.Uint256{}
	self.clear_dirty()

	if loadName != "" {
		current_file := zconfig.State1_file(loadName)
		if bytes, err := ioutil.ReadFile(current_file); err != nil {
			panic(err)
		} else {
			get := State1DataGet{}
			get.Unserial(bytes)
			self.data = get.out
			self.dataTo()
		}
	} else {
	}
}

func (self *State1) toData() {
	self.mu.RLock()
	defer self.mu.RUnlock()
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
		root := keys.Uint256(out.Pg.Root)
		self.G2wouts = append(self.G2wouts, root)
		self.G2outs[root] = out
		self.G2outs[out.Trace] = out
	}
}

func (self *State1) Finalize(saveName string) {
	self.toData()
	self.clear_dirty()
	current_file := zconfig.State1_file(saveName)
	serial := self.data.Serial()
	if err := ioutil.WriteFile(current_file, serial, os.ModePerm); err != nil {
		panic(err)
	} else {
	}
	zconfig.Remove_State1_dir_files(int(self.State0.Num()))
	return
}

func (state *State1) GetOut(root *keys.Uint256) (src *OutState1, e error) {
	state.mu.RLock()
	defer state.mu.RUnlock()
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

func (self *State1) addOut(tks []keys.Uint512, os *zstate.OutState0, os_tree *merkle.Tree) {

	t := utils.TR_enter(fmt.Sprintf("ADD_OUT----INIT num=%v", self.State0.Num()))

	out_hash := os.ToRootCM()
	out_leaf := merkle.Leaf{}
	copy(out_leaf[:], out_hash[:])

	pg, roots := witness.NewPathGenAndRoots(os_tree)
	if roots[0] != out_leaf {
		panic("gen path roots[0] != out leaf")
	}
	if pg.Index != os.Index%(1<<cpt.DEPTH) {
		panic(fmt.Sprintf("gen path index %v != os index %v", pg.Index, os.Index))
	}

	t.Renter("ADD_OUT----MAX_NUM")

	for _, wout := range self.G2wouts {
		if src, err := self.GetOut(&wout); err != nil {
			panic("gen witness wout can not find src")
			return
		} else {
			if self.data.MaxNum < src.Num {
				self.data.MaxNum = src.Num
			}
		}
	}

	t.Renter("ADD_OUT----RANGE.WOUTS")

	index_cur := witness.NewIndexCur(&pg)

	for i := len(self.G2wouts) - 1; i > -1; i-- {
		wout := self.G2wouts[i]
		if src, err := self.GetOut(&wout); err != nil {
			panic("gen witness wout can not find src")
			return
		} else {
			if src.Num < self.data.MaxNum {
				continue
			}
			if src == nil {
				panic("gen witness can not find wout in G2outs")
			} else {
				if !src.Pg.IsComplete() {
					temp_pg := &src.Pg
					witness.NextPathGen(&index_cur, temp_pg, &roots)
					root := temp_pg.Anchor.ToUint256()
					if src.WitnessIndex+1 != os.Index {
						panic("gen witness src.index+1!=os.Index")
					} else {
					}
					if *root != os_tree.RootKey() {
						panic("gen witness src wit root != out")
					} else {
					}
				} else {
				}
				src.WitnessIndex++
				self.add_out_dirty(&wout, src)
			}
		}
	}

	t.Renter("ADD_OUT----ADD_WOUTS")
	self.addWouts(tks, os, &pg)
	t.Leave()
	return
}

func (state *State1) addWouts(tks []keys.Uint512, os *zstate.OutState0, pg *witness.PathGen) {
	for _, tk := range tks {
		if os.IsO() {
			out_o := os.Out_O
			if out_o.Asset.Tkn == nil && out_o.Asset.Tkt == nil {
				break
			}
			no_tkn_value := false
			if out_o.Asset.Tkn != nil {
				if out_o.Asset.Tkn.Value.Cmp(&utils.U256_0) <= 0 {
					no_tkn_value = true
				}
			} else {
				no_tkn_value = true
			}
			no_tkt_value := false
			if out_o.Asset.Tkt != nil {
				if out_o.Asset.Tkt.Value == keys.Empty_Uint256 {
					no_tkt_value = true
				}
			} else {
				no_tkn_value = true
			}

			if no_tkt_value && no_tkn_value {
				break
			}

			if out_o.Addr == (keys.Uint512{}) {
				break
			}

			if succ, rsk := keys.IsMyPKr(&tk, &out_o.Addr); succ {
				out_z := &stx.Out_Z{}
				{
					desc_info := cpt.InfoDesc{}
					desc_info.Ar = utils.NewU256(os.Index).ToRef().ToUint256()
					desc_info.Rsk = rsk
					asset := os.Out_O.Asset.ToCompleteAsset()
					desc_info.Tkn_currency = asset.Tkn.Currency
					desc_info.Tkn_value = asset.Tkn.Value.ToUint256()
					desc_info.Tkt_category = asset.Tkt.Category
					desc_info.Tkt_value = asset.Tkt.Value
					desc_info.Memo = os.Out_O.Memo
					cpt.EncOutput(&desc_info)
					out_z = &stx.Out_Z{}
					out_z.PKr = os.Out_O.Addr
					out_z.EInfo = desc_info.Einfo
					out_z.AssetCM = desc_info.Asset_cm
					out_z.OutCM = *os.ToOutCM()
				}
				root := pg.Root.ToUint256()
				state.append_wout_dirty(root)
				wos := OutState1{}
				wos.Pg = *pg
				wos.Tk = tk
				wos.Out_O = *os.Out_O
				wos.WitnessIndex = os.Index
				wos.OutIndex = os.Index
				wos.Out_Z = out_z
				wos.Z = false
				if *pg.Leaf.ToUint256() != *os.ToRootCM() {
					panic("add wouts but RootCM not match!")
				}
				wos.Trace = cpt.GenTil(&tk, pg.Leaf.ToUint256())
				wos.Num = state.State0.Num()
				state.add_out_dirty(root, &wos)
				state.add_out_dirty(&wos.Trace, &wos)
				break
			}
		} else {
			if succ, rsk := keys.IsMyPKr(&tk, &os.Out_Z.PKr); succ {
				info_desc := cpt.InfoDesc{}
				info_desc.Rsk = rsk
				info_desc.Einfo = os.Out_Z.EInfo
				if err := cpt.DecOutput(&info_desc, &os.Out_Z.AssetCM); err == nil {
					root := pg.Root.ToUint256()
					state.append_wout_dirty(root)
					wos := OutState1{}
					wos.Pg = *pg
					wos.Out_O.Addr = os.Out_Z.PKr
					wos.Out_O.Asset = assets.NewAsset(
						&assets.Token{
							info_desc.Tkn_currency,
							utils.NewU256_ByKey(&info_desc.Tkn_value),
						},
						&assets.Ticket{
							info_desc.Tkt_category,
							info_desc.Tkt_value,
						},
					)
					wos.Out_O.Memo = info_desc.Memo
					wos.Out_Z = os.Out_Z.Clone().ToRef()
					wos.Tk = tk
					wos.WitnessIndex = os.Index
					wos.OutIndex = os.Index
					wos.Z = true
					if *pg.Leaf.ToUint256() != *os.ToRootCM() {
						panic("add wouts but RootCM not match!")
					}
					wos.Trace = cpt.GenTil(&tk, pg.Leaf.ToUint256())
					wos.Num = state.State0.Num()
					state.add_out_dirty(root, &wos)
					state.add_out_dirty(&wos.Trace, &wos)
				}
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
			i := 0
			i++
		} else {
			for i, wout := range state.G2wouts {
				root := src.Pg.Root
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
		t := utils.TR_enter("UpdateWitness---RootKey")
		tree := trees.Trees[trees.Start_index+uint64(i)]
		out := tree.RootKey()
		if os, err := state.State0.GetOut(&out); err != nil {
			panic(err)
		} else {
			if os == nil {
				panic("gen witness out from B2outs can not find in G2outs")
			} else {
			}
			t.Renter("UpdateWitness---ToRootCM")
			os_commitment := os.ToRootCM()
			if commitment != *os_commitment {
				panic("gen witness out!=os.Root()")
			} else {
			}
			t.Renter("UpdateWitness---addOut")
			state.addOut(tks, os, &tree)
		}
		t.Leave()
	}
	return
}

func (self *State1) GetOuts(tk *keys.Uint512) (outs []*OutState1, e error) {
	for _, root := range self.G2wouts {
		if src, err := self.GetOut(&root); err != nil {
			e = err
			return
		} else {
			if src != nil {
				if src.IsMine(tk) {
					if self.State0.HasIn(&src.Trace) {
						panic("get outs src.nil in state0")
					}
					if root != *src.Pg.Root.ToUint256() {
						panic("get outs wout.root!=src.Root")
					}
					outs = append(outs, src)
				}
			} else {
				e = errors.New("get outs can not find src by root")
			}
		}
	}
	SortOutStats(self.State0, outs)
	return
}
