package verify_0

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/zero/txtool/verify/verify_utils"

	"github.com/sero-cash/go-czero-import/c_czero"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/utils"
)

type verifyWithStateCtx struct {
	tx             *stx.T
	state          *zstate.ZState
	oin_proof_proc *utils.Procs
	balance_desc   c_type.BalanceDesc
}

func VerifyWithState(tx *stx.T, state *zstate.ZState) (e error) {
	hash_z := tx.ToHash_for_sign()
	ctx := verifyWithStateCtx{}
	ctx.tx = tx
	ctx.state = state
	ctx.balance_desc.Hash = hash_z
	return ctx.Verify()
}

func (self *verifyWithStateCtx) prepare() {
	self.oin_proof_proc = verify_input_o_procs_pool.GetProcs()
}

func (self *verifyWithStateCtx) clear() {
	verify_input_o_procs_pool.PutProcs(self.oin_proof_proc)
}

func (self *verifyWithStateCtx) verifyFee() (e error) {
	feeCC := self.tx.ToFeeCC_Czero()
	self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, feeCC[:]...)
	return
}

func (self *verifyWithStateCtx) verifyOs() (e error) {
	if self.tx.Tx0() != nil {
		for _, in_o := range self.tx.Tx0().Desc_O.Ins {
			if self.state.Num() >= seroparam.SIP2() {
				if ok := self.state.State.HasIn(&in_o.Nil); ok {
					e = verify_utils.ReportError("txs.verify in_o already in nils", self.tx)
					return
				}
			} else {
				if ok := self.state.State.HasIn(&in_o.Root); ok {
					e = verify_utils.ReportError("txs.verify in_o already in roots", self.tx)
					return
				} else {
				}
			}
			if src := self.state.State.GetOut(&in_o.Root); src != nil {
				desc := verify_input_o_desc{}
				desc.in = in_o
				desc.hash_z = self.balance_desc.Hash
				desc.src = *src
				self.oin_proof_proc.StartProc(&desc)
			} else {
				e = verify_utils.ReportError("txs.Verify: in_o not find in the outs!", self.tx)
				return
			}
		}

		for _, out_o := range self.tx.Tx0().Desc_O.Outs {
			assetCC := out_o.ToAssetCC_Czero()
			self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, assetCC[:]...)
		}
	}

	return
}

func (self *verifyWithStateCtx) Wait() (e error) {
	if self.oin_proof_proc.HasProc() {
		if e = self.oin_proof_proc.End(); e == nil {
			for _, p_run := range self.oin_proof_proc.Runs {
				desc := p_run.(*verify_input_o_desc)
				self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, desc.asset_cc[:]...)
			}
		} else {
			return
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyZs() (e error) {
	if self.tx.Tx0() != nil {
		for _, in_z := range self.tx.Tx0().Desc_Z.Ins {
			self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, in_z.AssetCM[:]...)
			if ok := self.state.State.HasIn(&in_z.Nil); ok {
				e = verify_utils.ReportError("txs.verify in already in nils", self.tx)
				return
			} else {
				if out := self.state.State.GetOut(&in_z.Anchor); out == nil {
					e = verify_utils.ReportError("txs.verify can not find out for anchor", self.tx)
					return
				} else {
				}
			}
		}
		for _, out_z := range self.tx.Tx0().Desc_Z.Outs {
			self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, out_z.AssetCM[:]...)
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyPkg() (e error) {
	if self.tx.Desc_Pkg.Create != nil {
		if pg := self.state.Pkgs.GetPkgById(&self.tx.Desc_Pkg.Create.Id); pg != nil {
			e = verify_utils.ReportError(fmt.Sprintf("pkg id already exists %v", hexutil.Encode(self.tx.Desc_Pkg.Create.Id[:])), self.tx)
			return
		} else {
			self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, self.tx.Desc_Pkg.Create.Pkg.AssetCM[:]...)
		}
	}

	if self.tx.Desc_Pkg.Transfer != nil {
		if pg := self.state.Pkgs.GetPkgById(&self.tx.Desc_Pkg.Transfer.Id); pg == nil || pg.Closed {
			e = verify_utils.ReportError(fmt.Sprintf("Can not find pkg of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Transfer.Id[:])), self.tx)
			return
		} else {
			if superzk.VerifyPKr(&self.balance_desc.Hash, &self.tx.Desc_Pkg.Transfer.Sign, &pg.Pack.PKr) {
			} else {
				e = verify_utils.ReportError(fmt.Sprintf("Can not verify pkg sign of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Transfer.Id[:])), self.tx)
				return
			}
		}
	}

	if self.tx.Desc_Pkg.Close != nil {
		if pg := self.state.Pkgs.GetPkgById(&self.tx.Desc_Pkg.Close.Id); pg == nil || pg.Closed {
			e = verify_utils.ReportError(fmt.Sprintf("Can not find pkg of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Close.Id[:])), self.tx)
			return
		} else {
			if superzk.VerifyPKr(&self.balance_desc.Hash, &self.tx.Desc_Pkg.Close.Sign, &pg.Pack.PKr) {
				self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, pg.Pack.Pkg.AssetCM[:]...)
			} else {
				e = verify_utils.ReportError(fmt.Sprintf("Can not verify pkg sign of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Close.Id[:])), self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyCmds() (e error) {
	if cc := self.tx.Desc_Cmd.ToAssetCC_Czero(); cc != nil {
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, cc[:]...)
	}
	return
}

func (self *verifyWithStateCtx) verifyBalance() (e error) {
	self.balance_desc.Bcr = self.tx.Bcr
	self.balance_desc.Bsign = self.tx.Bsign
	if err := c_czero.VerifyBalance(&self.balance_desc); err != nil {
		e = err
		return
	}
	return
}

func (self *verifyWithStateCtx) Verify() (e error) {
	self.prepare()
	defer self.clear()

	if e = self.verifyFee(); e != nil {
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

	if e = self.verifyCmds(); e != nil {
		return
	}

	if e = self.Wait(); e != nil {
		return
	}

	if e = self.verifyBalance(); e != nil {
		return
	}

	return
}
