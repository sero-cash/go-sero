package verify

import (
	"errors"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type verifyWithoutStateCtx struct {
	tx           *stx.T
	balance_desc cpt.BalanceDesc
}

func VerifyWithoutState(ehash *keys.Uint256, tx *stx.T) (e error) {
	//Verify EHash
	if *ehash != tx.Ehash {
		e = ReportError("ehash error", tx)
		return
	}
	ctx := verifyWithoutStateCtx{}
	ctx.tx = tx
	ctx.Verify()
	return
}

func (self *verifyWithoutStateCtx) prepare() {
	self.balance_desc.Hash = self.tx.ToHash_for_sign()
}

func (self *verifyWithoutStateCtx) verifyFee() (e error) {
	if !CheckUint(&self.tx.Fee.Value) {
		e = errors.New("txs.verify check fee too big")
		return
	}
	{
		asset_desc := cpt.AssetDesc{
			Tkn_currency: self.tx.Fee.Currency,
			Tkn_value:    self.tx.Fee.Value.ToUint256(),
			Tkt_category: keys.Empty_Uint256,
			Tkt_value:    keys.Empty_Uint256,
		}
		cpt.GenAssetCC(&asset_desc)
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
	}
	return
}

func (self *verifyWithoutStateCtx) verifyFrom() (e error) {
	if !keys.PKrValid(&self.tx.From) {
		e = ReportError("txs.verify from is invalid", self.tx)
		return
	}
	if !keys.VerifyPKr(&self.balance_desc.Hash, &self.tx.Sign, &self.tx.From) {
		e = ReportError("txs.verify from verify failed", self.tx)
		return
	}
	return
}

func (self *verifyWithoutStateCtx) verifyOs() (e error) {
	for _, out := range self.tx.Desc_O.Outs {
		if out.Asset.Tkn != nil {
			if !CheckUint(&out.Asset.Tkn.Value) {
				e = ReportError("o_out tkn value invalid", self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithoutStateCtx) verifyZs() (e error) {
	return
}

func (self *verifyWithoutStateCtx) verifyPkg() (e error) {
	if self.tx.Desc_Cmd.Count() > 0 && self.tx.Desc_Pkg.Count() > 0 {
		e = ReportError("pkg and cmd desc only exists one", self.tx)
		return
	}
	if !self.tx.Desc_Pkg.Valid() {
		e = ReportError("pkg desc is invalid", self.tx)
		return
	}
	return
}

func (self *verifyWithoutStateCtx) verifyCmds() (e error) {
	if !self.tx.Desc_Cmd.Valid() {
		e = ReportError("cmd desc is invalid", self.tx)
		return
	}
	if asset := self.tx.Desc_Cmd.OutAsset(); asset != nil {
		if asset.Tkn != nil {
			if !CheckUint(&asset.Tkn.Value) {
				e = ReportError("cmd asset tkn value invalid", self.tx)
				return
			}
		}
	}
	if pkr := self.tx.Desc_Cmd.ToPkr(); pkr != nil {
		if pkr != nil {
			if !keys.PKrValid(pkr) {
				e = ReportError("cmd pkr invalid", self.tx)
				return
			}
		}
	}

	return
}

func (self *verifyWithoutStateCtx) Verify() (e error) {
	self.prepare()
	if e = self.verifyFee(); e != nil {
		return
	}
	if e = self.verifyFrom(); e != nil {
		return
	}

	if e = self.verifyOs(); e != nil {
		return
	}

	if e = self.verifyZs(); e != nil {
		return
	}

	if e = self.verifyPkg(); e != nil {
		return
	}

	if e = self.verifyPkg(); e != nil {
		return
	}
	return
}

func VerifyWithState() {

}
