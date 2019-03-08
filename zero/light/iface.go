package light

import (
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
	Tx stx.T
}

type GenTxParam struct {
	Ins  []GIn
	Outs []GOut
}

type SLI interface {
	CreateKr() Kr
	DecOuts(outs []Out, skr *keys.PKr) ([]DOut, error)
	GenTx(param *GenTxParam) (GTx, error)
}

type TxStatusFlag int

type TxStatus struct {
	Flag TxStatusFlag
	Info string
}

const (
	TX_PROC = TxStatusFlag(0)
	TX_SUCC = TxStatusFlag(1)
	TX_FAIL = TxStatusFlag(2)
)

type SRI interface {
	GetBlocksInfo(start int, count int) ([]Block, error)
	GetAnchor(roots []keys.Uint256) ([]Witness, error)
	CommitTx(tx *GTx) error
	GetTxStatus(tx []keys.Uint256) ([]TxStatus, error)
}
