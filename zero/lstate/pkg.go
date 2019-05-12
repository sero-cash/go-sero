package lstate

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate/pkgstate"
)

type Pkg struct {
	Pkg pkgstate.OPkg
	Key keys.Uint256
}
