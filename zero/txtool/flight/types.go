package flight

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type PreTxParam struct {
	Gas      uint64
	GasPrice uint64
	From     c_type.PKr
	Ins      []c_type.Uint256
	Outs     []txtool.GOut
}
