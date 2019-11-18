package flight

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_1"
)

func SignTx0(param *txtool.GTxParam) (gtx txtool.GTx, e error) {
	e = fmt.Errorf("SignTx0 Error: signTx0 not support after sip5")
	return
}

func SignTx1(txParam *txtool.GTxParam) (tx stx.T, param txtool.GTxParam, keys []c_type.Uint256, bases []c_type.Uint256, e error) {
	if ctx, err := generate_1.SignTx(txParam); err != nil {
		e = err
		return
	} else {
		tx = ctx.Tx()
		param = ctx.Param()
		keys = ctx.Keys()
		bases = ctx.Bases()
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
