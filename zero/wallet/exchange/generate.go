package exchange

import (
	"math/big"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func (self *Exchange) GenTx(param prepare.PreTxParam) (txParam *txtool.GTxParam, e error) {
	txParam, e = prepare.GenTxParam(&param, self, &prepare.DefaultTxParamState{})
	if e == nil && txParam != nil {
		for _, in := range txParam.Ins {
			self.usedFlag.Store(in.Out.Root, 1)
		}
	}
	return
}

func (self *Exchange) buildTxParam(param *prepare.BeforeTxParam) (txParam *txtool.GTxParam, e error) {

	txParam, e = prepare.BuildTxParam(&prepare.DefaultTxParamState{}, param)

	if e == nil && txParam != nil {
		for _, in := range txParam.Ins {
			self.usedFlag.Store(in.Out.Root, 1)
		}
	}
	return
}

func (self *Exchange) FindRoots(accountKey *common.AccountKey, currency string, amount *big.Int) (roots prepare.Utxos, remain big.Int) {
	utxos, r := self.findUtxos(accountKey, currency, amount)
	for _, utxo := range utxos {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	remain = *r
	return
}

func (self *Exchange) FindRootsByTicket(accountKey *common.AccountKey, tickets []assets.Ticket) (roots prepare.Utxos, remain map[c_type.Uint256]c_type.Uint256) {
	utxos, remain := self.findUtxosByTicket(accountKey, tickets)
	for _, utxo := range utxos {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	return
}

func (self *Exchange) DefaultRefundTo(accountKey *common.AccountKey) (ret *c_type.PKr) {
	if value, ok := self.accounts.Load(*accountKey); ok {
		account := value.(*Account)
		return &account.mainPkr
	}
	return nil
}

func (self *Exchange) GetRoot(root *c_type.Uint256) (utxos *prepare.Utxo) {
	if u, e := self.getUtxo(*root); e != nil {
		return nil
	} else {
		return &prepare.Utxo{u.Root, u.Asset}
	}
}
