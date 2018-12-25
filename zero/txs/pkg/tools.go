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
