package pkg

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Key struct {
	K0 keys.Uint256
	K1 keys.Uint256
}

func (this Key) ToRef() (ret *Key) {
	ret = &this
	return
}

func (self *Key) Clone() (ret Key) {
	utils.DeepCopy(&ret, self)
	return
}
