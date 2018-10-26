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
	"fmt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/tx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/state1"
	"github.com/sero-cash/go-sero/zero/utils"
)

type preTxDesc struct {
	ins  []state1.OutState1
	outs []tx.Out
}

type preTx struct {
	last_anchor keys.Uint256
	uouts       []state1.OutState1
	desc_o      preTxDesc
	desc_z      preTxDesc
}

type cyState struct {
	balance utils.U256
}

type cyStateMap map[keys.Uint256]*cyState

func (self cyStateMap) add(key *keys.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.AddU(value)
	} else {
		self[*key] = &cyState{
			*value.ToRef(),
		}
	}
}
func (self cyStateMap) sub(key *keys.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.SubU(value)
	} else {
		self[*key] = &cyState{
			*value.ToRef(),
		}
	}
}

func newcyStateMap(fee *utils.U256) (ret cyStateMap) {
	ret = make(map[keys.Uint256]*cyState)
	sero_currency := keys.Uint256{}
	copy(sero_currency[:], []byte("SERO"))
	ret[sero_currency] = &cyState{
		balance: *fee.ToRef(),
	}
	return
}

func preGen(ts *tx.T, state *zstate.State) (p preTx, e error) {
	state1 := state1.CurrentState1()
	p.last_anchor = state.State0.Cur.Tree.RootKey()
	cy_state_map := newcyStateMap(&ts.Fee)
	tk_map := make(map[keys.Uint256]int)
	for _, in := range ts.Ins {
		if src, err := state1.GetOut(&in.Root); err == nil {
			added := false
			if src.Out_O.Pkg.Tkn != nil {
				cy_state_map.add(&src.Out_O.Pkg.Tkn.Currency, &src.Out_O.Pkg.Tkn.Value)
				added = true
			}
			if src.Out_O.Pkg.Tkt != nil {
				if _, ok := tk_map[src.Out_O.Pkg.Tkt.Value]; !ok {
					tk_map[src.Out_O.Pkg.Tkt.Value] = 1
				} else {
					e = fmt.Errorf("in tkt duplicate: %v", src.Out_O.Pkg.Tkt.Value)
					return
				}
				added = true
			}
			if added {
				if src.Out_Z == nil {
					p.desc_o.ins = append(p.desc_o.ins, *src)
				} else {
					p.desc_z.ins = append(p.desc_o.ins, *src)
				}
			}
			p.uouts = append(p.uouts, *src)
		} else {
			e = err
			return
		}
	}
	for _, out := range ts.Outs {
		added := false
		if out.Pkg.Tkn != nil {
			cy_state_map.sub(&out.Pkg.Tkn.Currency, &out.Pkg.Tkn.Value)
			added = true
		}
		if out.Pkg.Tkt != nil {
			if c, ok := tk_map[out.Pkg.Tkt.Value]; ok {
				if c > 0 {
					tk_map[out.Pkg.Tkt.Value] = c - 1
				} else {
					e = fmt.Errorf("out tkt duplicate: %v", out.Pkg.Tkt.Value)
					return
				}
			} else {
				e = fmt.Errorf("out tkt not in ins: %v", out.Pkg.Tkt.Value)
				return
			}
			added = true
		}
		if added {
			switch out.Z {
			case tx.TYPE_N:
				fallthrough
			case tx.TYPE_O:
				p.desc_o.outs = append(p.desc_o.outs, out)
			default:
				p.desc_z.outs = append(p.desc_z.outs, out)
			}
		}
	}

	for currency, state := range cy_state_map {
		if state.balance.Cmp(&utils.U256_0) != 0 {
			e = fmt.Errorf("currency %v banlance != 0", currency)
			return
		} else {
		}
	}

	for ticket, c := range tk_map {
		if c > 0 {
			e = fmt.Errorf("tikect not use %v", ticket)
			return
		}
	}

	return
}
