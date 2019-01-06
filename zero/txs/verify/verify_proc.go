package verify

import (
	"errors"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_pkg_procs_pool = utils.NewProcsPool(func() int { return G_v_thread_num })

type verify_pkg_desc struct {
	desc cpt.PkgVerifyDesc
	e    error
}

func (self *verify_pkg_desc) Run() bool {
	if err := cpt.VerifyPkg(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

var verify_input_procs_pool = utils.NewProcsPool(func() int { return G_v_thread_num })

type verify_input_desc struct {
	desc cpt.InputVerifyDesc
	e    error
}

func (self *verify_input_desc) Run() bool {
	if err := cpt.VerifyInput(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

var verify_output_procs_pool = utils.NewProcsPool(func() int { return G_v_thread_num })

type verify_output_desc struct {
	desc cpt.OutputVerifyDesc
	pkr  keys.PKr
	e    error
}

func (self *verify_output_desc) Run() bool {
	if keys.PKrValid(&self.pkr) {
		if err := cpt.VerifyOutput(&self.desc); err != nil {
			self.e = err
			return false
		} else {
			return true
		}
	} else {
		self.e = errors.New("z_out pkr is invalid !")
		return false
	}
}

func verifyDesc_Zs(tx *stx.T, balance_desc *cpt.BalanceDesc) (e error) {
	var verify_pkg_procs = verify_input_procs_pool.GetProcs()
	defer verify_pkg_procs_pool.PutProcs(verify_pkg_procs)

	if tx.Desc_Pkg.Create != nil {
		create := tx.Desc_Pkg.Create
		balance_desc.Zout_acms = append(balance_desc.Zout_acms, create.Pkg.AssetCM[:]...)

		g := verify_pkg_desc{}
		g.desc.AssetCM = create.Pkg.AssetCM
		g.desc.PkgCM = create.Pkg.PkgCM
		g.desc.Proof = create.Proof

		verify_pkg_procs.StartProc(&g)
	}

	var verify_input_procs = verify_input_procs_pool.GetProcs()
	defer verify_input_procs_pool.PutProcs(verify_input_procs)

	for _, in_z := range tx.Desc_Z.Ins {
		balance_desc.Zin_acms = append(balance_desc.Zin_acms, in_z.AssetCM[:]...)

		g := verify_input_desc{}
		g.desc.Nil = in_z.Nil
		g.desc.Anchor = in_z.Anchor
		g.desc.AssetCM = in_z.AssetCM
		g.desc.Proof = in_z.Proof

		verify_input_procs.StartProc(&g)
	}

	var verify_output_procs = verify_output_procs_pool.GetProcs()
	defer verify_output_procs_pool.PutProcs(verify_output_procs)

	for _, out_z := range tx.Desc_Z.Outs {
		balance_desc.Zout_acms = append(balance_desc.Zout_acms, out_z.AssetCM[:]...)

		g := verify_output_desc{}
		g.desc.AssetCM = out_z.AssetCM
		g.desc.RPK = out_z.RPK
		g.pkr = out_z.PKr
		g.desc.OutCM = out_z.OutCM
		g.desc.Proof = out_z.Proof

		verify_output_procs.StartProc(&g)
	}

	if verify_pkg_procs.HasProc() {
		if p_runs := verify_pkg_procs.Wait(); p_runs != nil {
		} else {
			e = errors.New("verify pkg desc_z failed!!!")
			return
		}
	}

	if verify_output_procs.HasProc() {
		if o_runs := verify_output_procs.Wait(); o_runs != nil {
		} else {
			e = errors.New("verify output desc_z failed!!!")
			return
		}
	}

	if verify_input_procs.HasProc() {
		if i_runs := verify_input_procs.Wait(); i_runs != nil {
		} else {
			e = errors.New("verify input desc_z failed!!!")
			return
		}
	}
	return
}
