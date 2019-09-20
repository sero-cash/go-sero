package verify_0

import (
	"errors"
	"fmt"

	"github.com/sero-cash/go-sero/zero/txtool/verify/verify_utils"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type verifyWithoutStateCtx struct {
	tx              *stx.T
	num             uint64
	hash            c_type.Uint256
	oout_count      int
	oin_count       int
	zout_count      int
	zin_proof_proc  *utils.Procs
	zout_proof_proc *utils.Procs
	pkg_proof_proc  *utils.Procs
}

func VerifyWithoutState(ehash *c_type.Uint256, tx *stx.T, num uint64) (e error) {
	if *ehash != tx.Ehash {
		e = verify_utils.ReportError("ehash error", tx)
		return
	}
	ctx := verifyWithoutStateCtx{}
	ctx.tx = tx
	ctx.num = num
	return ctx.Verify()
}

func (self *verifyWithoutStateCtx) prepare() {
	self.hash = self.tx.ToHash_for_sign()
	self.zin_proof_proc = verify_input_procs_pool.GetProcs()
	self.zout_proof_proc = verify_output_procs_pool.GetProcs()
	self.pkg_proof_proc = verify_pkg_procs_pool.GetProcs()

}

func (self *verifyWithoutStateCtx) clear() {
	verify_input_procs_pool.PutProcs(self.zin_proof_proc)
	verify_output_procs_pool.PutProcs(self.zout_proof_proc)
	verify_pkg_procs_pool.PutProcs(self.pkg_proof_proc)
}

func (self *verifyWithoutStateCtx) verifyFee() (e error) {
	if !verify_utils.CheckUint(&self.tx.Fee.Value) {
		e = errors.New("txs.verify check fee too big")
		return
	}
	self.tx.ToFeeCC()
	self.oout_count++
	return
}

func (self *verifyWithoutStateCtx) verifyFrom() (e error) {
	if !superzk.IsPKrValid(&self.tx.From) {
		e = verify_utils.ReportError("txs.verify from is invalid", self.tx)
		return
	}
	if !superzk.VerifyPKr(&self.hash, &self.tx.Sign, &self.tx.From) {
		e = verify_utils.ReportError("txs.verify from verify failed", self.tx)
		return
	}
	return
}

func (self *verifyWithoutStateCtx) verifyOs() (e error) {
	if self.tx.Tx0() != nil {
		if self.num >= seroparam.SIP4() {
			if len(self.tx.Tx0().Desc_O.Outs) > 0 {
				e = verify_utils.ReportError("after SIP4, o_outs can not used", self.tx)
				return
			}
		}
		for i, out := range self.tx.Tx0().Desc_O.Outs {
			self.oout_count++
			if out.Asset.Tkn != nil {
				if !verify_utils.CheckUint(&out.Asset.Tkn.Value) {
					e = verify_utils.ReportError("o_out tkn value invalid", self.tx)
					return
				}
			}
			self.tx.Tx0().Desc_O.Outs[i].ToAssetCC()
		}

		if self.num >= seroparam.VP0() {
			if len(self.tx.Tx0().Desc_O.Ins) > seroparam.MAX_O_INS_LENGTH {
				e = verify_utils.ReportError(fmt.Sprintf("txs.verify O ins length > %v, current is %v", seroparam.MAX_O_INS_LENGTH, len(self.tx.Tx0().Desc_O.Ins)), self.tx)
				return
			}
		}
		for range self.tx.Tx0().Desc_O.Ins {
			self.oin_count++
		}
	}
	return
}

func (self *verifyWithoutStateCtx) verifyZs() (e error) {
	if self.tx.Tx0() != nil {
		for _, out := range self.tx.Tx0().Desc_Z.Outs {
			self.zout_count++
			if !superzk.IsPKrValid(&out.PKr) {
				e = verify_utils.ReportError("z_out pkr invalid", self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithoutStateCtx) verifyPkg() (e error) {
	if self.tx.Desc_Cmd.Count() > 0 && self.tx.Desc_Pkg.Count() > 0 {
		e = verify_utils.ReportError("pkg and cmd desc only exists one", self.tx)
		return
	}
	if !self.tx.Desc_Pkg.Valid() {
		e = verify_utils.ReportError("pkg desc is invalid", self.tx)
		return
	}
	if self.tx.Desc_Pkg.Create != nil {
		self.zout_count++
	}
	return
}

func (self *verifyWithoutStateCtx) verifyCmds() (e error) {
	if self.num < seroparam.SIP4() {
		if self.tx.Desc_Cmd.Count() > 0 {
			e = verify_utils.ReportError("can not use tx cmd until SIP4", self.tx)
		}
		return
	}
	if !self.tx.Desc_Cmd.Valid() {
		e = verify_utils.ReportError("cmd desc is invalid", self.tx)
		return
	}
	if asset := self.tx.Desc_Cmd.OutAsset(); asset != nil {
		self.oout_count++
		if asset.Tkn != nil {
			if !verify_utils.CheckUint(&asset.Tkn.Value) {
				e = verify_utils.ReportError("cmd asset tkn value invalid", self.tx)
				return
			}
		}
		self.tx.Desc_Cmd.ToAssetCC()
	}
	if pkr := self.tx.Desc_Cmd.ToPkr(); pkr != nil {
		if !superzk.IsPKrValid(pkr) {
			e = verify_utils.ReportError("cmd pkr invalid", self.tx)
			return
		}
	}
	if self.tx.Desc_Cmd.RegistPool != nil {
		if self.tx.Desc_Cmd.RegistPool.FeeRate > seroparam.HIGHEST_STAKING_NODE_FEE_RATE {
			e = verify_utils.ReportError(fmt.Sprintf("regist pool the fee rate must < %v%%", seroparam.HIGHEST_STAKING_NODE_FEE_RATE), self.tx)
			return
		}
		if self.tx.Desc_Cmd.RegistPool.FeeRate < seroparam.LOWEST_STAKING_NODE_FEE_RATE {
			e = verify_utils.ReportError(fmt.Sprintf("regist pool fee must >= %v%%", seroparam.LOWEST_STAKING_NODE_FEE_RATE/100), self.tx)
			return
		}
	}
	if self.tx.Desc_Cmd.Contract != nil {
		if self.tx.Desc_Cmd.Contract.To != nil {
			empty := c_type.PKr{}
			if *self.tx.Desc_Cmd.Contract.To == empty {
				e = verify_utils.ReportError("contract target can not be zero", self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithoutStateCtx) Verify() (e error) {
	self.prepare()
	defer self.clear()

	self.ProcessVerifyProof()

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

	if e = self.verifyCmds(); e != nil {
		return
	}

	if self.oout_count > seroparam.MAX_O_OUT_LENGTH {
		e = verify_utils.ReportError("oout count > 10", self.tx)
		return
	}
	if self.num >= seroparam.SIP2() {
		if self.zout_count > seroparam.MAX_Z_OUT_LENGTH_SIP2 {
			e = verify_utils.ReportError("verify error: out_size > 500", self.tx)
			return
		}
	} else {
		if self.zout_count > seroparam.MAX_Z_OUT_LENGTH_OLD {
			e = verify_utils.ReportError("verify error: out_size > 6", self.tx)
			return
		}
	}

	if e = self.WaitVerifyProof(); e != nil {
		return
	}

	return
}
