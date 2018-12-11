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
	"encoding/hex"
	"fmt"

	"github.com/sero-cash/go-sero/zero/txs/zstate/pkgstate"

	"github.com/sero-cash/go-sero/zero/txs/lstate"

	"github.com/sero-cash/go-sero/zero/txs/pkg"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/tx"
)

type preTxDesc struct {
	ins  []lstate.OutState
	outs []tx.Out
}

type prePkgPack struct {
	pkg tx.PkgPack
}

type prePkgOpen struct {
	opkg pkgstate.OPkg
}

type prePkgChange struct {
	pkr  keys.Uint512
	zpkg pkgstate.ZPkg
}

type prePkgDesc struct {
	pack   *prePkgPack
	change *prePkgChange
	open   *prePkgOpen
}

type preTx struct {
	last_anchor keys.Uint256
	uouts       []lstate.OutState
	desc_o      preTxDesc
	desc_z      preTxDesc
	desc_pkg    prePkgDesc
}

func preGen(ts *tx.T, state1 *lstate.State) (p preTx, e error) {
	p.last_anchor = state1.State.State.Cur.Tree.RootKey()
	ck_state := NewCKState(&ts.Fee)

	for _, in := range ts.Ins {
		if src, err := state1.GetOut(&in.Root); err == nil {
			if added, err := ck_state.AddIn(&src.Out_O.Asset); err != nil {
				e = err
				return
			} else {
				if added {
					if src.Out_Z == nil {
						p.desc_o.ins = append(p.desc_o.ins, *src)
					} else {
						p.desc_z.ins = append(p.desc_z.ins, *src)
					}
				}
				p.uouts = append(p.uouts, *src)
			}
		} else {
			e = err
			return
		}
	}

	for _, out := range ts.Outs {
		if added, err := ck_state.AddOut(&out.Asset); err != nil {
			e = err
			return
		} else {
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
	}

	if ts.PkgPack != nil {
		if _, err := ck_state.AddOut(&ts.PkgPack.Pkg.Asset); err != nil {
			e = err
			return
		} else {
			p.desc_pkg.pack = &prePkgPack{}
			p.desc_pkg.pack.pkg = *ts.PkgPack
		}
	}

	if ts.PkgOpen != nil {
		if zpkg := state1.State.Pkgs.GetPkg(&ts.PkgOpen.Id); zpkg == nil {
			e = fmt.Errorf("Get Pkg error %v", hex.EncodeToString(ts.PkgOpen.Id[:]))
			return
		} else {
			if opkg, err := pkg.DePkg(&ts.PkgOpen.Key, &zpkg.Pack.Pkg); err != nil {
				e = fmt.Errorf("Decode Pkg error %v", hex.EncodeToString(ts.PkgOpen.Id[:]))
				return
			} else {
				if _, err := ck_state.AddIn(&opkg.Asset); err != nil {
					e = err
					return
				} else {
					p.desc_pkg.open = &prePkgOpen{}
					p.desc_pkg.open.opkg.O = opkg
					p.desc_pkg.open.opkg.Z = *zpkg
				}
			}
		}
	}

	if ts.PkgChange != nil {
		if zpkg := state1.State.Pkgs.GetPkg(&ts.PkgOpen.Id); zpkg == nil {
			e = fmt.Errorf("Get Pkg error %v", hex.EncodeToString(ts.PkgOpen.Id[:]))
			return
		} else {
			p.desc_pkg.change = &prePkgChange{}
			p.desc_pkg.change.pkr = ts.PkgChange.PKr
			p.desc_pkg.change.zpkg = *zpkg
		}
	}

	if err := ck_state.Check(); err != nil {
		e = err
		return
	} else {
		return
	}
}
