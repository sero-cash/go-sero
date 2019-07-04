package ssi

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type Out struct {
	Root keys.Uint256
	Hash keys.Uint256
	PKr  keys.PKr
}
type Block struct {
	Num  hexutil.Uint64
	Hash keys.Uint256
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
	From     txtool.Kr
	Ins      []GIn
	Outs     []txtool.GOut
}

type ISSI interface {
	GetBlocksInfo(start uint64, count uint64) ([]Block, error)
	Detail(root []keys.Uint256, skr *keys.PKr) ([]txtool.DOut, error)
	//GenTx(param *GenTxParam) (keys.Uint256, error)
	GenTx(param *GenTxParam) (keys.Uint256, error)
	CommitTx(txhash *keys.Uint256) error
}
