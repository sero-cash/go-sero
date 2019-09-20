package prepare

import (
	"errors"
	"math/big"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-sero/zero/txtool/flight"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type Reception struct {
	Addr  c_type.PKr
	Asset assets.Asset
}

type PkgCloseCmd struct {
	Id  c_type.Uint256
	Key c_type.Uint256
}

func (self *PkgCloseCmd) Asset() (ret assets.Asset, e error) {
	if p := txtool.Ref_inst.CurrentState().Pkgs.GetPkgById(&self.Id); p == nil {
		e = errors.New("close pkg but not find the pkg")
		return
	} else {
		if opkg, err := pkg.DePkg(&self.Key, &p.Pack.Pkg); err != nil {
			e = err
			return
		} else {
			ret = opkg.Asset
			return
		}
	}
}

type PkgTransferCmd struct {
	Id  c_type.Uint256
	PKr c_type.PKr
}

type PkgCreateCmd struct {
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
	PkgCreate   *PkgCreateCmd
	PkgTransfer *PkgTransferCmd
	PkgClose    *PkgCloseCmd
}

func (self *Cmds) ToPkr() (ret *c_type.PKr) {
	if self.BuyShare != nil {
		return &self.BuyShare.Vote
	}
	if self.RegistPool != nil {
		return &self.RegistPool.Vote
	}
	if self.PkgCreate != nil {
		if self.PkgCreate.PKr.IsEndEmpty() {
			return
		} else {
			return &self.PkgCreate.PKr
		}
	}
	if self.PkgTransfer != nil {
		if self.PkgTransfer.PKr.IsEndEmpty() {
			return
		} else {
			return &self.PkgTransfer.PKr
		}
	}
	return
}

func (self *Cmds) InAsset() (asset *assets.Asset, e error) {
	if self.PkgClose != nil {
		if a, err := self.PkgClose.Asset(); err != nil {
			e = err
			return
		} else {
			asset = &a
			return
		}
	} else {
		return
	}
}

func (self *Cmds) OutAsset() *assets.Asset {
	if self.PkgCreate != nil {
		return &self.PkgCreate.Asset
	}
	if self.BuyShare != nil {
		asset := self.BuyShare.Asset()
		return &asset
	}
	if self.RegistPool != nil {
		asset := self.RegistPool.Asset()
		return &asset
	}
	if self.Contract != nil {
		return &self.Contract.Asset
	}
	return nil
}

func (self *Cmds) Valid() bool {
	count := 0
	if self.PkgCreate != nil {
		count++
	}
	if self.PkgTransfer != nil {
		count++
	}
	if self.PkgClose != nil {
		count++
	}
	if self.BuyShare != nil {
		count++
	}
	if self.RegistPool != nil {
		count++
	}
	if self.ClosePool != nil {
		count++
	}
	if self.Contract != nil {
		count++
	}
	if count <= 1 {
		return true
	} else {
		return false
	}
}

type PreTxParam struct {
	From       c_type.Uint512
	RefundTo   *c_type.PKr
	Receptions []Reception
	Cmds       Cmds
	Fee        assets.Token
	GasPrice   *big.Int
	Roots      []c_type.Uint256
}

type ADDRESS_VERSION int

const (
	AV_UNKNOW  = ADDRESS_VERSION(0)
	AV_CZERO   = ADDRESS_VERSION(1)
	AV_SUPERZK = ADDRESS_VERSION(2)
)

func (self *PreTxParam) IsSzk() (ret ADDRESS_VERSION, e error) {
	pkrs := []c_type.PKr{}
	if self.RefundTo != nil {
		pkrs = append(pkrs, *self.RefundTo)
	}
	for _, recept := range self.Receptions {
		pkrs = append(pkrs, recept.Addr)
	}
	if pkr := self.Cmds.ToPkr(); pkr != nil {
		pkrs = append(pkrs, *pkr)
	}
	for _, pkr := range pkrs {
		if ret != AV_UNKNOW {
			temp := AV_CZERO
			if c_superzk.IsSzkPKr(&pkr) {
				temp = AV_SUPERZK
			}
			if ret != temp {
				e = errors.New("the address mode is inconsistency")
				return
			}
		} else {
			if c_superzk.IsSzkPKr(&pkr) {
				ret = AV_SUPERZK
			} else {
				ret = AV_CZERO
			}
		}
	}
	return
}

type Utxo struct {
	Root  c_type.Uint256
	Asset assets.Asset
}

type Utxos []Utxo

func (self *Utxos) Roots() (roots []c_type.Uint256) {
	for _, utxo := range *self {
		roots = append(roots, utxo.Root)
	}
	return
}

type TxParamGenerator interface {
	FindRoots(pk *c_type.Uint512, currency string, amount *big.Int) (utxos Utxos, remain big.Int)
	FindRootsByTicket(pk *c_type.Uint512, tickets map[c_type.Uint256]c_type.Uint256) (roots Utxos, remain map[c_type.Uint256]c_type.Uint256)
	GetRoot(root *c_type.Uint256) (utxos *Utxo)
	DefaultRefundTo(from *c_type.Uint512, av ADDRESS_VERSION) (ret *c_type.PKr)
}

type TxParamState interface {
	GetAnchor(roots []c_type.Uint256) (wits []txtool.Witness, e error)
	GetOut(root *c_type.Uint256) (out *localdb.RootState)
	GetPkgById(id *c_type.Uint256) (ret *localdb.ZPkg)
	GetSeroGasLimit(to *common.Address, tfee *assets.Token, gasPrice *big.Int) (gaslimit uint64, e error)
}

type DefaultTxParamState struct {
}

func (self *DefaultTxParamState) GetAnchor(roots []c_type.Uint256) (wits []txtool.Witness, e error) {
	return flight.SRI_Inst.GetAnchor(roots)
}

func (self *DefaultTxParamState) GetOut(root *c_type.Uint256) (out *localdb.RootState) {
	return flight.GetOut(root, 0)
}

func (self *DefaultTxParamState) GetPkgById(id *c_type.Uint256) (ret *localdb.ZPkg) {
	return txtool.Ref_inst.CurrentState().Pkgs.GetPkgById(id)
}

func (self *DefaultTxParamState) GetSeroGasLimit(to *common.Address, tfee *assets.Token, gasPrice *big.Int) (gaslimit uint64, e error) {
	return txtool.Ref_inst.Bc.GetSeroGasLimit(to, tfee, gasPrice)
}
