package light

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/light/light_issi"
	"github.com/sero-cash/go-sero/zero/light/light_types"
)

type SSI struct {
}

var SSI_Inst = SSI{}

func (self *SSI) GetBlocksInfo(start uint64, count uint64) (blocks []light_issi.Block, e error) {

	return
}

func (self *SSI) Detail(root []keys.Uint256, skr *keys.PKr) (douts []light_types.DOut, e error) {

	return
}

func (self *SSI) GenTx(param light_issi.GenTxParam) (hash keys.Uint256, e error) {

	return
}

func (self *SSI) CommitTx(txhash keys.Uint256) (e error) {

	return
}
