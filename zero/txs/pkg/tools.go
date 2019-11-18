package pkg

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

func DePkg(key *c_type.Uint256, pkg *Pkg_Z) (ret Pkg_O, e error) {
	if asset, memo, ar, err := c_superzk.DecEInfo(key, &pkg.EInfo); err != nil {
		e = err
		return
	} else {
		ret.Memo = memo
		ret.Asset = assets.NewAssetByType(&asset)
		ret.Ar = ar
	}
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
	if cm, _, err := c_superzk.GenAssetCM_PC(o.Asset.ToTypeAsset().NewRef(), &o.Ar); err != nil {
		e = err
		return
	} else {
		if z.AssetCM != cm {
			e = errors.New("pkg asset_cm is not match")
			return
		} else {
			return
		}
	}
	return
}
