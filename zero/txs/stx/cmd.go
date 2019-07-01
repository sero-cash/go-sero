package stx

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type BuyShareCmd struct {
	asset assets.Asset
	Price utils.U256
	Vote  keys.PKr
	Pool  keys.Uint256
}

type RegistPoolCmd struct {
	asset   assets.Asset
	Id      keys.Uint256
	Vote    keys.PKr
	FeeRate uint32
}

type ContractCmd struct {
	asset assets.Asset
	To    keys.Uint512
	Data  []byte
}

type DescCmd struct {
	BuyShare   *BuyShareCmd
	RegistPool *RegistPoolCmd
	Contract   *ContractCmd
}

func (self *DescCmd) Valid() bool {
	if self.BuyShare == nil && self.RegistPool == nil && self.Contract == nil {
		return false
	}
	return true
}
