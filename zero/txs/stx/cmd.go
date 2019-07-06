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

func (self *BuyShareCmd) Asset() (ret assets.Asset) {
	ret.Tkn = &assets.Token{
		utils.CurrencyToUint256("SERO"),
		self.Value,
	}
	return
}

type RegistPoolCmd struct {
	Value   utils.U256
	Vote    keys.PKr
	FeeRate uint32
}

func (self *RegistPoolCmd) Asset() (ret assets.Asset) {
	ret.Tkn = &assets.Token{
		utils.CurrencyToUint256("SERO"),
		self.Value,
	}
	return
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
	count := 0
	if self.BuyShare != nil {
		count++
	}
	if self.RegistPool != nil {
		count++
	}
	if self.ClosePool != nil {
		count++
	}
	if self.Contract != nil {
		count++
	}
	if count <= 1 {
		return true
	} else {
		return false
	}
}
