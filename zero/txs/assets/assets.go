package assets

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Token struct {
	Currency keys.Uint256
	Value    utils.U256
}

func (self *Token) ToHash() (ret keys.Uint256) {
	if self == nil {
		return
	} else {
		hash := crypto.Keccak256(
			self.Currency[:],
			self.Value.ToUint256().NewRef()[:],
		)
		copy(ret[:], hash)
		return
	}
}

type Ticket struct {
	Value keys.Uint256
}

func (self *Ticket) ToHash() (ret keys.Uint256) {
	if self == nil {
		return
	} else {
		hash := crypto.Keccak256(
			self.Value[:],
		)
		copy(ret[:], hash)
		return
	}
}

type Package struct {
	Tkn *Token  `rlp:"nil"`
	Tkt *Ticket `rlp:"nil"`
}

func (self *Package) ToHash() (ret keys.Uint256) {
	hash := crypto.Keccak256(
		self.Tkn.ToHash().NewRef()[:],
		self.Tkt.ToHash().NewRef()[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *Package) Clone() (ret Package) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Package) ToRef() (ret *Package) {
	ret = &Package{Tkn: this.Tkn, Tkt: this.Tkt}
	return
}

func NewPackageByToken(tkn *Token) (ret Package) {
	ret.Tkn = tkn
	ret.Tkt = nil
	return
}
