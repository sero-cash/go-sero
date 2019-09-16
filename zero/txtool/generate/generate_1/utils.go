package generate_1

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func ConfirmOutC(key *c_type.Uint256, outc *stx_v1.Out_C) (dout *txtool.TDOut, ar c_type.Uint256) {
	info := c_superzk.DecInfoDesc{}
	info.Key = *key
	info.Einfo = outc.EInfo
	c_superzk.DecOutput(&info)
	asset_desc := c_superzk.AssetDesc{}
	asset_desc.Asset = info.Asset_ret
	asset_desc.Ar = info.Ar_ret
	ar = asset_desc.Ar
	if e := c_superzk.GenAssetCM(&asset_desc); e != nil {
		return
	}
	if asset_desc.Asset_cm_ret == outc.AssetCM {
		dout = &txtool.TDOut{}
		dout.Asset = assets.NewAssetByType(&info.Asset_ret)
		dout.Memo = info.Memo_ret
		return
	} else {
		return
	}
}

func ConfirmOutZ(key *c_type.Uint256, flag bool, outz *stx_v0.Out_Z) (dout *txtool.TDOut) {
	if asset, rsk, memo, e := c_superzk.Czero_decEInfo(key, flag, &outz.EInfo); e != nil {
		return
	} else {
		if out_cm, err := c_superzk.Czero_genOutCM(&asset, &memo, &outz.PKr, &rsk); err != nil {
			return
		} else {
			if out_cm != outz.OutCM {
				return
			} else {
				dout = &txtool.TDOut{}
				dout.Asset = assets.NewAssetByType(&asset)
				dout.Memo = memo
				return
			}
		}
	}
}
