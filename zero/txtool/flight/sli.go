package flight

import (
	"encoding/json"
	"log"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-sero/zero/txtool/generate"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type SLI struct {
}

var SLI_Inst = SLI{}

func DecTraceOuts(outs []txtool.Out, skr *keys.PKr) (douts []txtool.TDOut) {
	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := txtool.TDOut{}

		if out.State.OS.Out_O != nil {
			dout.Asset = out.State.OS.Out_O.Asset.Clone()
			dout.Memo = out.State.OS.Out_O.Memo
			dout.Nils = append(dout.Nils, cpt.GenTil(&sk, out.State.OS.RootCM))
			dout.Nils = append(dout.Nils, cpt.GenNil(&sk, out.State.OS.RootCM))
			dout.Nils = append(dout.Nils, out.Root)
		} else {
			key, flag := keys.FetchKey(&sk, &out.State.OS.Out_Z.RPK)
			info_desc := cpt.InfoDesc{}
			info_desc.Key = key
			info_desc.Flag = flag
			info_desc.Einfo = out.State.OS.Out_Z.EInfo
			cpt.DecOutput(&info_desc)

			if e := stx.ConfirmOut_Z(&info_desc, out.State.OS.Out_Z); e == nil {
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
				dout.Nils = append(dout.Nils, cpt.GenTil(&sk, out.State.OS.RootCM))
				dout.Nils = append(dout.Nils, cpt.GenNil(&sk, out.State.OS.RootCM))
			}
		}
		douts = append(douts, dout)
	}
	return
}

func (self *SLI) DecOuts(outs []txtool.Out, skr *keys.PKr) (douts []txtool.DOut) {
	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := txtool.DOut{}

		data, _ := json.Marshal(out)
		log.Printf("DecOuts out : %s", string(data))

		if out.State.OS.Out_O != nil {
			dout.Asset = out.State.OS.Out_O.Asset.Clone()
			dout.Memo = out.State.OS.Out_O.Memo
			dout.Nil = cpt.GenNil(&sk, out.State.OS.RootCM)
			log.Printf("DecOuts Out_O!= nil")
		} else {
			key, flag := keys.FetchKey(&sk, &out.State.OS.Out_Z.RPK)
			info_desc := cpt.InfoDesc{}
			info_desc.Key = key
			info_desc.Flag = flag
			info_desc.Einfo = out.State.OS.Out_Z.EInfo
			cpt.DecOutput(&info_desc)

			data, _ := json.Marshal(info_desc)
			log.Printf("DecOuts info_desc : %s", string(data))

			if e := stx.ConfirmOut_Z(&info_desc, out.State.OS.Out_Z); e == nil {
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
				dout.Nil = cpt.GenNil(&sk, out.State.OS.RootCM)
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

func (self *SLI) GenTx(param *txtool.GTxParam) (gtx txtool.GTx, e error) {

	if tx, err := generate.GenTx(param); err != nil {
		e = err
		return
	} else {
		gtx.Tx = tx
		gtx.Gas = hexutil.Uint64(param.Gas)
		gtx.GasPrice = hexutil.Big(param.GasPrice)
		gtx.Hash = tx.ToHash()
		return
	}
}

func SignTx(sk *keys.Uint512, paramTx *txtool.GTxParam) (tx txtool.GTx, err error) {
	copy(paramTx.From.SKr[:], sk[:])
	for i := range paramTx.Ins {
		copy(paramTx.Ins[i].SKr[:], sk[:])
	}
	if gtx, e := SLI_Inst.GenTx(paramTx); e != nil {
		err = e
		return
	} else {
		tx = gtx
		return
	}
}
