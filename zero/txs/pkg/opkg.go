package pkg

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Pkg_O struct {
	PKr   keys.Uint512
	Asset assets.Asset
}

func (this Pkg_O) ToRef() (ret *Pkg_O) {
	ret = &this
	return
}

func (self *Pkg_O) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.PKr[:])
	d.Write(self.Asset.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Pkg_O) Clone() (ret Pkg_O) {
	utils.DeepCopy(&ret, self)
	return
}
