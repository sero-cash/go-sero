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
	"fmt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/tx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/state1"
	"github.com/sero-cash/go-sero/zero/utils"
)

func verifyIn_O(d *keys.Uint256, sign *keys.Uint256, r *keys.Uint256) error {
	if keys.VerifyOAddr(d, sign, r) {
		return nil
	} else {
		return errors.New("txs.verify in_o failed")
	}
}

func Gen(seed *keys.Uint256, t *tx.T, state *zstate.State) (s stx.T, e error) {
	if preTx, err := preGen(t, state); err == nil {
		if len(preTx.desc_os) > 2 {
			e = errors.New("pre tx currency size > 2!!!")
			return
		}
		s.Ehash = t.Ehash
		var type_o_addr *keys.Uint512
		for _, desc_o := range preTx.desc_os {
			s_desc_o := stx.Desc_O{}
			s_desc_o.Fee = desc_o.fee
			s_desc_o.Z2O = desc_o.z2o
			s_desc_o.Z2OIndex = uint64(preTx.C2I.addC(&desc_o.currency))
			s_desc_o.Currency = desc_o.currency
			for _, in := range desc_o.ins {
				s_desc_o.Ins = append(
					s_desc_o.Ins,
					stx.In_O{
						Root: *in.Pg.Root.ToUint256(),
					},
				)
			}
			for _, out := range desc_o.outs {
				out_o := stx.Out_O{}
				out_o.Value = out.Value
				out_o.Memo = out.Memo
				switch out.Z {
				case tx.TYPE_O:
					pkr := keys.Addr2PKr(&out.Addr, keys.RandUint256().NewRef())
					out_o.Addr = pkr
				case tx.TYPE_N:
					out_o.Addr = out.Addr
					type_o_addr = &out.Addr
				default:
					panic("Gen desc_o out but z is type_z")
				}
				s_desc_o.Outs = append(s_desc_o.Outs, out_o)
			}
			s.Desc_Os = append(s.Desc_Os, s_desc_o)
		}

		addr := keys.Seed2Addr(seed)
		var from_r *keys.Uint256
		if type_o_addr != nil {
			from_r = new(keys.Uint256)
			copy(from_r[:], type_o_addr[:16])
		} else {
		}
		s.From = keys.Addr2PKr(&addr, from_r)

		hash_o := s.ToHash_for_z()
		if desc_zs, err := genDesc_Zs(seed, &preTx, &hash_o); err != nil {
			e = err
		} else {
			s.Desc_Zs = desc_zs
		}

		hash_z := s.ToHash_for_o()

		for _, desc_o := range s.Desc_Os {
			if p_desc_o, ok := preTx.desc_os[desc_o.Currency]; !ok {
				panic(fmt.Errorf("can not find desc_o for %v !!!", desc_o.Currency))
				return
			} else {
				for i := range p_desc_o.ins {
					if sign, err := keys.SignOAddr(seed, &hash_z, nil, nil); err == nil {
						desc_o.Ins[i].Sign = sign
					} else {
						e = err
						return
					}
				}
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
	for _, desc_o := range s.Desc_Os {
		if !CheckInt(&desc_o.Z2O) {
			e = errors.New("verify check z2o too big")
			return
		}
		if !CheckUint(&desc_o.Fee) {
			e = errors.New("verify check fee too big")
			return
		}
		balance := utils.NewI256(0)
		balance.AddI(&desc_o.Z2O)
		balance.SubU(&desc_o.Fee)
		for _, in := range desc_o.Ins {
			if ok := state.State0.HasIn(&in.Root); ok {
				e = errors.New("in already in nils")
				return
			} else {
			}
			if src, err := state.State0.GetOut(&in.Root); e == nil {
				if src.IsO() {
					if err := verifyIn_O(&hash_z, &in.Sign, nil); err == nil {
						if !CheckUint(&src.Out_O.Out.Value) {
							e = errors.New("verify check out value too big")
							return
						}
						balance.AddU(&src.Out_O.Out.Value)
						if !CheckInt(&balance) {
							e = errors.New("verify check balance too big")
							return
						}
					} else {
						e = err
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

		for _, out := range desc_o.Outs {
			balance.SubU(&out.Value)
			if !CheckInt(&balance) {
				e = errors.New("verify check balance too big")
				return
			}
		}

		if balance.Cmp(&utils.I256_0) != 0 {
			e = errors.New("Verify o banlance is not 0")
			return
		} else {
		}
	}

	for _, desc_z := range s.Desc_Zs {
		if ok := state.State0.HasIn(&desc_z.In.Nil); ok {
			e = errors.New("Verify in already in nils")
			return
		} else {
		}
		if out, err := state.State0.GetOut(&desc_z.In.Anchor); err != nil {
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
			if out.Out_O.Currency == *currency {
				roots = append(roots, *root)
				amount.AddU(&out.Out_O.Out.Value)
				out_o := out.Out_O
				if value.Cmp(out_o.Out.Value.ToI256().ToRef()) < 0 {
					value = utils.NewI256(0)
					break
				} else {
					value.SubU(&out_o.Out.Value)
				}
			} else {
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
