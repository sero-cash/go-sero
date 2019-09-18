package txtool

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type GIn struct {
	SKr     c_type.PKr
	Out     Out
	Witness Witness
	A       *c_type.Uint256
}

type GOut struct {
	PKr   c_type.PKr
	Asset assets.Asset
	Memo  c_type.Uint512
}

func (self *GOut) ToSzkEncInfoDesc() (ret c_superzk.EncInfoDesc) {
	ret.Memo = self.Memo
	type_asset := self.Asset.ToTypeAsset()
	ret.Asset = type_asset
	return
}

type GTx struct {
	Gas      hexutil.Uint64
	GasPrice hexutil.Big
	Tx       stx.T
	Hash     c_type.Uint256
	Roots    []c_type.Uint256
	Keys     []c_type.Uint256
	Bases    []c_type.Uint256
}

type GPkgCloseCmd struct {
	Id      c_type.Uint256
	Owner   c_type.PKr
	AssetCM c_type.Uint256
	Ar      c_type.Uint256
}

type GPkgTransferCmd struct {
	Id    c_type.Uint256
	Owner c_type.PKr
	PKr   c_type.PKr
}

type GPkgCreateCmd struct {
	Id    c_type.Uint256
	PKr   c_type.PKr
	Asset assets.Asset
	Memo  c_type.Uint512
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

func (self *GTxParam) IsSzk() bool {
	return superzk.IsSzkPKr(&self.From.PKr)
}
