package ethapi

import (
	"fmt"
	"math/big"

	"github.com/sero-cash/go-sero/accounts"

	"github.com/sero-cash/go-sero/zero/txs/stx"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txtool/prepare"
	"github.com/sero-cash/go-sero/zero/utils"
)

type PkgCloseArgs struct {
	Id  c_type.Uint256
	Key c_type.Uint256
}

func (self *PkgCloseArgs) toCmd() *prepare.PkgCloseCmd {
	if self == nil {
		return nil
	}
	return &prepare.PkgCloseCmd{
		self.Id,
		self.Key,
	}
}

type PkgTransferArgs struct {
	Id  c_type.Uint256
	PKr AllMixedAddress
}

func (self *PkgTransferArgs) toCmd() *prepare.PkgTransferCmd {
	if self == nil {
		return nil
	}
	return &prepare.PkgTransferCmd{
		self.Id,
		self.PKr.ToPKr(),
	}
}

type PkgCreateArgs struct {
	Id       c_type.Uint256
	PKr      AllMixedAddress
	Currency Smbol
	Value    *Big
	Memo     c_type.Uint512
}

func (self *PkgCreateArgs) toCmd() *prepare.PkgCreateCmd {
	if self == nil {
		return nil
	}
	asset := assets.Asset{}
	if !self.Currency.IsEmpty() && self.Value != nil {
		asset.Tkn = &assets.Token{
			utils.CurrencyToUint256(string(self.Currency)),
			utils.U256(*self.Value.ToInt()),
		}
	}
	return &prepare.PkgCreateCmd{
		self.Id,
		self.PKr.ToPKr(),
		asset,
		self.Memo,
	}
}

type BuyShareArgs struct {
	Value Big
	Vote  PKrAddress
	Pool  *c_type.Uint256
}

func (self *BuyShareArgs) toCmd() *stx.BuyShareCmd {
	if self == nil {
		return nil
	}
	return &stx.BuyShareCmd{
		utils.U256(*self.Value.ToInt()),
		*self.Vote.ToPKr(),
		self.Pool,
	}
}

type RegistPoolArgs struct {
	Value   utils.U256
	Vote    PKrAddress
	FeeRate uint32
}

func (self *RegistPoolArgs) toCmd() *stx.RegistPoolCmd {
	if self == nil {
		return nil
	}
	return &stx.RegistPoolCmd{
		utils.U256(*self.Value.ToInt()),
		*self.Vote.ToPKr(),
		self.FeeRate,
	}
}

type ClosePoolArgs struct {
}

func (self *ClosePoolArgs) toCmd() *stx.ClosePoolCmd {
	if self == nil {
		return nil
	}
	return &stx.ClosePoolCmd{}
}

type ContractArgs struct {
	Currency Smbol
	Value    *Big
	To       *ContractAddress
	Data     hexutil.Bytes
}

func (self *ContractArgs) toCmd() *stx.ContractCmd {
	if self == nil {
		return nil
	}
	asset := assets.Asset{}
	if !self.Currency.IsEmpty() && self.Value != nil {
		asset.Tkn = &assets.Token{
			utils.CurrencyToUint256(string(self.Currency)),
			utils.U256(*self.Value.ToInt()),
		}
	}
	var pkr *c_type.PKr
	if self.To != nil {
		temp := c_type.PKr(*self.To)
		pkr = &temp

	}
	return &stx.ContractCmd{
		asset,
		pkr,
		self.Data,
	}
}

type CmdsArgs struct {
	//Share
	BuyShare *BuyShareArgs
	//Pool
	RegistPool *RegistPoolArgs
	ClosePool  *ClosePoolArgs
	//Contract
	Contract *ContractArgs
	//Package
	PkgCreate   *PkgCreateArgs
	PkgTransfer *PkgTransferArgs
	PkgClose    *PkgCloseArgs
}

func (self *CmdsArgs) toCmds() prepare.Cmds {
	return prepare.Cmds{
		self.BuyShare.toCmd(),
		self.RegistPool.toCmd(),
		self.ClosePool.toCmd(),
		self.Contract.toCmd(),
		self.PkgCreate.toCmd(),
		self.PkgTransfer.toCmd(),
		self.PkgClose.toCmd(),
	}
}

type GenTxArgs struct {
	From       PKAddress
	RefundTo   *PKrAddress
	Receptions []ReceptionArgs
	Cmds       *CmdsArgs
	Gas        uint64
	GasPrice   *Big
	Roots      []c_type.Uint256
}

func (args GenTxArgs) check() error {
	if len(args.Receptions) == 0 && args.Cmds == nil {
		return errors.New("have no receptions")
	}
	if args.GasPrice == nil {
		return fmt.Errorf("gasPrice not specified")
	}

	if args.RefundTo != nil {
		if !superzk.IsPKrValid(args.RefundTo.ToPKr()) {
			return errors.New("RefundTo is not a valid pkr")
		}
	}

	if args.Cmds != nil {
		if args.Cmds.RegistPool != nil || args.Cmds.ClosePool != nil {
			if args.RefundTo == nil {
				return errors.New("Close | Regist StakingNode must need fixed refund address")
			}
		}
	}

	for _, rec := range args.Receptions {
		_, err := validAddress(rec.Addr)
		if err != nil {
			return err
		}
		if rec.Currency.IsEmpty() {
			return errors.Errorf("%v reception currency is nil", hexutil.Encode(rec.Addr[:]))
		}
		if rec.Value == nil {
			return errors.Errorf("%v reception value is nil", hexutil.Encode(rec.Addr[:]))
		}
	}
	return nil

}

func (args GenTxArgs) toTxParam(fromAccount accounts.Account) prepare.PreTxParam {
	gasPrice := args.GasPrice.ToInt()

	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}
	receptions := []prepare.Reception{}
	for _, rec := range args.Receptions {
		pkr := MixAdrressToPkr(rec.Addr)
		var currency c_type.Uint256
		bytes := common.LeftPadBytes([]byte(string(rec.Currency)), 32)
		copy(currency[:], bytes)
		receptions = append(receptions, prepare.Reception{
			pkr,
			assets.Asset{Tkn: &assets.Token{
				Currency: currency,
				Value:    utils.U256(*rec.Value.ToInt())},
			},
		})
	}
	var refundPkr *c_type.PKr
	if args.RefundTo != nil {
		refundPkr = args.RefundTo.ToPKr()
	}
	cmds := prepare.Cmds{}
	if args.Cmds != nil {
		cmds = args.Cmds.toCmds()
	}
	return prepare.PreTxParam{
		fromAccount.Key,
		refundPkr,
		receptions,
		cmds,
		assets.Token{
			utils.CurrencyToUint256("SERO"),
			utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(args.Gas)), args.GasPrice.ToInt())),
		},
		gasPrice,
		args.Roots,
	}
}
