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

package generate

import (
	"errors"

	"github.com/sero-cash/go-sero/zero/txs/lstate"

	"github.com/sero-cash/go-sero/zero/txs/pkg"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

var gen_pkg_procs_pool = utils.NewProcsPool(func() int { return G_p_thread_num })

type gen_pkg_desc struct {
	desc cpt.PkgDesc
	e    error
}

func (self *gen_pkg_desc) Run() bool {
	if err := cpt.GenPkgProof(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

var gen_input_procs_pool = utils.NewProcsPool(func() int { return G_p_thread_num })

type gen_input_desc struct {
	desc  cpt.InputDesc
	index int
	e     error
}

func (self *gen_input_desc) Run() bool {
	if err := cpt.GenInputProof(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}

var gen_output_procs_pool = utils.NewProcsPool(func() int { return G_p_thread_num })

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

func genDesc_Zs(state *lstate.State, seed *keys.Uint256, ptx *preTx, balance_desc *cpt.BalanceDesc, tx *stx.T) (e error) {
	var gen_pkg_procs = gen_pkg_procs_pool.GetProcs()
	defer gen_pkg_procs_pool.PutProcs(gen_pkg_procs)
	if ptx.desc_pkg.create != nil {
		asset := ptx.desc_pkg.create.pkg.Pkg.Asset.ToFlatAsset()

		g := gen_pkg_desc{}
		g.desc.Tkn_currency = asset.Tkn.Currency
		g.desc.Tkn_value = asset.Tkn.Value.ToUint256()
		g.desc.Tkt_category = asset.Tkt.Category
		g.desc.Tkt_value = asset.Tkt.Value
		g.desc.Memo = ptx.desc_pkg.create.pkg.Pkg.Memo
		g.desc.Key = pkg.GetKey(&tx.From, keys.Seed2Tk(seed).NewRef())

		gen_pkg_procs.StartProc(&g)

	}

	var gen_input_procs = gen_input_procs_pool.GetProcs()
	defer gen_input_procs_pool.PutProcs(gen_input_procs)

	for i, in := range ptx.desc_z.ins {
		g := gen_input_desc{}
		g.desc.Seed = *seed
		g.desc.Pkr = in.Out_Z.PKr
		g.desc.RPK = in.Out_Z.RPK
		g.desc.Einfo = in.Out_Z.EInfo
		g.desc.Index = in.OutIndex
		pos, paths, anchor := state.State.State.MTree.GetPaths(in.RootCM)
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

	for i, out := range ptx.desc_z.outs {
		asset := out.Asset.ToFlatAsset()

		g := gen_output_desc{}
		g.desc.Tkn_currency = asset.Tkn.Currency
		g.desc.Tkn_value = asset.Tkn.Value.ToUint256()
		g.desc.Tkt_category = asset.Tkt.Category
		g.desc.Tkt_value = asset.Tkt.Value
		g.desc.Memo = out.Memo
		g.desc.Pkr = out.Addr
		g.desc.Height = state.State.Num()
		g.index = i

		gen_output_procs.StartProc(&g)
	}

	if gen_pkg_procs.HasProc() {
		if pkg_runs := gen_pkg_procs.Wait(); pkg_runs != nil {
			for _, g := range pkg_runs {
				desc := g.(*gen_pkg_desc).desc

				tx.Desc_Pkg.Create.Proof = desc.Proof_ret
				tx.Desc_Pkg.Create.Pkg.EInfo = desc.Einfo_ret
				tx.Desc_Pkg.Create.Pkg.AssetCM = desc.Asset_cm_ret
				tx.Desc_Pkg.Create.Pkg.PkgCM = desc.Pkg_cm_ret

				balance_desc.Zout_acms = append(balance_desc.Zout_acms, desc.Asset_cm_ret[:]...)
				balance_desc.Zout_ars = append(balance_desc.Zout_ars, desc.Ar_ret[:]...)
				break
			}
		} else {
			e = errors.New("gen output desc_z failed!!!")
			return
		}
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
				tx.Desc_Z.Outs[output_desc.index] = out_z

				balance_desc.Zout_acms = append(balance_desc.Zout_acms, desc.Asset_cm_ret[:]...)
				balance_desc.Zout_ars = append(balance_desc.Zout_ars, desc.Ar_ret[:]...)
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
				tx.Desc_Z.Ins[input_desc.index] = in_z

				balance_desc.Zin_acms = append(balance_desc.Zin_acms, desc.Asset_cm_ret[:]...)
				balance_desc.Zin_ars = append(balance_desc.Zin_ars, desc.Ar_ret[:]...)
			}
		} else {
			e = errors.New("gen input desc_z failed!!!")
			return
		}
	}

	return
}
