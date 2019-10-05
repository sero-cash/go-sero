package stx_v0

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/utils"
)

type In_Z struct {
	Anchor  c_type.Uint256
	Nil     c_type.Uint256
	Trace   c_type.Uint256
	AssetCM c_type.Uint256
	Proof   c_type.Proof
}

func (self *In_Z) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Anchor[:])
	d.Write(self.Nil[:])
	d.Write(self.Trace[:])
	d.Write(self.AssetCM[:])
	d.Write(self.Proof.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

func (self *In_Z) ToHash_for_sign() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Anchor[:])
	d.Write(self.Nil[:])
	d.Write(self.Trace[:])
	d.Write(self.AssetCM[:])
	d.Write(self.Proof.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

type Out_Z struct {
	AssetCM c_type.Uint256
	OutCM   c_type.Uint256
	RPK     c_type.Uint256
	EInfo   c_type.Einfo
	PKr     c_type.PKr
	Proof   c_type.Proof
}

func (self *Out_Z) Clone() (ret Out_Z) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Out_Z) ToRef() (ret *Out_Z) {
	ret = &Out_Z{}
	*ret = this
	return
}

func (self *Out_Z) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.AssetCM[:])
	d.Write(self.OutCM[:])
	d.Write(self.EInfo[:])
	d.Write(self.PKr[:])
	d.Write(self.Proof.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Out_Z) ToHash_for_gen() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.PKr[:])
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Out_Z) ToHash_for_sign() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.AssetCM[:])
	d.Write(self.OutCM[:])
	d.Write(self.EInfo[:])
	d.Write(self.PKr[:])
	d.Write(self.Proof.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return
}

type Desc_Z struct {
	Ins  []In_Z
	Outs []Out_Z
}

func (self *Desc_Z) HasContent() bool {
	return len(self.Ins) > 0 || len(self.Outs) > 0
}

func (self *Desc_Z) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Desc_Z) ToHash_for_gen() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, out := range self.Outs {
		d.Write(out.ToHash_for_gen().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Desc_Z) ToHash_for_sign() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash_for_sign().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash_for_sign().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}
