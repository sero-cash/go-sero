package txtool

import (
	"math/big"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type GIn struct {
	SKr     keys.PKr
	Out     Out
	Witness Witness
}

type GOut struct {
	PKr   keys.PKr
	Asset assets.Asset
	Memo  keys.Uint512
}

type GTx struct {
	Gas      hexutil.Uint64
	GasPrice hexutil.Big
	Tx       stx.T
	Hash     keys.Uint256
	Roots    []keys.Uint256
	Keys     []keys.Uint256
}

type GPkgCloseCmd struct {
	Id      keys.Uint256
	Owner   keys.PKr
	AssetCM keys.Uint256
	Ar      keys.Uint256
}

type GPkgTransferCmd struct {
	Id    keys.Uint256
	Owner keys.PKr
	PKr   keys.PKr
}

type GPkgCreateCmd struct {
	Id    keys.Uint256
	PKr   keys.PKr
	Asset assets.Asset
	Memo  keys.Uint512
}

type Cmds struct {
	//Share
	BuyShare *stx.BuyShareCmd
	//Pool
	RegistPool *stx.RegistPoolCmd
	ClosePool  *stx.ClosePoolCmd
	//Contract
	Contract *stx.ContractCmd
	//Package
	PkgCreate   *GPkgCreateCmd
	PkgTransfer *GPkgTransferCmd
	PkgClose    *GPkgCloseCmd
}

type GTxParam struct {
	Gas      uint64
	GasPrice *big.Int
	Fee      assets.Token
	From     Kr
	Ins      []GIn
	Outs     []GOut
	Cmds     Cmds
}
