package pkg

import (
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

func DePkg(key *c_type.Uint256, pkg *Pkg_Z) (ret Pkg_O, e error) {
	desc := c_czero.InfoDesc{}
	desc.Key = *key
	desc.Flag = true
	desc.Einfo = pkg.EInfo
	c_czero.DecOutput(&desc)
	ret.Memo = desc.Memo
	ret.Asset = assets.NewAssetByType(&desc.Asset)
	ret.Ar = desc.Rsk
	return
}

func GetKey(pkr *c_type.PKr, tk *c_type.Tk) (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(pkr[:])
	d.Write(tk[:])
	copy(ret[:], d.Sum(nil))
	return
}

func ConfirmPkg(o *Pkg_O, z *Pkg_Z) (e error) {
	asset := o.Asset.ToFlatAsset()
	desc := c_czero.ConfirmPkgDesc{}
	desc.Tkn_currency = asset.Tkn.Currency
	desc.Tkn_value = asset.Tkn.Value.ToUint256()
	desc.Tkt_category = asset.Tkt.Category
	desc.Tkt_value = asset.Tkt.Value
	desc.Memo = o.Memo
	desc.Ar_ret = o.Ar
	desc.Pkg_cm = z.PkgCM
	e = c_czero.ConfirmPkg(&desc)
	return
}
