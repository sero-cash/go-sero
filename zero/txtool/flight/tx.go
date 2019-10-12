package flight

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_0"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_1"
)

func SignTx0(param *txtool.GTxParam) (gtx txtool.GTx, e error) {
	if tx, keys, bases, err := generate_0.GenTx(param); err != nil {
		e = err
		return
	} else {
		gtx.Tx = tx
		gtx.Keys = keys
		gtx.Bases = bases
		gtx.Gas = hexutil.Uint64(param.Gas)
		gtx.GasPrice = hexutil.Big(*param.GasPrice)
		gtx.Hash = gtx.Tx.ToHash()
		return
	}
}

func SignTx1(txParam *txtool.GTxParam) (tx stx.T, param txtool.GTxParam, keys []c_type.Uint256, e error) {
	if ctx, err := generate_1.SignTx(txParam); err != nil {
		e = err
		return
	} else {
		tx = ctx.Tx()
		param = ctx.Param()
		keys = ctx.Keys()
		return
	}
}

func ProveTx1(tx *stx.T, param *txtool.GTxParam) (gtx txtool.GTx, e error) {
	if ctx, err := generate_1.ProveTx(tx, param); err != nil {
		e = err
		return
	} else {
		gtx.Tx = ctx.Tx()
		gtx.Gas = hexutil.Uint64(param.Gas)
		gtx.GasPrice = hexutil.Big(*param.GasPrice)
		gtx.Hash = gtx.Tx.ToHash()
		return

	}
}
