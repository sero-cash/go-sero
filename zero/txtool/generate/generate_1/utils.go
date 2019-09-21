package generate_1

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func GenAssetCC(a *assets.Asset) (ret c_superzk.AssetDesc) {
	ret = c_superzk.AssetDesc{
		Asset: a.ToTypeAsset(),
	}
	c_superzk.GenAssetCC(&ret)
	return
}

func GenTokenCC(t *assets.Token) (ret c_superzk.AssetDesc) {
	ret = c_superzk.AssetDesc{
		Asset: t.ToTypeAsset(),
	}
	c_superzk.GenAssetCC(&ret)
	return
}

func ConfirmOutC(key *c_type.Uint256, outc *stx_v1.Out_C) (dout *txtool.TDOut) {
	info := c_superzk.DecInfoDesc{}
	info.Key = *key
	info.Einfo = outc.EInfo
	c_superzk.DecOutput(&info)
	asset_desc := c_superzk.AssetDesc{}
	asset_desc.Asset = info.Asset_ret
	asset_desc.Ar = info.Ar_ret
	c_superzk.GenAssetCM(&asset_desc)
	if asset_desc.Asset_cm_ret == outc.AssetCM {
		dout = &txtool.TDOut{}
		dout.Asset = assets.NewAssetByType(&info.Asset_ret)
		dout.Memo = info.Memo_ret
		return
	} else {
		return
	}
}
