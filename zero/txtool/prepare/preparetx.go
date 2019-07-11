package prepare

import (
	"bytes"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
)

func GenTxParam(param *PreTxParam, gen TxParamGenerator) (txParam *txtool.GTxParam, e error) {
	if len(param.Receptions) > 500 {
		return nil, errors.New("receptions count must <= 500")
	}
	utxos, err := SelectUtxos(param, gen)
	if err != nil {
		return nil, err
	}

	if param.RefundTo == nil {
		if param.RefundTo = gen.DefaultRefundTo(&param.From); param.RefundTo == nil {
			return nil, errors.New("can not find default refund to")
		}
	}
	txParam, e = BuildTxParam(&DefaultTxParamState{}, utxos, param.RefundTo, param.Receptions, &param.Cmds, &param.Fee, param.GasPrice)
	return
}

func IsPk(addr keys.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func CreatePkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(utils.EncodeNumber(index), 32))
	if index == 0 {
		return keys.Addr2PKr(pk, nil)
	} else {
		return keys.Addr2PKr(pk, &r)
	}
}
