package verify

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txtool/verify/verify_0"
)

func VerifyWithoutState(ehash *c_type.Uint256, tx *stx.T, num uint64) (e error) {
	if tx.IsSzk() {
		//e = errors.New("szk is empty")
		return
	} else {
		return verify_0.VerifyWithoutState(ehash, tx, num)
	}
}

func VerifyWithState(tx *stx.T, state *zstate.ZState) (e error) {
	if tx.IsSzk() {
		//e = errors.New("szk is empty")
		return
	} else {
		return verify_0.VerifyWithState(tx, state)
	}
}
