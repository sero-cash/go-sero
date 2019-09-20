package stx_v1

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type In_P struct {
	Root  c_type.Uint256
	Nil   c_type.Uint256
	NSign c_type.Uint512
	ASign c_type.Uint512
}

func (self *In_P) Tx1_Hash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Root[:],
		self.Nil[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *In_P) ToHash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Root[:],
		self.Nil[:],
		self.NSign[:],
		self.ASign[:],
	)
	copy(ret[:], hash)
	return ret
}

type In_P0 struct {
	Root  c_type.Uint256
	Nil   c_type.Uint256
	Trace c_type.Uint256
	Key   *c_type.Uint256 `rlp:"nil"`
	Sign  c_type.Uint512
}

func (self *In_P0) Tx1_Hash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Root[:])
	d.Write(self.Nil[:])
	d.Write(self.Trace[:])
	if self.Key != nil {
		d.Write(self.Key[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *In_P0) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Root[:])
	d.Write(self.Nil[:])
	d.Write(self.Trace[:])
	if self.Key != nil {
		d.Write(self.Key[:])
	}
	d.Write(self.Sign[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

type Out_P struct {
	PKr   c_type.PKr
	Asset assets.Asset
	Memo  c_type.Uint512
}

func (self *Out_P) Clone() (ret Out_P) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Out_P) ToRef() (ret *Out_P) {
	ret = &this
	return
}
