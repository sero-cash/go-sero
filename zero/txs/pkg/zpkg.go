package pkg

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/rlp"
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

func (self *Pkg_Z) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}

type Pkg_ZGet struct {
	out *Pkg_Z
}

func (self *Pkg_ZGet) Out() *Pkg_Z {
	return self.out
}

func (self *Pkg_ZGet) Unserial(v []byte) (e error) {
	if len(v) < 2 {
		self.out = nil
		return
	} else {
		self.out = &Pkg_Z{}
		if err := rlp.DecodeBytes(v, &self.out); err != nil {
			e = err
			self.out = nil
			return
		} else {
			return
		}
	}
}
