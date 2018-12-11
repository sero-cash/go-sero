package stx

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/utils"
)

type PkgOpen struct {
	Id   keys.Uint256
	Sign keys.Uint256
}

func (this PkgOpen) ToRef() (ret *PkgOpen) {
	ret = &this
	return
}

func (self *PkgOpen) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.Sign[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgOpen) Clone() (ret PkgOpen) {
	utils.DeepCopy(&ret, self)
	return
}

type PkgChange struct {
	Id  keys.Uint256
	Pkr keys.Uint512
}

func (this PkgChange) ToRef() (ret *PkgChange) {
	ret = &this
	return
}

func (self *PkgChange) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.Pkr[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgChange) Clone() (ret PkgChange) {
	utils.DeepCopy(&ret, self)
	return
}

type PkgPack struct {
	Id  keys.Uint256
	Pkr keys.Uint512
	Pkg pkg.Pkg_Z
}

func (this PkgPack) ToRef() (ret *PkgPack) {
	ret = &this
	return
}

func (self *PkgPack) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.Pkr[:])
	d.Write(self.Pkg.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgPack) Clone() (ret PkgPack) {
	utils.DeepCopy(&ret, self)
	return
}

type PkgDesc_Z struct {
	Pack   *PkgPack   `rlp:"nil"`
	Change *PkgChange `rlp:"nil"`
	Open   *PkgOpen   `rlp:"nil"`
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
	if self.Change != nil {
		d.Write(self.Change.ToHash().NewRef()[:])
	}
	if self.Open != nil {
		d.Write(self.Open.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgDesc_Z) Clone() (ret PkgDesc_Z) {
	utils.DeepCopy(&ret, self)
	return
}
