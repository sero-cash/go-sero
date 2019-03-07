package light

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type Kr struct {
	PKr keys.PKr
	SKr keys.PKr
}

type Out struct {
	Root  keys.Uint256
	Index uint64
	Z_out *stx.Out_Z
	O_out *stx.Out_O
}

type DOut struct {
	Asset assets.Asset
	Memo  keys.Uint512
	Nil   keys.Uint256
}

type Block struct {
	Num  uint64
	Outs []Out
	Nils []keys.Uint256
}

type Witness struct {
	pos    uint64
	paths  [cpt.DEPTH]keys.Uint256
	Anchor keys.Uint256
}

type Tx struct {
	Hash keys.Uint256
	Tx   stx.T
}
