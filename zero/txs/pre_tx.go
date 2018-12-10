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

	"github.com/sero-cash/go-sero/zero/txs/lstate"

	"github.com/sero-cash/go-sero/zero/txs/pkg"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/tx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type preTxDesc struct {
	ins  []lstate.OutState
	outs []tx.Out
}

type prePkgOpen struct {
	id   uint64
	opkg pkg.Pkg_O
	zpkg pkg.Pkg_Z
}
type prePkgDesc struct {
	pack *pkg.Pkg_O
	open *prePkgOpen
}

type preTx struct {
	last_anchor keys.Uint256
	uouts       []lstate.OutState
	desc_o      preTxDesc
	desc_z      preTxDesc
	desc_pkg    prePkgDesc
}

type cyState struct {
	balance utils.I256
}

type cyStateMap map[keys.Uint256]*cyState

func (self cyStateMap) add(key *keys.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.AddU(value)
	} else {
		self[*key] = &cyState{
			*value.ToI256().ToRef(),
		}
	}
}
func (self cyStateMap) sub(key *keys.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.SubU(value)
	} else {
		self[*key] = &cyState{}
		self[*key].balance.SubU(value)
	}
}

func newcyStateMap() (ret cyStateMap) {
	ret = make(map[keys.Uint256]*cyState)
	return
}

func preGen(ts *tx.T, state1 *lstate.State) (p preTx, e error) {
	p.last_anchor = state1.State0.Cur.Tree.RootKey()
	cy_state_map := newcyStateMap()
	cy_state_map.sub(&ts.Fee.Currency, &ts.Fee.Value)
	tk_map := make(map[keys.Uint256]int)
	for _, in := range ts.Ins {
		if src, err := state1.GetOut(&in.Root); err == nil {
			added := false
			if src.Out_O.Asset.Tkn != nil {
				cy_state_map.add(&src.Out_O.Asset.Tkn.Currency, &src.Out_O.Asset.Tkn.Value)
				added = true
			}
			if src.Out_O.Asset.Tkt != nil {
				if _, ok := tk_map[src.Out_O.Asset.Tkt.Value]; !ok {
					tk_map[src.Out_O.Asset.Tkt.Value] = 1
				} else {
					e = fmt.Errorf("in tkt duplicate: %v", src.Out_O.Asset.Tkt.Value)
					return
				}
				added = true
			}
			if added {
				if src.Out_Z == nil {
					p.desc_o.ins = append(p.desc_o.ins, *src)
				} else {
					p.desc_z.ins = append(p.desc_z.ins, *src)
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
		if out.Asset.Tkn != nil {
			cy_state_map.sub(&out.Asset.Tkn.Currency, &out.Asset.Tkn.Value)
			added = true
		}
		if out.Asset.Tkt != nil {
			if c, ok := tk_map[out.Asset.Tkt.Value]; ok {
				if c > 0 {
					tk_map[out.Asset.Tkt.Value] = c - 1
				} else {
					e = fmt.Errorf("out tkt duplicate: %v", out.Asset.Tkt.Value)
					return
				}
			} else {
				e = fmt.Errorf("out tkt not in ins: %v", out.Asset.Tkt.Value)
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
	if ts.PkgPack != nil {
		out := ts.PkgPack
		if out.Asset.Tkn != nil {
			cy_state_map.sub(&out.Asset.Tkn.Currency, &out.Asset.Tkn.Value)
		}
		if out.Asset.Tkt != nil {
			if c, ok := tk_map[out.Asset.Tkt.Value]; ok {
				if c > 0 {
					tk_map[out.Asset.Tkt.Value] = c - 1
				} else {
					e = fmt.Errorf("out tkt duplicate: %v", out.Asset.Tkt.Value)
					return
				}
			} else {
				e = fmt.Errorf("out tkt not in ins: %v", out.Asset.Tkt.Value)
				return
			}
		}
		p.desc_pkg.pack = &pkg.Pkg_O{
			out.PKr,
			out.Asset,
			keys.Uint512{},
		}
	}

	for currency, state := range cy_state_map {
		if state.balance.Cmp(&utils.I256_0) != 0 {
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
