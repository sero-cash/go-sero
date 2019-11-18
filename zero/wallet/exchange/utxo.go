package exchange

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

type Utxo struct {
	Pkr    c_type.PKr
	Root   c_type.Uint256
	TxHash c_type.Uint256
	Nil    c_type.Uint256
	Num    uint64
	Asset  assets.Asset
	IsZ    bool
	Ignore bool
	flag   int
}

func (utxo *Utxo) NilTxType() string {
	if c_superzk.IsSzkNil(&utxo.Nil) {
		return "SZK"
	} else {
		return "CZERO"
	}
}
