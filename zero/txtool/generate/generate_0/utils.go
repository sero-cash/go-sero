package generate_0

import (
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"
)

func ConfirmOutZ(key *c_type.Uint256, flag bool, outz *stx_v0.Out_Z) (dout *txtool.TDOut) {
	info := c_czero.InfoDesc{}
	info.Key = *key
	info.Flag = flag
	info.Einfo = outz.EInfo
	c_czero.DecOutput(&info)
	if e := stx_v0.ConfirmOut_Z(&info, outz); e == nil {
		dout = &txtool.TDOut{}
		dout.Asset = assets.NewAsset(
			&assets.Token{
				info.Tkn_currency,
				utils.NewU256_ByKey(&info.Tkn_value),
			},
			&assets.Ticket{
				info.Tkt_category,
				info.Tkt_value,
			},
		)
		dout.Memo = info.Memo
	}
	return
}
