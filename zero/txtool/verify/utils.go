package verify

import (
	"fmt"

	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

func CheckUint(i *utils.U256) bool {
	if len(i.ToInt().Bytes()) > 32 {
		return false
	}
	u := i.ToUint256()
	if u[31] == 0 && u[30] == 0 {
		return true
	} else {
		return false
	}
}
func ReportError(str string, tx *stx.T) (e error) {
	h := hexutil.Encode(tx.ToHash().NewRef()[:])
	log.Error("Verify Tx Error", "reason", str, "hash", h)
	return fmt.Errorf("Verify Tx Error: resean=%v , hash=%v", str, h)
}
