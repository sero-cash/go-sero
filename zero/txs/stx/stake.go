package stx

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
)

type BuyShare struct {
	Value utils.U256
	Price utils.U256
	Vote  keys.PKr
	Pool  keys.Uint256
}

type RegistPool struct {
	Id      keys.Uint256
	Vote    keys.PKr
	FeeRate uint32
}

type StakeDesc struct {
	BuyShare   *BuyShare
	RegistPool *RegistPool
}
