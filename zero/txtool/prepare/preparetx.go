package prepare

import (
	"bytes"
	"encoding/binary"

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
	txParam, e = BuildTxParam(utxos, param.RefundTo, param.Receptions, &param.Cmds, &param.Fee, param.GasPrice)
	return
}

func IsPk(addr keys.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func CreatePkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(EncodeNumber(index), 32))
	if index == 0 {
		return keys.Addr2PKr(pk, nil)
	} else {
		return keys.Addr2PKr(pk, &r)
	}
}

func EncodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func DecodeNumber(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}
