package stx

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/utils"
)

type PkgClose struct {
	Id   keys.Uint256
	Sign keys.Uint512
}

func (this PkgClose) ToRef() (ret *PkgClose) {
	ret = &this
	return
}

func (self *PkgClose) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.Sign[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgClose) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgClose) ToHash_for_sign() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgClose) Clone() (ret PkgClose) {
	utils.DeepCopy(&ret, self)
	return
}

type PkgTransfer struct {
	Id   keys.Uint256
	PKr  keys.PKr
	Sign keys.Uint512
}

func (this PkgTransfer) ToRef() (ret *PkgTransfer) {
	ret = &this
	return
}

func (self *PkgTransfer) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.PKr[:])
	d.Write(self.Sign[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgTransfer) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.PKr[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgTransfer) ToHash_for_sign() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.PKr[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgTransfer) Clone() (ret PkgTransfer) {
	utils.DeepCopy(&ret, self)
	return
}

type PkgCreate struct {
	Id  keys.Uint256
	PKr keys.PKr
	Pkg pkg.Pkg_Z
}

func (this PkgCreate) ToRef() (ret *PkgCreate) {
	ret = &this
	return
}

func (self *PkgCreate) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.PKr[:])
	d.Write(self.Pkg.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgCreate) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.PKr[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgCreate) ToHash_for_sign() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Id[:])
	d.Write(self.PKr[:])
	d.Write(self.Pkg.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgCreate) Clone() (ret PkgCreate) {
	utils.DeepCopy(&ret, self)
	return
}

type PkgDesc_Z struct {
	Create   *PkgCreate   `rlp:"nil"`
	Transfer *PkgTransfer `rlp:"nil"`
	Close    *PkgClose    `rlp:"nil"`
}

func (this PkgDesc_Z) ToRef() (ret *PkgDesc_Z) {
	ret = &this
	return
}

func (self *PkgDesc_Z) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	if self.Create != nil {
		d.Write(self.Create.ToHash().NewRef()[:])
	}
	if self.Transfer != nil {
		d.Write(self.Transfer.ToHash().NewRef()[:])
	}
	if self.Close != nil {
		d.Write(self.Close.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgDesc_Z) ToHash_for_gen() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	if self.Create != nil {
		d.Write(self.Create.ToHash_for_gen().NewRef()[:])
	}
	if self.Transfer != nil {
		d.Write(self.Transfer.ToHash_for_gen().NewRef()[:])
	}
	if self.Close != nil {
		d.Write(self.Close.ToHash_for_gen().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgDesc_Z) ToHash_for_sign() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	if self.Create != nil {
		d.Write(self.Create.ToHash_for_sign().NewRef()[:])
	}
	if self.Transfer != nil {
		d.Write(self.Transfer.ToHash_for_sign().NewRef()[:])
	}
	if self.Close != nil {
		d.Write(self.Close.ToHash_for_sign().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *PkgDesc_Z) Clone() (ret PkgDesc_Z) {
	utils.DeepCopy(&ret, self)
	return
}
