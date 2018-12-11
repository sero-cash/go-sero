package pkg

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Pkg_Z struct {
	Temp Pkg_O
}

func (this Pkg_Z) ToRef() (ret *Pkg_Z) {
	ret = &this
	return
}

func (self *Pkg_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Temp.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Pkg_Z) Clone() (ret Pkg_Z) {
	utils.DeepCopy(&ret, self)
	return
}
