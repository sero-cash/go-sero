package ssi

import (
	"encoding/json"
	"log"

	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_0"

	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_1"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func DecNilOuts(outs []txtool.Out, skr *c_type.PKr) (douts []txtool.DOut) {
	sk := c_type.Uint512{}
	copy(sk[:], skr[:])
	tk := superzk.Sk2Tk(&sk)
	for _, out := range outs {
		dout := txtool.DOut{}

		data, _ := json.Marshal(out)
		log.Printf("DecOuts out : %s", string(data))

		if out.State.OS.Out_O != nil {
			dout.Asset = out.State.OS.Out_O.Asset.Clone()
			dout.Memo = out.State.OS.Out_O.Memo
			if nl, e := c_superzk.Czero_genNil(&sk, out.State.OS.RootCM); e == nil {
				dout.Nil = nl
			}
			log.Printf("DecOuts Out_O")
		} else if out.State.OS.Out_Z != nil {
			if key, flag, e := c_superzk.Czero_fetchKey(&tk, &out.State.OS.Out_Z.RPK); e == nil {
				if o := generate_0.ConfirmOutZ(&key, flag, out.State.OS.Out_Z); o != nil {
					dout.Asset = o.Asset
					dout.Memo = o.Memo
					if nl, e := c_superzk.Czero_genNil(&sk, out.State.OS.RootCM); e == nil {
						dout.Nil = nl
						log.Printf("DecOuts success")
					}
				}
			}
			log.Printf("DecOuts Out_Z")
		} else if out.State.OS.Out_P != nil {
			if nl, e := c_superzk.GenNil(&tk, out.State.OS.RootCM, &out.State.OS.Out_P.PKr); e == nil {
				dout.Asset = out.State.OS.Out_P.Asset.Clone()
				dout.Memo = out.State.OS.Out_P.Memo
				dout.Nil = nl
				log.Printf("DecOuts success")
			}
			log.Printf("DecOuts Out_P")
		} else if out.State.OS.Out_C != nil {
			if key, e := c_superzk.FetchKey(&out.State.OS.Out_C.PKr, &tk, &out.State.OS.Out_C.RPK); e == nil {
				if o := generate_1.ConfirmOutC(&key, out.State.OS.Out_C); o != nil {
					if nl, e := c_superzk.GenNil(&tk, out.State.OS.RootCM.NewRef(), out.State.OS.ToPKr()); e == nil {
						dout.Asset = o.Asset
						dout.Memo = o.Memo
						dout.Nil = nl
						log.Printf("DecOuts success")
					}
				}
			}
			log.Printf("DecOuts Out_C")
		}
		douts = append(douts, dout)

		data, _ = json.Marshal(douts)
		log.Printf("DecOuts douts : %s", string(data))
	}
	return
}
