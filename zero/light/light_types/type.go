package light_types

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type Kr struct {
	SKr keys.PKr
	PKr keys.PKr
}

type Out struct {
	Root  keys.Uint256
	State localdb.RootState
}

type DOut struct {
	Asset assets.Asset
	Memo  keys.Uint512
	Nil   keys.Uint256
}

type Block struct {
	Num  hexutil.Uint64
	Hash keys.Uint256
	Outs []Out
	Nils []keys.Uint256
}

type Witness struct {
	Pos    hexutil.Uint64
	Paths  [cpt.DEPTH]keys.Uint256
	Anchor keys.Uint256
}

type Tx struct {
	Hash keys.Uint256
	Tx   stx.T
}
