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

package generate

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

type prePkgCreate struct {
	pkg tx.PkgCreate
}

type prePkgClose struct {
	opkg pkgstate.OPkg
}

type prePkgTransfer struct {
	pkr  keys.PKr
	zpkg pkgstate.ZPkg
}

type prePkgDesc struct {
	create   *prePkgCreate
	transfer *prePkgTransfer
	close    *prePkgClose
}

type preTx struct {
	uouts    []lstate.OutState
	desc_o   preTxDesc
	desc_z   preTxDesc
	desc_pkg prePkgDesc
}

func preGen(ts *tx.T, state1 *lstate.State) (p preTx, e error) {
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
				if out.IsZ {
					p.desc_z.outs = append(p.desc_z.outs, out)
				} else {
					p.desc_o.outs = append(p.desc_o.outs, out)
				}
			}
		}
	}

	if ts.PkgCreate != nil {
		if _, err := ck_state.AddOut(&ts.PkgCreate.Pkg.Asset); err != nil {
			e = err
			return
		} else {
			p.desc_pkg.create = &prePkgCreate{}
			p.desc_pkg.create.pkg = *ts.PkgCreate
		}
	}

	if ts.PkgClose != nil {
		if zpkg := state1.State.Pkgs.GetPkg(&ts.PkgClose.Id); zpkg == nil {
			e = fmt.Errorf("Get Pkg error %v", hex.EncodeToString(ts.PkgClose.Id[:]))
			return
		} else {
			if opkg, err := pkg.DePkg(&ts.PkgClose.Key, &zpkg.Pack.Pkg); err != nil {
				e = fmt.Errorf("Decode Pkg error %v", hex.EncodeToString(ts.PkgClose.Id[:]))
				return
			} else {
				if e = pkg.ConfirmPkg(&opkg, &zpkg.Pack.Pkg); e != nil {
					return
				} else {
					if _, e = ck_state.AddIn(&opkg.Asset); e != nil {
						return
					} else {
						p.desc_pkg.close = &prePkgClose{}
						p.desc_pkg.close.opkg.O = opkg
						p.desc_pkg.close.opkg.Z = *zpkg
					}
				}
			}
		}
	}

	if ts.PkgTransfer != nil {
		if zpkg := state1.State.Pkgs.GetPkg(&ts.PkgTransfer.Id); zpkg == nil {
			e = fmt.Errorf("Get Pkg error %v", hex.EncodeToString(ts.PkgTransfer.Id[:]))
			return
		} else {
			p.desc_pkg.transfer = &prePkgTransfer{}
			p.desc_pkg.transfer.pkr = ts.PkgTransfer.PKr
			p.desc_pkg.transfer.zpkg = *zpkg
		}
	}

	if err := ck_state.Check(); err != nil {
		e = err
		return
	} else {
		return
	}
}
