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

package txtool

import (
	"errors"

	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

var gen_input_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_p_thread_num })

type gen_input_desc struct {
	desc  cpt.InputDesc
	index int
	e     error
}

func (self *gen_input_desc) Run() bool {
	if err := cpt.GenInputProofBySk(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

var gen_output_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_p_thread_num })

type gen_output_desc struct {
	desc  cpt.OutputDesc
	index int
	e     error
}

func (self *gen_output_desc) Run() bool {
	if err := cpt.GenOutputProof(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

func (self *gen_ctx) genDesc_Zs() (e error) {
	var gen_input_procs = gen_input_procs_pool.GetProcs()
	defer gen_input_procs_pool.PutProcs(gen_input_procs)

	for i, in := range self.Z_Ins {
		g := gen_input_desc{}
		g.desc.Sk = in.SKr.ToUint512()
		g.desc.Pkr = in.Out.State.OS.Out_Z.PKr
		g.desc.RPK = in.Out.State.OS.Out_Z.RPK
		g.desc.Einfo = in.Out.State.OS.Out_Z.EInfo
		g.desc.Index = in.Out.State.OS.Index
		pos, paths, anchor := in.Witness.Pos, in.Witness.Paths, in.Witness.Anchor
		g.desc.Position = uint32(pos)
		g.desc.Anchor = anchor
		for i, path := range paths {
			copy(g.desc.Path[len(g.desc.Path)-32-(i*32):], path[:])
		}

		g.index = i
		gen_input_procs.StartProc(&g)
	}

	var gen_output_procs = gen_output_procs_pool.GetProcs()
	defer gen_output_procs_pool.PutProcs(gen_output_procs)

	for i, out := range self.param.Outs {
		asset := out.Asset.ToFlatAsset()

		g := gen_output_desc{}
		g.desc.Tkn_currency = asset.Tkn.Currency
		g.desc.Tkn_value = asset.Tkn.Value.ToUint256()
		g.desc.Tkt_category = asset.Tkt.Category
		g.desc.Tkt_value = asset.Tkt.Value
		g.desc.Memo = out.Memo
		g.desc.Pkr = out.PKr
		g.desc.Height = 606007
		g.index = i

		gen_output_procs.StartProc(&g)
	}

	if gen_output_procs.HasProc() {
		if o_runs := gen_output_procs.Wait(); o_runs != nil {
			for _, g := range o_runs {
				output_desc := g.(*gen_output_desc)
				desc := output_desc.desc
				out_z := stx.Out_Z{}
				out_z.AssetCM = desc.Asset_cm_ret
				out_z.OutCM = desc.Out_cm_ret
				out_z.RPK = desc.RPK_ret
				out_z.EInfo = desc.Einfo_ret
				out_z.Proof = desc.Proof_ret
				out_z.PKr = desc.Pkr
				self.s.Desc_Z.Outs[output_desc.index] = out_z

				self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, desc.Asset_cm_ret[:]...)
				self.balance_desc.Zout_ars = append(self.balance_desc.Zout_ars, desc.Ar_ret[:]...)
			}
		} else {
			e = errors.New("gen output desc_z failed!!!")
			return
		}
	}

	if gen_input_procs.HasProc() {
		if i_runs := gen_input_procs.Wait(); i_runs != nil {
			for _, g := range i_runs {
				input_desc := g.(*gen_input_desc)
				desc := input_desc.desc
				in_z := stx.In_Z{}
				in_z.Anchor = desc.Anchor
				in_z.AssetCM = desc.Asset_cm_ret
				in_z.Nil = desc.Nil_ret
				in_z.Trace = desc.Til_ret
				in_z.Proof = desc.Proof_ret
				self.s.Desc_Z.Ins[input_desc.index] = in_z

				self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, desc.Asset_cm_ret[:]...)
				self.balance_desc.Zin_ars = append(self.balance_desc.Zin_ars, desc.Ar_ret[:]...)
			}
		} else {
			e = errors.New("gen input desc_z failed!!!")
			return
		}
	}

	return
}
