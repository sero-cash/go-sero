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
		ck := NewCKState(true, &param.Fee)

		if cmdsAsset := param.Cmds.OutAsset(); cmdsAsset != nil {
			ck.AddOut(cmdsAsset)
		}

		if cmdsAsset, err := param.Cmds.InAsset(); err != nil {
			e = err
			return
		} else {
			if cmdsAsset != nil {
				ck.AddIn(cmdsAsset)
			}
		}

		for _, reception := range param.Receptions {
			ck.AddOut(&reception.Asset)
		}

		for currency, value := range ck.cy {
			sign := value.balance.ToIntRef().Sign()
			if sign > 0 {
				outs, remain := generator.FindRoots(&param.From, utils.Uint256ToCurrency(&currency), new(big.Int).Abs(value.balance.ToIntRef()))
				if remain.Sign() <= 0 {
					utxos = append(utxos, outs...)
				} else {
					e = errors.New("no enough unlocked utxos")
					return
				}
			}
		}

		for _, utxo := range utxos {
			ck.AddIn(&utxo.Asset)
		}

		if len(ck.tk) > 0 {
			outs, _ := generator.FindRootsByTicket(&param.From, ck.tk)
			for _, out := range outs {
				ck.AddIn(&out.Asset)
			}
		}
		return
	}
}

func BuildTxParam(
	state TxParamState,
	utxos Utxos,
	refundTo *c_type.PKr,
	receptions []Reception,
	cmds *Cmds,
	fee *assets.Token,
	gasPrice *big.Int,
) (txParam *txtool.GTxParam, e error) {

	txParam = &txtool.GTxParam{}

	ck := NewCKState(false, fee)

	txParam.Fee = *fee
	txParam.GasPrice = gasPrice

	txParam.From = txtool.Kr{PKr: *refundTo}

	wits, err := state.GetAnchor(utxos.Roots())
	if err != nil {
		e = err
		return
	}

	Ins := []txtool.GIn{}
	oins_count := 0
	for index, utxo := range utxos {
		if out := state.GetOut(&utxo.Root); out != nil {
			if added, err := ck.AddIn(&utxo.Asset); err != nil {
				e = err
				return
			} else {
				if added {
					Ins = append(Ins, txtool.GIn{Out: txtool.Out{Root: utxo.Root, State: *out}, Witness: wits[index]})
					if out.OS.Out_O != nil {
						oins_count++
					}
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

	if cmds != nil {
		if cmdsAsset, err := cmds.InAsset(); err != nil {
			e = err
			return
		} else {
			if cmdsAsset != nil {
				ck.AddIn(cmdsAsset)
			}
		}
	}

	Outs := []txtool.GOut{}
	for _, reception := range receptions {
		pkr := reception.Addr
		if IsPk(reception.Addr) {
			pk := reception.Addr.ToUint512()
			pkr = CreatePkr(&pk, 0)
		}
		ck.AddOut(&reception.Asset)
		Outs = append(Outs, txtool.GOut{PKr: pkr, Asset: reception.Asset})
	}

	if cmds != nil {
		if cmdsAsset := cmds.OutAsset(); cmdsAsset != nil {
			ck.AddOut(cmdsAsset)
		}
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
	txParam.Cmds.Contract = cmds.Contract
	txParam.Cmds.BuyShare = cmds.BuyShare
	txParam.Cmds.RegistPool = cmds.RegistPool
	txParam.Cmds.ClosePool = cmds.ClosePool

	if cmds.PkgCreate != nil {
		if pkg := state.GetPkgById(&cmds.PkgCreate.Id); pkg != nil {
			e = errors.New("create pkg but the pkg id is exsits")
			return
		}
		txParam.Cmds.PkgCreate = &txtool.GPkgCreateCmd{}
		txParam.Cmds.PkgCreate.Id = cmds.PkgCreate.Id
		txParam.Cmds.PkgCreate.PKr = cmds.PkgCreate.PKr
		txParam.Cmds.PkgCreate.Asset = cmds.PkgCreate.Asset
		txParam.Cmds.PkgCreate.Memo = cmds.PkgCreate.Memo
	}

	if cmds.PkgTransfer != nil {
		if pkg := state.GetPkgById(&cmds.PkgTransfer.Id); pkg == nil {
			e = errors.New("transfer pkg but the pkg id is not exsits")
			return
		} else {
			if !pkg.Closed {
				txParam.Cmds.PkgTransfer = &txtool.GPkgTransferCmd{}
				txParam.Cmds.PkgTransfer.Id = cmds.PkgTransfer.Id
				txParam.Cmds.PkgTransfer.PKr = cmds.PkgTransfer.PKr
				txParam.Cmds.PkgTransfer.Owner = pkg.Pack.PKr
			} else {
				e = errors.New("transfer pkg but the pkg is closed")
				return
			}
		}
	}

	if cmds.PkgClose != nil {
		if p := state.GetPkgById(&cmds.PkgTransfer.Id); p == nil {
			e = errors.New("close pkg but the pkg id is not exsits")
			return
		} else {
			if !p.Closed {
				txParam.Cmds.PkgClose.Id = cmds.PkgClose.Id
				txParam.Cmds.PkgClose.Owner = p.Pack.PKr
				txParam.Cmds.PkgClose.AssetCM = p.Pack.Pkg.AssetCM
				if opkg, err := pkg.DePkg(&cmds.PkgClose.Key, &p.Pack.Pkg); err != nil {
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
	if cmds.Contract != nil {
		if cmds.Contract.To != nil {
			contractTo = &common.Address{}
			copy(contractTo[:], cmds.Contract.To[:])
		}
	}

	if gaslimit, err := state.GetSeroGasLimit(contractTo, &txParam.Fee, txParam.GasPrice); err != nil {
		e = err
		return
	} else {
		txParam.Gas = gaslimit
	}

	return

}
