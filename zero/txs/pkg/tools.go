package pkg

import "github.com/sero-cash/go-czero-import/keys"

func Pack(key *Key, pkg *Pkg_O) (ret Pkg_Z) {
	ret.Temp = pkg.Clone()
	return
}

func Check(k0 *keys.Uint256, pkg *Pkg_Z) (ret Pkg_O, e error) {
	ret = pkg.Temp.Clone()
	return
}

func Verify(k1 *keys.Uint256, opkg *Pkg_O, zpkg *Pkg_Z) (e error) {
	return
}
