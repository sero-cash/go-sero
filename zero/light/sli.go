package light

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type SLI struct {
}

var SLI_Inst = SLI{}

func (self *SLI) CreateKr() (kr Kr) {
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

func (self *SLI) DecOuts(outs []Out, skr *keys.PKr) (douts []DOut) {
	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := DOut{}
		if out.Out_O != nil {
			dout.Asset = out.Out_O.Asset.Clone()
			dout.Memo = out.Out_O.Memo
			dout.Nil = cpt.GenNil(&sk, out.RootCM)
		} else {
			key, flag := keys.FetchKey(&sk, &out.Out_Z.RPK)
			info_desc := cpt.InfoDesc{}
			info_desc.Key = key
			info_desc.Flag = flag
			info_desc.Einfo = out.Out_Z.EInfo
			cpt.DecOutput(&info_desc)
			if e := stx.ConfirmOut_Z(&info_desc, out.Out_Z); e == nil {
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
				dout.Nil = cpt.GenNil(&sk, out.RootCM)
			}
		}
		douts = append(douts, dout)
	}
	return
}

func (self *SLI) GenTx(param *GenTxParam) (gtx GTx, e error) {
	return
}
