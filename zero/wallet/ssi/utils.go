package ssi

import (
	"encoding/json"
	"log"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"
)

func DecNilOuts(outs []txtool.Out, skr *c_type.PKr) (douts []txtool.DOut) {
	sk := c_type.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := txtool.DOut{}

		data, _ := json.Marshal(out)
		log.Printf("DecOuts out : %s", string(data))

		if out.State.OS.Out_O != nil {
			dout.Asset = out.State.OS.Out_O.Asset.Clone()
			dout.Memo = out.State.OS.Out_O.Memo
			dout.Nil = c_czero.GenNil(&sk, out.State.OS.RootCM)
			log.Printf("DecOuts Out_O!= nil")
		} else {
			key, flag := c_czero.FetchKey(&sk, &out.State.OS.Out_Z.RPK)
			info_desc := c_czero.InfoDesc{}
			info_desc.Key = key
			info_desc.Flag = flag
			info_desc.Einfo = out.State.OS.Out_Z.EInfo
			c_czero.DecOutput(&info_desc)

			data, _ := json.Marshal(info_desc)
			log.Printf("DecOuts info_desc : %s", string(data))

			if e := stx_v1.ConfirmOut_Z(&info_desc, out.State.OS.Out_Z); e == nil {
				dout.Asset = assets.NewAsset(
					&assets.Token{
						info_desc.Tkn_currency,
						utils.NewU256_ByKey(&info_desc.Tkn_value),
					},
					&assets.Ticket{
						info_desc.Tkt_category,
						info_desc.Tkt_value,
					},
				)
				dout.Memo = info_desc.Memo
				dout.Nil = c_czero.GenNil(&sk, out.State.OS.RootCM)
				log.Printf("DecOuts success")
			}
			log.Printf("DecOuts Out_O == nil")
		}
		douts = append(douts, dout)

		data, _ = json.Marshal(douts)
		log.Printf("DecOuts douts : %s", string(data))
	}
	return
}
