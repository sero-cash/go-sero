package verify

import (
	"errors"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

func verifyDesc_Zs(tx *stx.T, balance_desc *cpt.BalanceDesc, height uint64) (e error) {
	var verify_pkg_procs = verify_input_procs_pool.GetProcs()
	defer verify_pkg_procs_pool.PutProcs(verify_pkg_procs)

	if tx.Desc_Pkg.Create != nil {
		create := tx.Desc_Pkg.Create

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
		g.desc.Height = height

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
