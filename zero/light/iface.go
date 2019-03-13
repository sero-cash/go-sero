package light

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type GIn struct {
	SKr     keys.PKr
	Out     Out
	Witness Witness
}

type GOut struct {
	PKr   keys.PKr
	Asset assets.Asset
	Memo  keys.Uint512
}

type GTx struct {
	Gas      uint64
	GasPrice big.Int
	Tx       stx.T
}

type GenTxParam struct {
	Gas      uint64
	GasPrice big.Int
	Ins      []GIn
	Outs     []GOut
}

type ISLI interface {
	CreateKr() Kr
	DecOuts(outs []Out, skr *keys.PKr) ([]DOut, error)
	GenTx(param *GenTxParam) (GTx, error)
}

type SRI interface {
	GetBlocksInfo(start uint64, count uint64) ([]Block, error)
	GetAnchor(roots []keys.Uint256) ([]Witness, error)
	CommitTx(tx *GTx) error
}
