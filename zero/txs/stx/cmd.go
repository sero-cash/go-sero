package stx

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type BuyShareCmd struct {
	Value utils.U256
	Price utils.U256
	Vote  keys.PKr
	Pool  keys.Uint256
}

type RegistPoolCmd struct {
	Value   utils.U256
	Vote    keys.PKr
	FeeRate uint32
}

type ClosePoolCmd struct{}

type ContractCmd struct {
	Asset assets.Asset
	To    keys.PKr
	Data  []byte
}

type DescCmd struct {
	BuyShare   *BuyShareCmd
	RegistPool *RegistPoolCmd
	ClosePool  *ClosePoolCmd
	Contract   *ContractCmd
}

func (self *DescCmd) Valid() bool {
	if self.BuyShare == nil && self.RegistPool == nil && self.Contract == nil {
		return false
	}
	return true
}
