package flight

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type PreTxParam struct {
	Gas      uint64
	GasPrice uint64
	From     keys.PKr
	Ins      []keys.Uint256
	Outs     []txtool.GOut
}
