package prepare

import (
	"errors"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type Reception struct {
	Addr  keys.PKr
	Asset assets.Asset
}

type PkgCloseCmd struct {
	Id  keys.Uint256
	Key keys.Uint256
}

func (self *PkgCloseCmd) Asset() (ret assets.Asset, e error) {
	if p := txtool.Ref_inst.GetState().Pkgs.GetPkgById(&self.Id); p == nil {
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
	Id  keys.Uint256
	PKr keys.PKr
}

type PkgCreateCmd struct {
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
	PkgCreate   *PkgCreateCmd
	PkgTransfer *PkgTransferCmd
	PkgClose    *PkgCloseCmd
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
	From       keys.Uint512
	RefundTo   *keys.PKr
	Receptions []Reception
	Cmds       Cmds
	Fee        assets.Token
	GasPrice   *big.Int
	Roots      []keys.Uint256
}

type Utxo struct {
	Root  keys.Uint256
	Asset assets.Asset
}

type Utxos []Utxo

func (self *Utxos) Roots() (roots []keys.Uint256) {
	for _, utxo := range *self {
		roots = append(roots, utxo.Root)
	}
	return
}

type TxParamGenerator interface {
	FindRoots(pk *keys.Uint512, currency string, amount *big.Int) (utxos Utxos, remain big.Int)
	FindRootsByTicket(pk *keys.Uint512, tickets map[keys.Uint256]keys.Uint256) (roots Utxos, remain map[keys.Uint256]keys.Uint256)
	GetRoot(root *keys.Uint256) (utxos *Utxo)
	DefaultRefundTo(from *keys.Uint512) (ret *keys.PKr)
}
