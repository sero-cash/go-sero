package assets

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/utils"
)

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
