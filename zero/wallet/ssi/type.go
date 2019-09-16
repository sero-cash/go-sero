package ssi

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type Out struct {
	Root c_type.Uint256
	Hash c_type.Uint256
	PKr  c_type.PKr
}
type Block struct {
	Num  hexutil.Uint64
	Hash c_type.Uint256
	Outs []Out
	Nils []c_type.Uint256
}

type GIn struct {
	SKr  c_type.PKr
	Root c_type.Uint256
}

type PreTxParam struct {
	Gas      uint64
	GasPrice uint64
	From     txtool.Kr
	Ins      []GIn
	Outs     []txtool.GOut
}

type ISSI interface {
	GetBlocksInfo(start uint64, count uint64) ([]Block, error)
	Detail(root []c_type.Uint256, skr *c_type.PKr) ([]txtool.DOut, error)
	GenTx(param *PreTxParam) (c_type.Uint256, error)
	CommitTx(txhash *c_type.Uint256) error
}
