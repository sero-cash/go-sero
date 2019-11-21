package generate_1

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type prove_ctx struct {
	param  txtool.GTxParam
	tx     stx.T
	p0_ins []*txtool.GIn
	p_ins  []*txtool.GIn
	c_ins  []*txtool.GIn
	c_outs []*txtool.GOut
	p_outs []*txtool.GOut
}

func (self *prove_ctx) Tx() (ret stx.T) {
	ret = self.tx
	return
}

func (self *prove_ctx) prepare() {
	for i := range self.param.Ins {
		in := &self.param.Ins[i]
		if in.Out.State.OS.Out_O != nil {
			self.p0_ins = append(self.p0_ins, in)
			continue
		}
		if in.Out.State.OS.Out_Z != nil {
			self.p0_ins = append(self.p0_ins, in)
			continue
		}
		if in.Out.State.OS.Out_P != nil {
			self.p_ins = append(self.p_ins, in)
			continue
		}
		if in.Out.State.OS.Out_C != nil {
			if self.param.Z != nil && *self.param.Z {
				self.c_ins = append(self.c_ins, in)
			} else {
				self.p_ins = append(self.p_ins, in)
			}
			continue
		}
	}
	for i := range self.param.Outs {
		out := &self.param.Outs[i]
		if c_superzk.IsSzkPKr(&out.PKr) {
			if self.param.Z != nil && *self.param.Z {
				self.c_outs = append(self.c_outs, out)
			} else {
				self.p_outs = append(self.p_outs, out)
			}
		} else {
			self.p_outs = append(self.p_outs, out)
		}
	}
}

func (self *prove_ctx) prove() (e error) {
	var gen_output_procs = gen_output_procs_pool.GetProcs()
	defer gen_output_procs_pool.PutProcs(gen_output_procs)
	for i, out := range self.c_outs {

		g := gen_output_desc{}
		g.asset_cm = self.tx.Tx1.Outs_C[i].AssetCM
		g.ar = *out.Ar
		g.asset = out.Asset.ToTypeAsset()
		g.index = i

		sip6 := seroparam.SIP6()
		if seroparam.Is_Offline() {
			sip6 = uint64(2123558)
		}
		if self.param.Num != nil && *self.param.Num >= sip6 {
			g.isEx = true
		}

		gen_output_procs.StartProc(&g)
	}

	var gen_input_procs = gen_input_procs_pool.GetProcs()
	defer gen_input_procs_pool.PutProcs(gen_input_procs)
	for i, in := range self.c_ins {
		t_in := self.tx.Tx1.Ins_C[i]
		g := gen_input_desc{}
		g.asset_cm_new = t_in.AssetCM
		g.zpka = t_in.ZPKa
		g.nil = t_in.Nil
		g.anchor = t_in.Anchor
		g.asset_cc = *in.CC
		g.ar_old = *in.ArOld
		g.ar_new = *in.Ar
		g.index = in.Out.State.OS.Index
		copy(g.zpkr[:], in.Out.State.OS.ToPKr()[:32])
		g.vskr = *in.Vskr
		copy(g.baser[:], in.Out.State.OS.ToPKr()[64:])
		g.a = *in.A
		for i, path := range in.Witness.Paths {
			copy(g.paths[len(g.paths)-32-(i*32):], path[:])
		}
		g.pos = uint64(in.Witness.Pos)
		gen_input_procs.StartProc(&g)
	}

	var gen_pkg_procs = gen_pkg_procs_pool.GetProcs()
	defer gen_pkg_procs_pool.PutProcs(gen_pkg_procs)
	if self.param.Cmds.PkgCreate != nil {
		g := gen_pkg_desc{}
		g.asset = self.param.Cmds.PkgCreate.Asset.ToTypeAsset()
		g.ar = self.param.Cmds.PkgCreate.Ar
		g.asset_cm = self.tx.Desc_Pkg.Create.Pkg.AssetCM
		gen_pkg_procs.StartProc(&g)
	}

	//-----------------

	if gen_output_procs.HasProc() {
		if e = gen_output_procs.End(); e == nil {
			for _, g := range gen_output_procs.Runs {
				output_desc := g.(*gen_output_desc)
				self.tx.Tx1.Outs_C[output_desc.index].Proof = output_desc.proof
			}
		} else {
			return
		}
	}

	if gen_input_procs.HasProc() {
		if e = gen_input_procs.End(); e == nil {
			for i, g := range gen_input_procs.Runs {
				input_desc := g.(*gen_input_desc)
				self.tx.Tx1.Ins_C[i].Proof = input_desc.proof
			}
		} else {
			return
		}
	}

	if gen_pkg_procs.HasProc() {
		if e = gen_pkg_procs.End(); e == nil {
			pkg_desc := gen_pkg_procs.Runs[0].(*gen_pkg_desc)
			self.tx.Desc_Pkg.Create.Proof = pkg_desc.proof
		} else {
			return
		}
	}

	return
}

func ProveTx(tx *stx.T, param *txtool.GTxParam) (ctx prove_ctx, e error) {
	ctx.param = *param
	ctx.tx = *tx
	ctx.prepare()

	if e = ctx.prove(); e != nil {
		return
	}
	return
}
