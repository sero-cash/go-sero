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

package txs

import (
	"errors"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/tx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/state1"
	"github.com/sero-cash/go-sero/zero/utils"
)

func Gen(seed *keys.Uint256, t *tx.T, state *zstate.State) (s stx.T, e error) {
	if preTx, err := preGen(t, state); err == nil {
		s.Ehash = t.Ehash
		s.Fee = t.Fee
		var type_o_addr *keys.Uint512
		for _, in_o := range preTx.desc_o.ins {
			s_in_o := stx.In_O{}
			s_in_o.Root = *in_o.Pg.Root.ToUint256()
			s.Desc_O.Ins = append(s.Desc_O.Ins, s_in_o)
		}
		for _, out_o := range preTx.desc_o.outs {
			s_out_o := stx.Out_O{}
			s_out_o.Pkg = out_o.Pkg.Clone()
			s_out_o.Memo = out_o.Memo
			switch out_o.Z {
			case tx.TYPE_O:
				//pkr := keys.Addr2PKr(&out_o.Addr, keys.RandUint256().NewRef())
				pkr := keys.Seed2Tk(seed)
				out_o.Addr = pkr
			case tx.TYPE_N:
				out_o.Addr = out_o.Addr
				type_o_addr = &out_o.Addr
			default:
				panic("Gen desc_o out but z is type_z")
			}
			s.Desc_O.Outs = append(s.Desc_O.Outs, s_out_o)
		}

		//addr := keys.Seed2Addr(seed)
		var from_r *keys.Uint256
		if type_o_addr != nil {
			from_r = new(keys.Uint256)
			copy(from_r[:], type_o_addr[:16])
		} else {
		}
		//s.From = keys.Addr2PKr(&addr, from_r)
		s.From = keys.Seed2Tk(seed)

		hash_o := s.ToHash_for_z()
		if desc_z, err := genDesc_Zs(seed, &preTx, &hash_o); err != nil {
			e = err
		} else {
			s.Desc_Z = desc_z
		}

		hash_z := s.ToHash_for_o()

		if sign, err := keys.SignOAddr(seed, &hash_z, &s.From); err != nil {
			e = err
			return
		} else {
			s.Sign = sign
		}

		for i, s_in_o := range preTx.desc_o.ins {
			if sign, err := keys.SignOAddr(seed, &hash_z, &s_in_o.Out_O.Addr); err != nil {
				e = err
				return
			} else {
				s.Desc_O.Ins[i].Sign = sign
			}
		}

		for _, used_out := range preTx.uouts {
			state1.UpdateOutStat(&state.State0, &used_out)
		}

		return
	} else {
		e = err
		return
	}
}

func CheckUint(i *utils.U256) bool {
	u := i.ToUint256()
	m := u[31] & (0xF << 4)
	if m != 0 {
		return false
	} else {
		return true
	}
}

func CheckInt(i *utils.I256) bool {
	abs := i.Abs()
	return CheckUint(&abs)
}

func Verify(s *stx.T, state *zstate.State) (e error) {
	hash_z := s.ToHash_for_o()
	if !CheckUint(&s.Fee) {
		e = errors.New("verify check fee too big")
		return
	}
	for _, in_o := range s.Desc_O.Ins {
		if ok := state.State0.HasIn(&in_o.Root); ok {
			e = errors.New("in already in nils")
			return
		} else {
		}
		if src, err := state.State0.GetOut(&in_o.Root); e == nil {
			if src.IsO() {
				if keys.VerifyOAddr(&hash_z, &in_o.Sign, &src.Out_O.Addr) {
				} else {
					e = errors.New("txs.verify in_o failed")
					return
				}
			} else {
				e = errors.New("txs.Verify src is z,but in is o")
				return
			}
		} else {
			e = err
			return
		}
	}
	for _, out_o := range s.Desc_O.Outs {
		if out_o.Pkg.Tkn != nil {
			if !CheckUint(&out_o.Pkg.Tkn.Value) {
				e = errors.New("verify check balance too big")
				return
			}
		}
	}

	for _, in_z := range s.Desc_Z.Ins {
		if ok := state.State0.HasIn(&in_z.Nil); ok {
			e = errors.New("Verify in already in nils")
			return
		} else {
		}
		if out, err := state.State0.GetOut(&in_z.Anchor); err != nil {
			e = err
			return
		} else {
			if out == nil {
				e = errors.New("Verify can not find out for anchor")
			} else {
			}
		}
	}

	if err := verifyDesc_Zs(s); err != nil {
		e = err
		return
	} else {
	}

	return
}

func GetOuts(tk *keys.Uint512, state *zstate.State) (outs []*state1.OutState1, e error) {
	st1 := state1.CurrentState1()
	for _, root := range st1.G2wouts {
		if src, err := st1.GetOut(&root); err != nil {
			e = err
			return
		} else {
			if src != nil {
				if src.IsMine(tk) {
					if state.State0.HasIn(&src.Trace) {
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
	state1.SortOutStats(&state.State0, outs)
	return
}

func GetRoots(tk *keys.Uint512, state *zstate.State, v *utils.U256, currency *keys.Uint256) (roots []keys.Uint256, amount utils.U256, e error) {
	value := v.ToI256()
	if outs, err := GetOuts(tk, state); err != nil {
		e = err
		return
	} else {
		for _, out := range outs {
			root := out.Pg.Root.ToUint256()
			if out.Out_O.Pkg.Tkn != nil {
				if out.Out_O.Pkg.Tkn.Currency == *currency {
					roots = append(roots, *root)
					amount.AddU(&out.Out_O.Pkg.Tkn.Value)
					out_o := out.Out_O
					if value.Cmp(out_o.Pkg.Tkn.Value.ToI256().ToRef()) < 0 {
						value = utils.NewI256(0)
						break
					} else {
						value.SubU(&out_o.Pkg.Tkn.Value)
					}
				} else {
				}
			}
		}
		if value.Cmp(&utils.I256_0) == 0 {
			return
		} else {
			e = errors.New("can not find enough outs")
			return
		}
	}
}
