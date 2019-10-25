package flight

import (
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_0"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_1"

	"github.com/sero-cash/go-czero-import/c_type"
)

func GenTx(param *txtool.GTxParam) (gtx txtool.GTx, e error) {
	var need_szk = true

	if (txtool.Ref_inst.Bc != nil) && (!seroparam.Is_Offline()) {
		if txtool.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()+1 < seroparam.SIP5() {
			need_szk = false
		} else {
			param.Z = &need_szk
		}
	} else {
		if param.Z == nil {
			need_szk = false
		}
	}

	if need_szk {
		if tx, param, keys, bases, err := SignTx1(param); err != nil {
			e = err
			return
		} else {
			if gtx, e = ProveTx1(&tx, &param); e != nil {
				return
			} else {
				gtx.Keys = keys
				gtx.Bases = bases
				return
			}
		}
	} else {
		return SignTx0(param)
	}
}

func SignTx(sk *c_type.Uint512, paramTx *txtool.GTxParam) (tx txtool.GTx, e error) {
	copy(paramTx.From.SKr[:], sk[:])
	for i := range paramTx.Ins {
		copy(paramTx.Ins[i].SKr[:], sk[:])
	}
	return GenTx(paramTx)
}

func SignLight(sk *c_type.Uint512, paramTx *txtool.GTxParam) (tx stx.T, param txtool.GTxParam, keys []c_type.Uint256, bases []c_type.Uint256, e error) {
	copy(paramTx.From.SKr[:], sk[:])
	for i := range paramTx.Ins {
		copy(paramTx.Ins[i].SKr[:], sk[:])
	}
	return SignTx1(paramTx)
}

func DecOut(tk *c_type.Tk, outs []txtool.Out) (douts []txtool.TDOut) {
	for _, out := range outs {

		dout := txtool.TDOut{}

		if out.State.OS.Out_O != nil {

			dout.Asset = out.State.OS.Out_O.Asset.Clone()
			dout.Memo = out.State.OS.Out_O.Memo
			if til, e := c_superzk.Czero_genTrace(tk, out.State.OS.RootCM); e == nil {
				dout.Nils = append(dout.Nils, til)
				// dout.Nils = append(dout.Nils, c_czero.GenNil(tk, out.State.OS.RootCM))
				dout.Nils = append(dout.Nils, out.Root)
			}

		} else if out.State.OS.Out_Z != nil {

			if key, flag, e := c_superzk.Czero_fetchKey(tk, &out.State.OS.Out_Z.RPK); e == nil {
				if confirm_out := generate_0.ConfirmOutZ(&key, flag, out.State.OS.Out_Z); confirm_out != nil {
					dout = *confirm_out
					if til, e := c_superzk.Czero_genTrace(tk, out.State.OS.RootCM); e == nil {
						dout.Nils = append(dout.Nils, til)
						// dout.Nils = append(dout.Nils, c_czero.GenNil(tk, out.State.OS.RootCM))
					}
				}
			}

		} else if out.State.OS.Out_P != nil {

			if nl, e := c_superzk.GenNil(tk, out.State.OS.RootCM.NewRef(), &out.State.OS.Out_P.PKr); e == nil {
				dout.Asset = out.State.OS.Out_P.Asset
				dout.Memo = out.State.OS.Out_P.Memo
				dout.Nils = append(dout.Nils, nl)
			}

		} else if out.State.OS.Out_C != nil {

			if key, _, e := c_superzk.FetchKey(&out.State.OS.Out_C.PKr, tk, &out.State.OS.Out_C.RPK); e == nil {
				if confirm_out, _ := generate_1.ConfirmOutC(&key, out.State.OS.Out_C); confirm_out != nil {
					if nl, e := c_superzk.GenNil(tk, out.State.OS.RootCM.NewRef(), &out.State.OS.Out_C.PKr); e == nil {
						dout = *confirm_out
						dout.Nils = append(dout.Nils, nl)
					}
				}
			}

		} else {

			panic(errors.New("OutState has no Out_X"))

		}

		douts = append(douts, dout)
	}
	return
}

func ConfirmOutZ(key *c_type.Uint256, z *stx_v0.Out_Z) (dout txtool.TDOut, e error) {
	if out := generate_0.ConfirmOutZ(key, true, z); out != nil {
		dout = *out
		return
	} else {
		e = errors.New("confirm outz error")
		return
	}
}

func ConfirmOutC(key *c_type.Uint256, c *stx_v1.Out_C) (dout txtool.TDOut, e error) {
	if out, _ := generate_1.ConfirmOutC(key, c); out != nil {
		dout = *out
		return
	} else {
		e = errors.New("confirm outz error")
		return
	}
}

func CurrencyToId(currency string) (ret c_type.Uint256) {
	bs := utils.CurrencyToBytes(currency)
	copy(ret[:], bs[:])
	return
}

func IdToCurrency(id *c_type.Uint256) (ret string) {
	return utils.Uint256ToCurrency(id)
}
