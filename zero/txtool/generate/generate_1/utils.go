package generate_1

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
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
