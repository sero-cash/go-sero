package stx

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/utils"
)

type PkgOpen_Z struct {
	Id uint64
	K0 keys.Uint256
}

type PkgDesc_Z struct {
	Pack *pkg.Pkg_Z `rlp:"nil"`
	Open *PkgOpen_Z `rlp:"nil"`
}

func (this PkgDesc_Z) ToRef() (ret *PkgDesc_Z) {
	ret = &this
	return
}

func (self *PkgDesc_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	if self.Pack != nil {
		d.Write(self.Pack.ToHash().NewRef()[:])
	}
	if self.Open != nil {
		d.Write(big.NewInt(int64(self.Open.Id)).Bytes())
		d.Write(self.Open.K0[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgDesc_Z) Clone() (ret PkgDesc_Z) {
	utils.DeepCopy(&ret, self)
	return
}
