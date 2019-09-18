package pkg

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Pkg_Z struct {
	AssetCM c_type.Uint256
	PkgCM   c_type.Uint256
	EInfo   c_type.Einfo
}

func (this Pkg_Z) ToRef() (ret *Pkg_Z) {
	ret = &this
	return
}

func (self *Pkg_Z) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.AssetCM[:])
	d.Write(self.PkgCM[:])
	d.Write(self.EInfo[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Pkg_Z) Clone() (ret Pkg_Z) {
	utils.DeepCopy(&ret, self)
	return
}
