package prepare

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common/hexutil"
)

func SelectUtxos(param *PreTxParam, generator TxParamGenerator) (utxos Utxos, e error) {
	if len(param.Roots) > 0 {
		for _, root := range param.Roots {
			if utxo := generator.GetRoot(&root); utxo == nil {
				return utxos, fmt.Errorf("can not find the utxo for root : %v", hexutil.Encode(root[:]))
			} else {
				utxos = append(utxos, *utxo)
			}
		}
		return
	} else {
		ck := assets.NewCKState(true, &param.Fee)

		if cmdsAsset := param.Cmds.OutAsset(); cmdsAsset != nil {
			ck.AddOut(cmdsAsset)
		}

		for _, reception := range param.Receptions {
			ck.AddOut(&reception.Asset)
		}

		if cmdsAsset, err := param.Cmds.InAsset(); err != nil {
			e = err
			return
		} else {
			if cmdsAsset != nil {
				ck.AddIn(cmdsAsset)
			}
		}

		if tks := ck.Tkts(); len(tks) > 0 {
			if outs, remain := generator.FindRootsByTicket(&param.From, tks); len(remain) == 0 {
				utxos = append(utxos, outs...)
				for _, out := range outs {
					ck.AddIn(&out.Asset)
				}
			} else {
				e = errors.New("no enough unlocked utxos")
				return
			}
		}

		for _, tkn := range ck.Tkns() {
			outs, remain := generator.FindRoots(&param.From, utils.Uint256ToCurrency(&tkn.Currency), tkn.Value.ToIntRef())
			if remain.Sign() <= 0 {
				utxos = append(utxos, outs...)
			} else {
				e = errors.New("no enough unlocked utxos")
				return
			}
		}

		return
	}
}

type BeforeTxParam struct {
	Fee        assets.Token
	GasPrice   big.Int
	Utxos      Utxos
	RefundTo   c_type.PKr
	Receptions []Reception
	Cmds       Cmds
}

func BuildTxParam(
	state TxParamState,
	param *BeforeTxParam,
) (txParam *txtool.GTxParam, e error) {

	txParam = &txtool.GTxParam{}

	ck := assets.NewCKState(false, &param.Fee)

	txParam.Fee = param.Fee
	txParam.GasPrice = &param.GasPrice

	txParam.From = txtool.Kr{PKr: param.RefundTo}

	wits, err := state.GetAnchor(param.Utxos.Roots())
	if err != nil {
		e = err
		return
	}

	Ins := []txtool.GIn{}
	oins_count := 0
	for index, utxo := range param.Utxos {
		if out := state.GetOut(&utxo.Root); out != nil {
			if ck.AddIn(&utxo.Asset) {
				Ins = append(Ins, txtool.GIn{Out: txtool.Out{Root: utxo.Root, State: *out}, Witness: wits[index]})
				if out.OS.Out_O != nil {
					oins_count++
				}
			}
		} else {
			e = fmt.Errorf("can not find Out for utxo %v", hexutil.Encode(utxo.Root[:]))
			return
		}
	}

	if oins_count > 2500 {
		e = fmt.Errorf("o_ins count > 2500")
		return
	}

	if cmdsAsset, err := param.Cmds.InAsset(); err != nil {
		e = err
		return
	} else {
		if cmdsAsset != nil {
			ck.AddIn(cmdsAsset)
		}
	}

	Outs := []txtool.GOut{}
	for _, reception := range param.Receptions {
		pkr := reception.Addr
		if IsPk(reception.Addr) {
			pk := reception.Addr.ToUint512()
			pkr = CreatePkr(&pk, 0)
		}
		ck.AddOut(&reception.Asset)
		Outs = append(Outs, txtool.GOut{PKr: pkr, Asset: reception.Asset})
	}

	if cmdsAsset := param.Cmds.OutAsset(); cmdsAsset != nil {
		ck.AddOut(cmdsAsset)
	}

	tkns, tkts := ck.GetList()
	var maxlen int
	if len(tkns) > maxlen {
		maxlen = len(tkns)
	}
	if len(tkts) > maxlen {
		maxlen = len(tkts)
	}

	for i := 0; i < maxlen; i++ {
		a := assets.Asset{}
		if i < len(tkns) {
			tkn := tkns[i]
			a.Tkn = &assets.Token{}
			a.Tkn.Currency = tkn.Currency
			a.Tkn.Value = tkn.Value
		}
		if i < len(tkts) {
			tkt := tkts[i]
			a.Tkt = &assets.Ticket{}
			a.Tkt.Category = tkt.Category
			a.Tkt.Value = tkt.Value
		}
		if a.HasAsset() {
			ck.AddOut(&a)
			Outs = append(Outs, txtool.GOut{PKr: txParam.From.PKr, Asset: a})
		}
	}

	if e = ck.Check(); e != nil {
		return
	}

	txParam.Ins = Ins
	txParam.Outs = Outs
	txParam.Cmds.Contract = param.Cmds.Contract
	txParam.Cmds.BuyShare = param.Cmds.BuyShare
	txParam.Cmds.RegistPool = param.Cmds.RegistPool
	txParam.Cmds.ClosePool = param.Cmds.ClosePool

	if param.Cmds.PkgCreate != nil {
		if pkg := state.GetPkgById(&param.Cmds.PkgCreate.Id); pkg != nil {
			e = errors.New("create pkg but the pkg id is exsits")
			return
		}
		txParam.Cmds.PkgCreate = &txtool.GPkgCreateCmd{}
		txParam.Cmds.PkgCreate.Id = param.Cmds.PkgCreate.Id
		txParam.Cmds.PkgCreate.PKr = param.Cmds.PkgCreate.PKr
		txParam.Cmds.PkgCreate.Asset = param.Cmds.PkgCreate.Asset
		txParam.Cmds.PkgCreate.Memo = param.Cmds.PkgCreate.Memo
	}

	if param.Cmds.PkgTransfer != nil {
		if pkg := state.GetPkgById(&param.Cmds.PkgTransfer.Id); pkg == nil {
			e = errors.New("transfer pkg but the pkg id is not exsits")
			return
		} else {
			if !pkg.Closed {
				txParam.Cmds.PkgTransfer = &txtool.GPkgTransferCmd{}
				txParam.Cmds.PkgTransfer.Id = param.Cmds.PkgTransfer.Id
				txParam.Cmds.PkgTransfer.PKr = param.Cmds.PkgTransfer.PKr
				txParam.Cmds.PkgTransfer.Owner = pkg.Pack.PKr
			} else {
				e = errors.New("transfer pkg but the pkg is closed")
				return
			}
		}
	}

	if param.Cmds.PkgClose != nil {
		if p := state.GetPkgById(&param.Cmds.PkgTransfer.Id); p == nil {
			e = errors.New("close pkg but the pkg id is not exsits")
			return
		} else {
			if !p.Closed {
				txParam.Cmds.PkgClose.Id = param.Cmds.PkgClose.Id
				txParam.Cmds.PkgClose.Owner = p.Pack.PKr
				txParam.Cmds.PkgClose.AssetCM = p.Pack.Pkg.AssetCM
				if opkg, err := pkg.DePkg(&param.Cmds.PkgClose.Key, &p.Pack.Pkg); err != nil {
					e = errors.New("close pkg but password is error")
					return
				} else {
					txParam.Cmds.PkgClose.Ar = opkg.Ar
				}
			} else {
				e = errors.New("close pkg but the pkg is closed")
				return
			}
		}
	}

	var contractTo *common.Address
	if param.Cmds.Contract != nil {
		if param.Cmds.Contract.To != nil {
			contractTo = &common.Address{}
			copy(contractTo[:], param.Cmds.Contract.To[:])
		}
	}

	if gaslimit, err := state.GetSeroGasLimit(contractTo, &txParam.Fee, txParam.GasPrice); err != nil {
		e = err
		return
	} else {
		txParam.Gas = gaslimit
	}

	if e = txParam.GenZ(); e != nil {
		return
	}

	return

}
