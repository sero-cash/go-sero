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
