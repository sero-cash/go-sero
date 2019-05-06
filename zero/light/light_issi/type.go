package light_issi

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/light/light_types"
)

type Out struct {
	Root keys.Uint256
	PKr  keys.PKr
}
type Block struct {
	Num  hexutil.Uint64
	Outs []Out
	Nils []keys.Uint256
}

type GIn struct {
	SKr  keys.PKr
	Root keys.Uint256
}

type GenTxParam struct {
	Gas      uint64
	GasPrice uint64
	From     light_types.Kr
	Ins      []GIn
	Outs     []light_types.GOut
}

type ISSI interface {
	GetBlocksInfo(start uint64, count uint64) ([]Block, error)
	Detail(root []keys.Uint256, skr *keys.PKr) ([]light_types.DOut, error)
	GenTx(param GenTxParam) (keys.Uint256, error)
	CommitTx(txhash keys.Uint256) error
}
