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

func (self *Token) Clone() (ret Token) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Token) ToRef() (ret *Token) {
	ret = &this
	return
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
	Category keys.Uint256
	Value    keys.Uint256
}

func (self *Ticket) Clone() (ret Ticket) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Ticket) ToRef() (ret *Ticket) {
	ret = &this
	return
}

func (self *Ticket) ToHash() (ret keys.Uint256) {
	if self == nil {
		return
	} else {
		hash := crypto.Keccak256(
			self.Category[:],
			self.Value[:],
		)
		copy(ret[:], hash)
		return
	}
}

type Asset struct {
	Tkn *Token  `rlp:"nil"`
	Tkt *Ticket `rlp:"nil"`
}

type CompleteAsset struct {
	Tkn Token
	Tkt Ticket
}

func (self *Asset) ToCompleteAsset() (ret CompleteAsset) {
	if self.Tkt != nil {
		ret.Tkt = *self.Tkt
	}
	if self.Tkn != nil {
		ret.Tkn = *self.Tkn
	}
	return
}

func (self *Asset) GetToken() (ret Token) {
	if self.Tkn != nil {
		ret = *self.Tkn
		return
	} else {
		return
	}
}

func (self *Asset) GetTicket() (ret Ticket) {
	if self.Tkt != nil {
		ret = *self.Tkt
		return
	} else {
		return
	}
}

func (self *Asset) ToHash() (ret keys.Uint256) {
	hash := crypto.Keccak256(
		self.Tkn.ToHash().NewRef()[:],
		self.Tkt.ToHash().NewRef()[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *Asset) Clone() (ret Asset) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Asset) ToRef() (ret *Asset) {
	ret = &Asset{Tkn: this.Tkn, Tkt: this.Tkt}
	return
}

func NewPackageByToken(tkn *Token) (ret Asset) {
	ret.Tkn = tkn
	ret.Tkt = nil
	return
}

func NewAsset(tkn *Token, tkt *Ticket) (ret Asset) {
	if tkn != nil {
		if tkn.Value.Cmp(&utils.U256_0) > 0 {
			ret.Tkn = tkn.Clone().ToRef()
		}
	}
	if tkt != nil {
		if tkt.Value != keys.Empty_Uint256 {
			ret.Tkt = tkt.Clone().ToRef()
		}
	}
	return
}
