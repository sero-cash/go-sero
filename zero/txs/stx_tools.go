// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package txs

import (
	"errors"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

var gen_input_procs_pool = utils.NewProcsPool(3)

type gen_input_desc struct {
	desc cpt.InputDesc
	e    error
}

func (self *gen_input_desc) Run() bool {
	if err := cpt.GenInputProof(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

var gen_output_procs_pool = utils.NewProcsPool(3)

type gen_output_desc struct {
	desc cpt.OutputDesc
	e    error
}

func (self *gen_output_desc) Run() bool {
	if err := cpt.GenOutputProof(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

func genDesc_Zs(seed *keys.Uint256, ptx *preTx, hash_o *keys.Uint256, balance_desc *cpt.BalanceDesc) (desc_z stx.Desc_Z, e error) {
	var gen_input_procs = gen_input_procs_pool.GetProcs()
	defer gen_input_procs_pool.PutProcs(gen_input_procs)

	for _, in := range ptx.desc_z.ins {
		g := gen_input_desc{}
		g.desc.Seed = *seed
		g.desc.Pkr = in.Out_Z.PKr
		g.desc.Einfo = in.Out_Z.EInfo
		g.desc.Index = in.OutIndex
		_, g.desc.Position, g.desc.Path, g.desc.Anchor = in.ToWitness()

		gen_input_procs.StartProc(&g)
	}

	var gen_output_procs = gen_output_procs_pool.GetProcs()
	defer gen_output_procs_pool.PutProcs(gen_output_procs)

	for _, out := range ptx.desc_z.outs {

		g := gen_output_desc{}

		g.desc.Seed = *seed
		asset := out.Asset.ToCompleteAsset()
		g.desc.Tkn_currency = asset.Tkn.Currency
		g.desc.Tkn_value = asset.Tkn.Value.ToUint256()
		g.desc.Tkt_category = asset.Tkt.Category
		g.desc.Tkt_value = asset.Tkt.Value
		g.desc.Memo = out.Memo
		g.desc.Pk = out.Addr

		gen_output_procs.StartProc(&g)
	}

	if i_runs := gen_input_procs.Wait(); i_runs != nil {
		if o_runs := gen_output_procs.Wait(); o_runs != nil {
			for _, g := range i_runs {
				desc := g.(*gen_input_desc).desc
				in_z := stx.In_Z{}
				in_z.Anchor = desc.Anchor
				in_z.AssetCM = desc.Asset_cm_ret
				in_z.Nil = desc.Nil_ret
				in_z.Trace = desc.Til_ret
				in_z.Proof = desc.Proof_ret
				desc_z.Ins = append(desc_z.Ins, in_z)
				balance_desc.Zin_acms = append(balance_desc.Zin_acms, desc.Asset_cm_ret[:]...)
				balance_desc.Zin_ars = append(balance_desc.Zin_ars, desc.Ar_ret[:]...)
			}
			for _, g := range o_runs {
				desc := g.(*gen_output_desc).desc
				out_z := stx.Out_Z{}
				out_z.AssetCM = desc.Asset_cm_ret
				out_z.OutCM = desc.Out_cm_ret
				out_z.PKr = desc.Pkr_ret
				out_z.EInfo = desc.Einfo_ret
				out_z.Proof = desc.Proof_ret
				desc_z.Outs = append(desc_z.Outs, out_z)
				balance_desc.Zout_acms = append(balance_desc.Zout_acms, desc.Asset_cm_ret[:]...)
				balance_desc.Zout_ars = append(balance_desc.Zout_ars, desc.Ar_ret[:]...)
			}
		} else {
			e = errors.New("gen output desc_z failed!!!")
			return
		}
	} else {
		e = errors.New("gen input desc_z failed!!!")
		return
	}

	return
}

var verify_input_procs_pool = utils.NewProcsPool(5)

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

var verify_output_procs_pool = utils.NewProcsPool(3)

type verify_output_desc struct {
	desc cpt.OutputVerifyDesc
	e    error
}

func (self *verify_output_desc) Run() bool {
	if err := cpt.VerifyOutput(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

func verifyDesc_Zs(tx *stx.T, balance_desc *cpt.BalanceDesc) (e error) {
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
		g.desc.Pkr = out_z.PKr
		g.desc.OutCM = out_z.OutCM
		g.desc.Proof = out_z.Proof

		verify_output_procs.StartProc(&g)
	}
	if i_runs := verify_input_procs.Wait(); i_runs != nil {
		if o_runs := verify_output_procs.Wait(); o_runs != nil {
			return
		} else {
			e = errors.New("verify output desc_z failed!!!")
			return
		}
	} else {
		e = errors.New("verify input desc_z failed!!!")
		return
	}
}
