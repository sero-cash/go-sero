package generate_0

import (
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func confirmOut_Z(deInfo *c_czero.InfoDesc, out_z *stx_v0.Out_Z) (e error) {
	desc := c_czero.ConfirmOutputDesc{}
	desc.Memo = deInfo.Memo
	desc.Asset = deInfo.Asset
	desc.Rsk = deInfo.Rsk
	desc.Pkr = out_z.PKr
	desc.Out_cm = out_z.OutCM
	e = c_czero.ConfirmOutput(&desc)
	return
}

func ConfirmOutZ(key *c_type.Uint256, flag bool, outz *stx_v0.Out_Z) (dout *txtool.TDOut) {
	info := c_czero.InfoDesc{}
	info.Key = *key
	info.Flag = flag
	info.Einfo = outz.EInfo
	c_czero.DecOutput(&info)
	if e := confirmOut_Z(&info, outz); e == nil {
		dout = &txtool.TDOut{}
		dout.Asset = assets.NewAssetByType(&info.Asset)
		dout.Memo = info.Memo
	}
	return
}
