package flight

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_0"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_1"

	"github.com/sero-cash/go-czero-import/c_type"
)

func GenTx(param *txtool.GTxParam) (gtx txtool.GTx, e error) {
	if param.IsSzk() {
		str, _ := json.Marshal(param)
		fmt.Println(string(str))
		if tx, param, keys, err := SignTx1(param); err != nil {
			e = err
			return
		} else {
			if gtx, e = ProveTx1(&tx, &param); e != nil {
				return
			} else {
				gtx.Keys = keys
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

func SignLight(sk *c_type.Uint512, paramTx *txtool.GTxParam) (tx stx.T, param txtool.GTxParam, keys []c_type.Uint256, e error) {
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
				//dout.Nils = append(dout.Nils, c_czero.GenNil(tk, out.State.OS.RootCM))
				dout.Nils = append(dout.Nils, out.Root)
			}

		} else if out.State.OS.Out_Z != nil {

			if key, flag, e := c_superzk.Czero_fetchKey(tk, &out.State.OS.Out_Z.RPK); e == nil {
				if confirm_out := generate_0.ConfirmOutZ(&key, flag, out.State.OS.Out_Z); confirm_out != nil {
					dout = *confirm_out
					if til, e := c_superzk.Czero_genTrace(tk, out.State.OS.RootCM); e == nil {
						dout.Nils = append(dout.Nils, til)
						//dout.Nils = append(dout.Nils, c_czero.GenNil(tk, out.State.OS.RootCM))
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

			if key, e := c_superzk.FetchKey(&out.State.OS.Out_C.PKr, tk, &out.State.OS.Out_C.RPK); e == nil {
				if confirm_out := generate_1.ConfirmOutC(&key, out.State.OS.Out_C); confirm_out != nil {
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
