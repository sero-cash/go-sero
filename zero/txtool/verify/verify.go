package verify

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txtool/verify/verify_1"
)

func VerifyWithoutState(ehash *c_type.Uint256, tx *stx.T, num uint64) (e error) {
	if num >= seroparam.SIP5() {
		return verify_1.VerifyWithoutState(ehash, tx)
	} else {
		return fmt.Errorf("VerifyWithoutState Error: verify_0 no longer be used")
		//return verify_0.VerifyWithoutState(ehash, tx, num)
	}
}

func VerifyWithState(tx *stx.T, state *zstate.ZState, num uint64) (e error) {
	if num >= seroparam.SIP5() {
		return verify_1.VerifyWithState(tx, state)
	} else {
		return fmt.Errorf("VerifyWithState Error: verify_0 no longer be used")
		//return verify_0.VerifyWithState(tx, state)
	}
}
