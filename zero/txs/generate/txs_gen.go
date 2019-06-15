package generate

import (
	"errors"

	"github.com/sero-cash/go-sero/zero/light/light_ref"

	"github.com/sero-cash/go-sero/zero/lstate/lstate_types"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/tx"
)

type gen_ctx struct {
	seed         *keys.Uint256
	t            *tx.T
	p            preTx
	balance_desc cpt.BalanceDesc
	s            stx.T
}

func prepareCtx(seed *keys.Uint256, t *tx.T) (ret gen_ctx, e error) {
	ret.seed = seed
	ret.t = t
	ret.p, e = preGen(t)
	return
}

func (self *gen_ctx) setData() {
	{
		self.s.Ehash = self.t.Ehash
		self.s.Fee = self.t.Fee
		asset_desc := cpt.AssetDesc{
			Tkn_currency: self.t.Fee.Currency,
			Tkn_value:    self.t.Fee.Value.ToUint256(),
			Tkt_category: keys.Empty_Uint256,
			Tkt_value:    keys.Empty_Uint256,
		}
		cpt.GenAssetCC(&asset_desc)
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
	}
	{
		for _, in_o := range self.p.desc_o.ins {
			s_in_o := stx.In_S{}
			s_in_o.Root = in_o.Root
			self.s.Desc_O.Ins = append(self.s.Desc_O.Ins, s_in_o)
			{
				asset := in_o.Out_O.Asset.ToFlatAsset()
				asset_desc := cpt.AssetDesc{
					Tkn_currency: asset.Tkn.Currency,
					Tkn_value:    asset.Tkn.Value.ToUint256(),
					Tkt_category: asset.Tkt.Category,
					Tkt_value:    asset.Tkt.Value,
				}
				cpt.GenAssetCC(&asset_desc)
				self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, asset_desc.Asset_cc[:]...)
			}
		}
	}
	{
		for _, out_o := range self.p.desc_o.outs {
			s_out_o := stx.Out_O{}
			s_out_o.Asset = out_o.Asset.Clone()
			s_out_o.Memo = out_o.Memo
			s_out_o.Addr = out_o.Addr
			self.s.Desc_O.Outs = append(self.s.Desc_O.Outs, s_out_o)
			{
				asset := s_out_o.Asset.ToFlatAsset()
				asset_desc := cpt.AssetDesc{
					Tkn_currency: asset.Tkn.Currency,
					Tkn_value:    asset.Tkn.Value.ToUint256(),
					Tkt_category: asset.Tkt.Category,
					Tkt_value:    asset.Tkt.Value,
				}
				cpt.GenAssetCC(&asset_desc)
				self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
			}
		}
	}
	{
		if self.p.desc_pkg.create != nil {
			create := self.p.desc_pkg.create
			self.s.Desc_Pkg.Create = &stx.PkgCreate{}
			self.s.Desc_Pkg.Create.PKr = create.pkg.PKr
			self.s.Desc_Pkg.Create.Id = create.pkg.Id
		}
		if self.p.desc_pkg.transfer != nil {
			change := self.p.desc_pkg.transfer
			self.s.Desc_Pkg.Transfer = &stx.PkgTransfer{}
			self.s.Desc_Pkg.Transfer.Id = change.zpkg.Pack.Id
			self.s.Desc_Pkg.Transfer.PKr = change.pkr
		}
		if self.p.desc_pkg.close != nil {
			open := self.p.desc_pkg.close
			self.s.Desc_Pkg.Close = &stx.PkgClose{}
			self.s.Desc_Pkg.Close.Id = open.opkg.Z.Pack.Id
			self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, open.opkg.Z.Pack.Pkg.AssetCM[:]...)
			self.balance_desc.Zin_ars = append(self.balance_desc.Zin_ars, open.opkg.O.Ar[:]...)
		}
	}
	{
		for i := 0; i < len(self.p.desc_z.ins); i++ {
			in := stx.In_Z{}
			self.s.Desc_Z.Ins = append(self.s.Desc_Z.Ins, in)
		}
		for _, out_z := range self.p.desc_z.outs {
			out := stx.Out_Z{}
			out.PKr = out_z.Addr
			self.s.Desc_Z.Outs = append(self.s.Desc_Z.Outs, out)
		}
	}
	{
		addr := keys.Seed2Addr(self.seed)
		var from_r keys.Uint256
		if self.t.FromRnd != nil {
			copy(from_r[:], self.t.FromRnd[:])
		} else {
			from_r = keys.RandUint256()
		}
		self.s.From = keys.Addr2PKr(&addr, &from_r)
	}
}

func (self *gen_ctx) proveTx() (e error) {
	if err := genDesc_Zs(self.seed, &self.p, &self.balance_desc, &self.s); err != nil {
		e = err
		return
	} else {
		return
	}
}

func (self *gen_ctx) signTx() (e error) {
	hash_z := self.s.ToHash_for_sign()
	self.balance_desc.Hash = hash_z

	if sign, err := keys.SignPKr(self.seed, &hash_z, &self.s.From); err != nil {
		e = err
		return
	} else {
		self.s.Sign = sign
	}

	for i, s_in_o := range self.p.desc_o.ins {
		g := cpt.InputSDesc{}
		g.Ehash = hash_z
		g.Seed = *self.seed
		g.Pkr = s_in_o.Out_Z.PKr
		g.RootCM = s_in_o.RootCM
		if err := cpt.GenInputSProof(&g); err != nil {
			e = err
			return
		} else {
			in := stx.In_S{}
			in.Sign = g.Sign_ret
			in.Nil = g.Nil_ret
			in.Root = s_in_o.Root
			self.s.Desc_O.Ins[i] = in
		}
	}

	if self.p.desc_pkg.transfer != nil {
		if sign, err := keys.SignPKr(self.seed, &hash_z, &self.p.desc_pkg.transfer.zpkg.Pack.PKr); err != nil {
			e = err
			return
		} else {
			self.s.Desc_Pkg.Transfer.Sign = sign
		}
	}

	if self.p.desc_pkg.close != nil {
		if sign, err := keys.SignPKr(self.seed, &hash_z, &self.p.desc_pkg.close.opkg.Z.Pack.PKr); err != nil {
			e = err
			return
		} else {
			self.s.Desc_Pkg.Close.Sign = sign
		}
	}

	{
		cpt.SignBalance(&self.balance_desc)
		if self.balance_desc.Bcr == keys.Empty_Uint256 {
			e = errors.New("sign balance failed!!!")
			return
		} else {
			self.s.Bcr = self.balance_desc.Bcr
			self.s.Bsign = self.balance_desc.Bsign
		}
	}
	return
}

func Gen(seed *keys.Uint256, t *tx.T) (s stx.T, e error) {
	return Gen_lstate(seed, t)
}

func Gen_lstate(seed *keys.Uint256, t *tx.T) (s stx.T, e error) {
	if ctx, err := prepareCtx(seed, t); err != nil {
		e = err
		return
	} else {
		ctx.setData()
		if e = ctx.proveTx(); e != nil {
			return
		}
		if e = ctx.signTx(); e != nil {
			return
		}
		for _, used_out := range ctx.p.uouts {
			lstate_types.UpdateOutStat(light_ref.Ref_inst.Bc.GetDB(), &used_out)
		}
		s = ctx.s
		return
	}
}
