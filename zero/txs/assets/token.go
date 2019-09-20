package assets

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Token struct {
	Currency c_type.Uint256
	Value    utils.U256
}

func (self *Token) ToTypeAsset() c_type.Asset {
	return c_type.Asset{
		Tkn_currency: self.Currency,
		Tkn_value:    self.Value.ToUint256(),
		Tkt_category: c_type.Empty_Uint256,
		Tkt_value:    c_type.Empty_Uint256,
	}
}

func (self *Token) Clone() (ret Token) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Token) ToRef() (ret *Token) {
	ret = &this
	return
}

func (self *Token) ToHash() (ret c_type.Uint256) {
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
