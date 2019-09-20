package generate_1

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func GenAssetCC(a *assets.Asset) (ret c_superzk.AssetDesc) {
	asset := a.ToFlatAsset()
	ret = c_superzk.AssetDesc{
		Asset: c_type.Asset{
			Tkn_currency: asset.Tkn.Currency,
			Tkn_value:    asset.Tkn.Value.ToUint256(),
			Tkt_category: asset.Tkt.Category,
			Tkt_value:    asset.Tkt.Value,
		},
	}
	ret = a.ToSzkAssetDesc()
	c_superzk.GenAssetCC(&ret)
	return
}

func GenTokenCC(t *assets.Token) (ret c_superzk.AssetDesc) {
	tkn := assets.Token{
		t.Currency,
		t.Value,
	}
	ret = c_superzk.AssetDesc{
		Asset: c_type.Asset{
			Tkn_currency: tkn.Currency,
			Tkn_value:    tkn.Value.ToUint256(),
			Tkt_category: c_type.Empty_Uint256,
			Tkt_value:    c_type.Empty_Uint256,
		},
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
		dout.Memo = info.Memo
		return
	} else {
		return
	}
}
