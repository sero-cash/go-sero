package light

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/light/light_generate"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type SLI struct {
}

var SLI_Inst = SLI{}

func (self *SLI) CreateKr() (kr light_types.Kr) {
	rnd := keys.RandUint256()
	zsk := keys.RandUint256()
	vsk := keys.RandUint256()
	zsk = cpt.Force_Fr(&zsk)
	vsk = cpt.Force_Fr(&vsk)

	skr := keys.PKr{}
	copy(skr[:], zsk[:])
	copy(skr[32:], vsk[:])
	copy(skr[64:], rnd[:])

	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	pk := keys.Sk2PK(&sk)

	pkr := keys.Addr2PKr(&pk, &rnd)
	kr.PKr = pkr
	kr.SKr = skr
	return
}

func (self *SLI) DecOuts(outs []light_types.Out, skr *keys.PKr) (douts []light_types.DOut) {
	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := light_types.DOut{}

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

func (self *SLI) GenTx(param *light_types.GenTxParam) (gtx light_types.GTx, e error) {

	if tx, err := light_generate.Generate(param); err != nil {
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

func SignTx(sk *keys.Uint512, paramTx *light_types.GenTxParam) (tx light_types.GTx, err error) {
	tk := keys.Sk2Tk(sk)
	copy(paramTx.From.SKr[:], sk[:])
	for i := range paramTx.Ins {
		out := paramTx.Ins[i]
		if pkr := paramTx.Ins[i].Out.State.OS.ToPKr(); pkr != nil {
			if keys.IsMyPKr(&tk, pkr) {
				copy(paramTx.Ins[i].SKr[:], sk[:])
			} else {
				err = fmt.Errorf("Sign Tx Error:  sk not match the out(%s) pkr(%s) is nil", hexutil.Encode(out.Out.Root[:]), hexutil.Encode(pkr[:]))
				return
			}
		} else {
			err = fmt.Errorf("Sign Tx Error: out(%s) pkr is nil", hexutil.Encode(out.Out.Root[:]))
			return
		}
	}
	if gtx, e := SLI_Inst.GenTx(paramTx); e != nil {
		err = e
		return
	} else {
		tx = gtx
		return
	}
}
