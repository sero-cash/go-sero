package pkg

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

func DePkg(key *keys.Uint256, pkg *Pkg_Z) (ret Pkg_O, e error) {
	desc := cpt.InfoDesc{}
	desc.Key = *key
	desc.Flag = true
	desc.Einfo = pkg.EInfo
	cpt.DecOutput(&desc)
	ret.Memo = desc.Memo
	if desc.Tkn_currency != keys.Empty_Uint256 {
		ret.Asset.Tkn = &assets.Token{
			desc.Tkn_currency,
			utils.NewU256_ByKey(&desc.Tkn_value),
		}
	}
	if desc.Tkt_category != keys.Empty_Uint256 {
		ret.Asset.Tkt = &assets.Ticket{
			desc.Tkt_category,
			desc.Tkt_value,
		}
	}
	ret.Ar = desc.Rsk
	return
}

func GetKey(pkr *keys.PKr, tk *keys.Uint512) (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(pkr[:])
	d.Write(tk[:])
	copy(ret[:], d.Sum(nil))
	return
}

func ConfirmPkg(o *Pkg_O, z *Pkg_Z) (e error) {
	asset := o.Asset.ToFlatAsset()
	desc := cpt.ConfirmPkgDesc{}
	desc.Tkn_currency = asset.Tkn.Currency
	desc.Tkn_value = asset.Tkn.Value.ToUint256()
	desc.Tkt_category = asset.Tkt.Category
	desc.Tkt_value = asset.Tkt.Value
	desc.Memo = o.Memo
	desc.Ar_ret = o.Ar
	desc.Pkg_cm = z.PkgCM
	e = cpt.ConfirmPkg(&desc)
	return
}
