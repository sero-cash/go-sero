package pkg

import "github.com/sero-cash/go-czero-import/keys"

func EnPkg(key *keys.Uint256, pkg *Pkg_O) (ret Pkg_Z) {
	ret.Temp = pkg.Clone()
	return
}

func DePkg(key *keys.Uint256, pkg *Pkg_Z) (ret Pkg_O, e error) {
	ret = pkg.Temp.Clone()
	return
}
