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
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
)

type genParams struct {
	common cpt.Common
	pre    cpt.Pre
	extra  cpt.Extra
	out    cpt.Out
	in     cpt.In
	proof  cpt.Proof
	result stx.Desc_Z
}

var gen_cache = make(chan int, 4)

func genDesc_Zs(seed *keys.Uint256, ptx *preTx, hash_o *keys.Uint256, balance_desc *cpt.BalanceDesc) (desc_z stx.Desc_Z, e error) {
	for _, in := range ptx.desc_z.ins {
		desc := cpt.InputDesc{}
		desc.Seed = *seed
		desc.Pkr = in.Out_Z.PKr
		desc.Einfo = in.Out_Z.EInfo
		//--
		desc.Index = in.OutIndex
		_, desc.Position, desc.Path, desc.Anchor = in.ToWitness()

		if err := cpt.GenInputProof(&desc); err != nil {
			e = err
			return
		}

		in_z := stx.In_Z{}
		in_z.Anchor = desc.Anchor
		in_z.AssetCM = desc.Asset_cm_ret
		in_z.Nil = desc.Nil_ret
		in_z.Trace = desc.Til_ret
		desc_z.Ins = append(desc_z.Ins, in_z)
		balance_desc.Zin_acms = append(balance_desc.Zin_acms, desc.Asset_cm_ret[:]...)
		balance_desc.Zin_ars = append(balance_desc.Zin_ars, desc.Ar_ret[:]...)
	}

	for _, out := range ptx.desc_z.outs {
		desc := cpt.OutputDesc{}
		desc.Seed = *seed
		asset := out.Asset.ToCompleteAsset()
		desc.Tkn_currency = asset.Tkn.Currency
		desc.Tkn_value = asset.Tkn.Value.ToUint256()
		desc.Tkt_category = asset.Tkt.Category
		desc.Tkt_value = asset.Tkt.Value
		desc.Memo = out.Memo
		desc.Pk = out.Addr
		if err := cpt.GenOutputProof(&desc); err != nil {
			e = err
			return
		} else {
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
	}

	return

	/*extras := []cpt.Extra{}

	z := [2]utils.U256{}

	for i, desc := range ptx.desc_zs {
		extra := cpt.Extra{
			cpt.Pre{
				uint32(i),
				cpt.Random(),
				[2]keys.Uint256{},
			},
			[2]keys.Uint256{},
			keys.Uint256{},
		}

		for i, currency := range ptx.C2I.currencys {
			z_i := desc.z2z.get(&currency)
			z_i.AddU(&z[i])
			extra.Z[i] = z_i.ToUint256()
		}

		if i == len(ptx.desc_zs)-1 {
			for i, currency := range ptx.C2I.currencys {
				if desc_o, ok := ptx.desc_os[currency]; ok {
					extra.O[i] = desc_o.z2o.ToUint256()
				}
			}
		}
		extras = append(extras, extra)
	}

	tr := utils.TR_enter("Verify DescZs")

	var wg sync.WaitGroup
	result := true

	log.Info("Generate desc_z: Batchs ", "num", (len(ptx.desc_zs)-1)/4+1)
	t_desc_zs := []*stx.Desc_Z{}
	for i, desc := range ptx.desc_zs {
		params := genParams{}

		params.common.Seed = *seed
		params.common.Hash_O = *hash_o
		params.common.Currency = desc.currency

		for i, currency := range ptx.C2I.currencys {
			params.common.C[i] = currency
		}

		if i == 0 {
			params.pre = extras[len(ptx.desc_zs)-1].Pre
		} else {
			params.pre = extras[i-1].Pre
		}

		params.extra = extras[i]

		if desc.out != nil {
			params.out.Currency = desc.currency
			params.out.Addr = desc.out.Addr
			params.out.Value = desc.out.Value.ToUint256()
			params.out.Info = desc.out.Memo
		}

		if desc.in != nil {
			params.in.Commitment, params.in.Index, params.in.Path, params.in.Anchor = desc.in.ToWitness()
			params.in.S1 = desc.in.Desc_Z.S1
			params.in.EText = desc.in.Desc_Z.Out.EInfo
		} else {
			params.in.Anchor = ptx.last_anchor
		}

		t_desc_zs = append(t_desc_zs, &stx.Desc_Z{})
		t_desc_z := t_desc_zs[len(t_desc_zs)-1]

		wg.Add(1)

		t_desc := desc
		go func(params *genParams, desc_z *stx.Desc_Z, pre_desc *preTxDesc_Z) {

			gen_cache <- 0
			defer func() {
				<-gen_cache
				wg.Done()
			}()

			if e := cpt.GenDesc_Z(&params.common, &params.pre, &params.extra, &params.out, &params.in, &params.proof); e == nil {
				desc_z.S1 = params.extra.S1_ret
				desc_z.R = params.extra.R
				desc_z.Proof.G = params.proof.G
				desc_z.In.Nil = params.in.Nil_ret
				desc_z.In.Trace = params.in.Trace_ret
				desc_z.In.Anchor = params.in.Anchor
				desc_z.Out.Commitment = params.out.Commitment_ret
				desc_z.Out.EInfo = params.out.EText_ret
				desc_z.Proof.G = params.proof.G
				if pre_desc.in != nil {
					if pre_desc.in.Trace != params.in.Trace_ret {
						panic("pre desc in trace != param in nil")
					}
				}
			} else {
				result = false
			}

		}(&params, t_desc_z, &t_desc)

	}

	wg.Wait()

	if !result {
		e = errors.New("gen desc_z failed!!!")
		return
	}

	for _, desc_z := range t_desc_zs {
		desc_zs = append(desc_zs, *desc_z)
	}

	tr.Leave()
	return*/
}

var ver_cache = make(chan int, 4)

func verifyDesc_Zs(tx *stx.T, balance_desc *cpt.BalanceDesc) (e error) {
	for _, in_z := range tx.Desc_Z.Ins {
		balance_desc.Zin_acms = append(balance_desc.Zin_acms, in_z.AssetCM[:]...)

		desc := cpt.InputVerifyDesc{}
		desc.Nil = in_z.Nil
		desc.Anchor = in_z.Anchor
		desc.AssetCM = in_z.AssetCM
		desc.Proof = in_z.Proof

		if err := cpt.VerifyInput(&desc); err != nil {
			e = err
			return
		}
	}
	for _, out_z := range tx.Desc_Z.Outs {
		balance_desc.Zout_acms = append(balance_desc.Zout_acms, out_z.AssetCM[:]...)

		desc := cpt.OutputVerifyDesc{}
		desc.AssetCM = out_z.AssetCM
		desc.Pkr = out_z.PKr
		desc.OutCM = out_z.OutCM
		desc.Proof = out_z.Proof

		if err := cpt.VerifyOutput(&desc); err != nil {
			e = err
			return
		}
	}
	return
	/*tr := utils.TR_enter("Verify DescZs")

	var wg sync.WaitGroup
	result := true

	hash_o := tx.ToHash_for_z()
	for i, desc := range tx.Desc_Zs {
		params := verParams{}

		params.hash_o = hash_o

		if i == 0 {
			params.pre.I = uint32(len(tx.Desc_Zs) - 1)
		} else {
			params.pre.I = uint32(i - 1)
		}
		params.pre.S1 = tx.Desc_Zs[params.pre.I].S1

		params.extra.S1 = desc.S1
		params.extra.I = uint32(i)
		if i == len(tx.Desc_Zs)-1 {
			for _, desc_o := range tx.Desc_Os {
				params.extra.O[desc_o.Z2OIndex] = desc_o.Z2O.ToUint256()
			}
		} else {
			params.extra.O[0] = keys.Uint256{}
			params.extra.O[1] = keys.Uint256{}
		}

		params.out.Commitment = desc.Out.Commitment

		params.in.Anchor = desc.In.Anchor
		params.in.Nil = desc.In.Nil

		params.proof.G = desc.Proof.G

		wg.Add(1)

		go func(params verParams) {
			ver_cache <- 0
			defer func() {
				<-ver_cache
				wg.Done()
			}()
			if err := cpt.VerifyDesc_Z(
				&params.hash_o,
				&params.pre,
				&params.extra,
				&params.out,
				&params.in,
				&params.proof); err != nil {

				result = false
				e = err
				return
			}
		}(params)
	}

	wg.Wait()

	if !result {
		e = errors.New("verify desc_z failed!!!")
		return
	}

	tr.Leave()
	return*/
}
