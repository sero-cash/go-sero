package exchange

import (
	"math/big"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func (self *Exchange) GenTx(param prepare.PreTxParam) (txParam *txtool.GTxParam, e error) {
	txParam, e = prepare.GenTxParam(&param, self, &prepare.DefaultTxParamState{})
	if e == nil {
		for _, in := range txParam.Ins {
			self.usedFlag.Store(in.Out.Root, 1)
		}
	}
	return
}

func (self *Exchange) buildTxParam(
	utxos prepare.Utxos,
	refundTo *keys.PKr,
	receptions []prepare.Reception,
	cmds *prepare.Cmds,
	fee *assets.Token,
	gasPrice *big.Int) (txParam *txtool.GTxParam, e error) {

	txParam, e = prepare.BuildTxParam(&prepare.DefaultTxParamState{}, utxos, refundTo, receptions, cmds, fee, gasPrice)

	if e == nil {
		for _, in := range txParam.Ins {
			self.usedFlag.Store(in.Out.Root, 1)
		}
	}
	return
}

func (self *Exchange) FindRoots(pk *keys.Uint512, currency string, amount *big.Int) (roots prepare.Utxos, remain big.Int) {
	utxos, r := self.findUtxos(pk, currency, amount)
	for _, utxo := range utxos {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	remain = *r
	return
}

func (self *Exchange) FindRootsByTicket(pk *keys.Uint512, tickets map[keys.Uint256]keys.Uint256) (roots prepare.Utxos, remain map[keys.Uint256]keys.Uint256) {
	utxos, remain := self.findUtxosByTicket(pk, tickets)
	for _, utxo := range utxos {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	return
}

func (self *Exchange) DefaultRefundTo(from *keys.Uint512) (ret *keys.PKr) {
	if value, ok := self.accounts.Load(from); ok {
		account := value.(*Account)
		return &account.mainPkr
	} else {
		return nil
	}
}

func (self *Exchange) GetRoot(root *keys.Uint256) (utxos *prepare.Utxo) {
	if u, e := self.getUtxo(*root); e != nil {
		return nil
	} else {
		return &prepare.Utxo{u.Root, u.Asset}
	}
}
