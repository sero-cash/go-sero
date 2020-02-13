package verify_1

import "github.com/sero-cash/go-czero-import/seroparam"

func (self *verifyWithoutStateCtx) ProcessVerifyProof() {
	for _, in := range self.tx.Tx1.Ins_C {
		g := verify_input_desc{}
		g.anchor = in.Anchor
		g.proof = in.Proof
		g.nil = in.Nil
		g.zpka = in.ZPKa
		g.asset_cm_new = in.AssetCM
		self.cin_proof_proc.StartProc(&g)
	}
	for _, out := range self.tx.Tx1.Outs_C {
		g := verify_output_desc{}
		g.proof = out.Proof
		g.asset_cm = out.AssetCM
		g.pkr = out.PKr
		if self.num >= seroparam.SIP7() {
			g.isEx = true
		}
		self.cout_proof_proc.StartProc(&g)
	}
	if self.tx.Desc_Pkg.Create != nil {
		g := verify_pkg_desc{}
		g.asset_cm = self.tx.Desc_Pkg.Create.Pkg.AssetCM
		g.proof = self.tx.Desc_Pkg.Create.Proof
		self.pkg_proof_proc.StartProc(&g)
	}
}

func (self *verifyWithoutStateCtx) WaitVerifyProof() (e error) {
	if self.cin_proof_proc.HasProc() {
		if e = self.cin_proof_proc.End(); e != nil {
			return
		}
	}
	if self.cout_proof_proc.HasProc() {
		if e = self.cout_proof_proc.End(); e != nil {
			return
		}
	}
	if self.pkg_proof_proc.HasProc() {
		if e = self.pkg_proof_proc.End(); e != nil {
			return
		}
	}
	return
}
