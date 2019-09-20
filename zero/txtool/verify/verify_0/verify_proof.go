package verify_0

func (self *verifyWithoutStateCtx) ProcessVerifyProof() {

	if self.tx.Tx0() != nil {
		for _, in_z := range self.tx.Tx0().Desc_Z.Ins {
			g := verify_input_desc{}
			g.desc.Nil = in_z.Nil
			g.desc.Anchor = in_z.Anchor
			g.desc.AssetCM = in_z.AssetCM
			g.desc.Proof = in_z.Proof
			self.zin_proof_proc.StartProc(&g)
		}

		for _, out_z := range self.tx.Tx0().Desc_Z.Outs {
			g := verify_output_desc{}
			g.desc.AssetCM = out_z.AssetCM
			g.desc.RPK = out_z.RPK
			g.pkr = out_z.PKr
			g.desc.OutCM = out_z.OutCM
			g.desc.Proof = out_z.Proof
			g.desc.Height = self.num
			self.zout_proof_proc.StartProc(&g)
		}
	}

	if self.tx.Desc_Pkg.Create != nil {
		create := self.tx.Desc_Pkg.Create

		g := verify_pkg_desc{}
		g.desc.AssetCM = create.Pkg.AssetCM
		g.desc.PkgCM = create.Pkg.PkgCM
		g.desc.Proof = create.Proof

		self.pkg_proof_proc.StartProc(&g)
	}

}

func (self *verifyWithoutStateCtx) WaitVerifyProof() (e error) {
	if self.zin_proof_proc.HasProc() {
		if e = self.zin_proof_proc.End(); e != nil {
			return
		}
	}
	if self.zout_proof_proc.HasProc() {
		if e = self.zout_proof_proc.End(); e != nil {
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
