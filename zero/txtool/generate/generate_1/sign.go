package generate_1

import (
	"errors"

	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_0"
)

type sign_ctx struct {
	param        txtool.GTxParam
	p0_ins       []*txtool.GIn
	p_ins        []*txtool.GIn
	c_ins        []*txtool.GIn
	balance_desc c_type.BalanceDesc
	keys         []c_type.Uint256
	s            stx.T
}

func (self *sign_ctx) Tx() (ret stx.T) {
	ret = self.s
	return
}

func (self *sign_ctx) Param() (ret txtool.GTxParam) {
	ret = self.param
	return
}

func (self *sign_ctx) Keys() (ret []c_type.Uint256) {
	ret = self.keys
	return
}

func SignTx(param *txtool.GTxParam) (ctx sign_ctx, e error) {
	ctx.param = *param
	if e = ctx.check(); e != nil {
		return
	}
	if e = ctx.prepare(); e != nil {
		return
	}
	if e = ctx.genFrom(); e != nil {
		return
	}
	if e = ctx.genFee(); e != nil {
		return
	}
	if e = ctx.genCmd(); e != nil {
		return
	}
	if e = ctx.genInsP0(); e != nil {
		return
	}
	if e = ctx.genInsP(); e != nil {
		return
	}
	if e = ctx.genInsC(); e != nil {
		return
	}
	if e = ctx.genOutsC(); e != nil {
		return
	}
	if e = ctx.genSign(); e != nil {
		return
	}
	return
}

func (self *sign_ctx) check() (e error) {
	if !self.param.IsSzk() {
		e = errors.New("param is czero, must use generate_0")
		return
	}
	sk := self.param.From.SKr.ToUint512()
	tk := superzk.Sk2Tk(&sk)
	if !superzk.IsMyPKr(&tk, &self.param.From.PKr) {
		e = errors.New("sk unmatch pkr for the From field")
		return
	}

	for _, in := range self.param.Ins {
		tk := superzk.Sk2Tk(in.SKr.ToUint512().NewRef())
		if in.Out.State.OS.Out_O != nil {
			if !c_superzk.CzeroIsMyPKr(&tk, &in.Out.State.OS.Out_O.Addr) {
				e = errors.New("sk unmatch pkr for out_o")
				return
			}
			continue
		}
		if in.Out.State.OS.Out_Z != nil {
			if !c_superzk.CzeroIsMyPKr(&tk, &in.Out.State.OS.Out_Z.PKr) {
				e = errors.New("sk unmatch pkr for out_z")
				return
			}
			continue
		}
		if in.Out.State.OS.Out_P != nil {
			if !superzk.IsMyPKr(&tk, &in.Out.State.OS.Out_P.PKr) {
				e = errors.New("sk unmatch pkr for out_z")
				return
			}
			continue
		}
		if in.Out.State.OS.Out_C != nil {
			if !superzk.IsMyPKr(&tk, &in.Out.State.OS.Out_C.PKr) {
				e = errors.New("sk unmatch pkr for out_z")
				return
			}
			continue
		}
	}
	for _, out := range self.param.Outs {
		if !c_superzk.IsSzkPKr(&out.PKr) {
			e = errors.New("output must be Szk pkr")
			return
		}
	}
	if self.param.Cmds.BuyShare != nil {
		if !c_superzk.IsSzkPKr(&self.param.Cmds.BuyShare.Vote) {
			e = errors.New("buyshare must be Szk pkr")
			return
		}
	}
	if self.param.Cmds.RegistPool != nil {
		if !c_superzk.IsSzkPKr(&self.param.Cmds.RegistPool.Vote) {
			e = errors.New("regist pool must be Szk pkr")
			return
		}
	}
	return
}

func (self *sign_ctx) prepare() (e error) {
	self.s.Tx1 = &stx_v1.Tx{}
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
			self.c_ins = append(self.c_ins, in)
			continue
		}
	}
	return
}

func (self *sign_ctx) genFrom() (e error) {
	self.s.From = self.param.From.PKr
	return
}

func (self *sign_ctx) genFee() (e error) {
	{
		self.s.Fee = self.param.Fee

		asset_desc := c_superzk.AssetDesc{
			Asset: self.param.Fee.ToTypeAsset(),
		}
		if e = c_superzk.GenAssetCC(&asset_desc); e != nil {
			return
		}
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc_ret[:]...)
	}
	return
}

func (self *sign_ctx) genCmd() (e error) {
	var a *assets.Asset
	if self.param.Cmds.BuyShare != nil {
		self.s.Desc_Cmd.BuyShare = self.param.Cmds.BuyShare
		asset := self.param.Cmds.BuyShare.Asset()
		a = &asset
	}
	if self.param.Cmds.RegistPool != nil {
		self.s.Desc_Cmd.RegistPool = self.param.Cmds.RegistPool
		asset := self.param.Cmds.RegistPool.Asset()
		a = &asset
	}
	if self.param.Cmds.ClosePool != nil {
		self.s.Desc_Cmd.ClosePool = self.param.Cmds.ClosePool
	}
	if self.param.Cmds.Contract != nil {
		self.s.Desc_Cmd.Contract = self.param.Cmds.Contract
		a = &self.param.Cmds.Contract.Asset
	}
	if self.param.Cmds.PkgCreate != nil {
		create := self.param.Cmds.PkgCreate
		self.s.Desc_Pkg.Create = &stx.PkgCreate{}
		self.s.Desc_Pkg.Create.PKr = create.PKr
		self.s.Desc_Pkg.Create.Id = create.Id
	}
	if self.param.Cmds.PkgTransfer != nil {
		change := self.param.Cmds.PkgTransfer
		self.s.Desc_Pkg.Transfer = &stx.PkgTransfer{}
		self.s.Desc_Pkg.Transfer.Id = change.Id
		self.s.Desc_Pkg.Transfer.PKr = change.PKr
	}
	if self.param.Cmds.PkgClose != nil {
		close := self.param.Cmds.PkgClose
		self.s.Desc_Pkg.Close = &stx.PkgClose{}
		self.s.Desc_Pkg.Close.Id = close.Id
		self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, close.AssetCM[:]...)
		self.balance_desc.Zin_ars = append(self.balance_desc.Zin_ars, close.Ar[:]...)
	}
	if a != nil {
		if asset_desc, err := GenAssetCC(a); err != nil {
			e = err
			return
		} else {
			self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc_ret[:]...)
		}
	}
	return
}

func (self *sign_ctx) genInsP0() (e error) {
	for _, in := range self.p0_ins {
		sk := in.SKr.ToUint512()
		tk := superzk.Sk2Tk(&sk)

		t_in := stx_v1.In_P0{}
		t_in.Root = in.Out.Root
		t_in.Trace = c_superzk.GenCzeroTrace(&tk, in.Out.State.OS.RootCM)
		t_in.Nil = c_superzk.GenCzeroNil(&sk, in.Out.State.OS.RootCM)

		if in.Out.State.OS.Out_O != nil {
			if asset_desc, err := GenAssetCC(&in.Out.State.OS.Out_O.Asset); err != nil {
				e = err
				return
			} else {
				self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, asset_desc.Asset_cc_ret[:]...)
			}
		} else {
			key, flag := c_superzk.FetchCzeroKey(&tk, &in.Out.State.OS.Out_Z.RPK)
			t_in.Key = &key
			if out := generate_0.ConfirmOutZ(&key, flag, in.Out.State.OS.Out_Z); out == nil {
				e = errors.New("gen tx1 confirm outz error")
				return
			} else {
				if asset_desc, err := GenAssetCC(&out.Asset); err != nil {
					e = err
				} else {
					self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, asset_desc.Asset_cc_ret[:]...)
				}
			}
		}
		self.s.Tx1.Ins_P0 = append(self.s.Tx1.Ins_P0, t_in)
	}
	return
}

func (self *sign_ctx) genInsP() (e error) {
	for _, in := range self.p_ins {
		sk := in.SKr.ToUint512()
		tk := superzk.Sk2Tk(&sk)

		t_in := stx_v1.In_P{}
		t_in.Root = in.Out.Root
		t_in.Nil, e = c_superzk.GenNil(&tk, in.Out.State.OS.RootCM, in.Out.State.OS.ToPKr())
		if e != nil {
			return
		}
		if asset_desc, err := GenAssetCC(&in.Out.State.OS.Out_P.Asset); err != nil {
			e = err
			return
		} else {
			self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, asset_desc.Asset_cc_ret[:]...)
			self.s.Tx1.Ins_P = append(self.s.Tx1.Ins_P, t_in)
		}
	}
	return
}

func (self *sign_ctx) genInsC() (e error) {
	for _, in := range self.c_ins {
		sk := in.SKr.ToUint512()
		tk := superzk.Sk2Tk(&sk)

		t_in := stx_v1.In_C{}

		t_in.Nil, e = c_superzk.GenNil(&tk, in.Out.State.OS.RootCM, in.Out.State.OS.ToPKr())
		if e != nil {
			return
		}

		key, err := c_superzk.FetchKey(&in.Out.State.OS.Out_C.PKr, &tk, &in.Out.State.OS.Out_C.RPK)
		if err != nil {
			e = err
			return
		}

		info := c_superzk.DecInfoDesc{
			Key:   key,
			Einfo: in.Out.State.OS.Out_C.EInfo,
		}
		if e = c_superzk.DecOutput(&info); e != nil {
			return
		}

		self.keys = append(self.keys, key)

		asset_desc := c_superzk.AssetDesc{
			Asset: assets.NewAssetByType(&info.Asset_ret).ToRef().ToTypeAsset(),
			Ar:    c_superzk.RandomFr(),
		}
		if e = c_superzk.GenAssetCM(&asset_desc); e != nil {
			return
		}
		t_in.AssetCM = asset_desc.Asset_cm_ret
		in.A = &asset_desc.Ar

		zpka, a, err := c_superzk.GenZPKa(&in.Out.State.OS.Out_C.PKr)
		if err != nil {
			e = err
			return
		}
		t_in.ZPKa = zpka
		in.A = &a

		self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, asset_desc.Asset_cm_ret[:]...)
		self.balance_desc.Zin_ars = append(self.balance_desc.Zin_ars, asset_desc.Ar[:]...)
		self.s.Tx1.Ins_C = append(self.s.Tx1.Ins_C, t_in)
	}
	return
}

func (self *sign_ctx) genOutsC() (e error) {
	for _, out := range self.param.Outs {
		t_out := stx_v1.Out_C{}

		asset_desc := c_superzk.AssetDesc{
			Asset: out.Asset.ToTypeAsset(),
			Ar:    c_superzk.RandomFr(),
		}
		if e = c_superzk.GenAssetCM(&asset_desc); e != nil {
			return
		}
		t_out.AssetCM = asset_desc.Asset_cm_ret
		out.Ar = &asset_desc.Ar

		t_out.PKr = out.PKr
		key, rpk, _, err := c_superzk.GenKey(&out.PKr)
		if err != nil {
			e = err
			return
		}
		t_out.RPK = rpk

		info_desc := c_superzk.EncInfoDesc{
			Key:   key,
			Memo:  out.Memo,
			Asset: out.Asset.ToTypeAsset(),
			Ar:    asset_desc.Ar,
		}
		if e = c_superzk.EncOutput(&info_desc); e != nil {
			return
		}
		t_out.EInfo = info_desc.Einfo

		self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, asset_desc.Asset_cm_ret[:]...)
		self.balance_desc.Zout_ars = append(self.balance_desc.Zout_ars, asset_desc.Ar[:]...)
		self.s.Tx1.Outs_C = append(self.s.Tx1.Outs_C, t_out)
	}
	return
}

func (self *sign_ctx) genSign() (e error) {

	self.balance_desc.Hash = self.s.Tx1_Hash()

	if e = self.signFrom(); e != nil {
		return
	}
	if e = self.signInsP0(); e != nil {
		return
	}
	if e = self.signInsP(); e != nil {
		return
	}
	if e = self.signInsC(); e != nil {
		return
	}
	if e = self.signBalance(); e != nil {
		return
	}
	return
}

func (self sign_ctx) signFrom() (e error) {
	if sign, err := superzk.SignPKrBySk(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.s.From); err != nil {
		return err
	} else {
		self.s.Sign = sign
		return nil
	}
	return
}
func (self sign_ctx) signInsP0() (e error) {
	for i := range self.s.Tx1.Ins_P0 {
		t_in := self.p0_ins[i]
		if sign, err := c_superzk.SignCzeroNil(
			t_in.SKr.ToUint512().NewRef(),
			&self.balance_desc.Hash,
			t_in.Out.State.OS.RootCM.NewRef(),
			t_in.Out.State.OS.ToPKr(),
		); err != nil {
			e = err
			return
		} else {
			self.s.Tx1.Ins_P0[i].Sign = sign
		}
	}
	return
}
func (self sign_ctx) signInsP() (e error) {
	for i := range self.s.Tx1.Ins_P {
		t_in := self.p_ins[i]
		if sign, err := superzk.SignPKrBySk(t_in.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, t_in.Out.State.OS.ToPKr()); err != nil {
			return err
		} else {
			self.s.Tx1.Ins_P[i].ASign = sign
		}
		tk := superzk.Sk2Tk(t_in.SKr.ToUint512().NewRef())
		if sign, err := c_superzk.SignNil(
			&tk,
			&self.balance_desc.Hash,
			t_in.Out.State.OS.RootCM.NewRef(),
			t_in.Out.State.OS.ToPKr(),
		); err != nil {
			e = err
			return
		} else {
			self.s.Tx1.Ins_P[i].NSign = sign
		}
	}
	return
}
func (self sign_ctx) signInsC() (e error) {
	for i := range self.s.Tx1.Ins_C {
		t_in := self.c_ins[i]
		if sign, err := c_superzk.SignZPKa(t_in.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, t_in.A, t_in.Out.State.OS.ToPKr()); err != nil {
			e = err
			return
		} else {
			self.s.Tx1.Ins_C[i].Sign = sign
		}
	}
	return
}
func (self sign_ctx) signBalance() (e error) {
	if e = c_superzk.SignBalance(&self.balance_desc); e != nil {
		return
	}
	if self.balance_desc.Bcr == c_type.Empty_Uint256 {
		return errors.New("sign balance failed!!!")
	} else {
		self.s.Bcr = self.balance_desc.Bcr
		self.s.Bsign = self.balance_desc.Bsign
		return nil
	}
}
