package generate_0

import (
	"errors"
	"fmt"

	"github.com/sero-cash/go-czero-import/c_czero"

	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
)

type gen_ctx struct {
	param        txtool.GTxParam
	O_Ins        []txtool.GIn
	Z_Ins        []txtool.GIn
	balance_desc c_type.BalanceDesc
	Keys         []c_type.Uint256
	Bases        []c_type.Uint256
	s            stx.T
}

func GenTx(param *txtool.GTxParam) (ret stx.T, keys []c_type.Uint256, Bases []c_type.Uint256, e error) {
	if param.IsSzk() {
		e = errors.New("param is szk, must use gnerate_1")
		return
	}
	ctx := gen_ctx{}
	ctx.param = *param
	ctx.prepare()
	ctx.setData()
	if err := ctx.check(); err != nil {
		e = err
		return
	}
	if err := ctx.genDesc_Zs(); err != nil {
		e = err
		return
	}
	if err := ctx.signTx(); err != nil {
		e = err
		return
	}
	ret = ctx.s
	keys = ctx.Keys
	Bases = ctx.Bases
	return
}

func (self *gen_ctx) prepare() {
	for _, in := range self.param.Ins {
		if in.Out.State.OS.Out_O != nil {
			self.O_Ins = append(self.O_Ins, in)
			continue
		}
		if in.Out.State.OS.Out_Z != nil {
			self.Z_Ins = append(self.Z_Ins, in)
			continue
		}
	}
	self.s.Tx0 = &stx_v1.Tx{}
}

func (self *gen_ctx) check() (e error) {
	sk := self.param.From.SKr.ToUint512()
	tk := c_czero.Sk2Tk(&sk)
	if !c_czero.IsMyPKr(&tk, &self.param.From.PKr) {
		e = errors.New("sk unmatch pkr for the From field")
		return
	}
	return
}

func (self *gen_ctx) setFeeData() {
	{
		self.s.Fee = assets.Token{
			self.param.Fee.Currency,
			self.param.Fee.Value,
		}
		asset_desc := c_czero.AssetDesc{
			Tkn_currency: self.s.Fee.Currency,
			Tkn_value:    self.s.Fee.Value.ToUint256(),
			Tkt_category: c_type.Empty_Uint256,
			Tkt_value:    c_type.Empty_Uint256,
		}
		c_czero.GenAssetCC(&asset_desc)
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
	}
}

func (self *gen_ctx) setInsData() {
	{
		for _, in := range self.param.Ins {
			if in.Out.State.OS.Out_O != nil {
				s_in_o := stx_v1.In_S{}
				s_in_o.Root = in.Out.Root
				self.s.Tx0.Desc_O.Ins = append(self.s.Tx0.Desc_O.Ins, s_in_o)
				{
					asset := in.Out.State.OS.Out_O.Asset.ToFlatAsset()
					asset_desc := c_czero.AssetDesc{
						Tkn_currency: asset.Tkn.Currency,
						Tkn_value:    asset.Tkn.Value.ToUint256(),
						Tkt_category: asset.Tkt.Category,
						Tkt_value:    asset.Tkt.Value,
					}
					c_czero.GenAssetCC(&asset_desc)
					self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, asset_desc.Asset_cc[:]...)
				}
			} else {
				in := stx_v1.In_Z{}
				self.s.Tx0.Desc_Z.Ins = append(self.s.Tx0.Desc_Z.Ins, in)
			}
		}
	}
}

func (self *gen_ctx) setOutsData() {
	{
		for _, out_z := range self.param.Outs {
			out := stx_v1.Out_Z{}
			out.PKr = out_z.PKr
			self.s.Tx0.Desc_Z.Outs = append(self.s.Tx0.Desc_Z.Outs, out)
		}
	}
}

func (self *gen_ctx) setCmdsData() {
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
		asset := a.ToFlatAsset()
		asset_desc := c_czero.AssetDesc{
			Tkn_currency: asset.Tkn.Currency,
			Tkn_value:    asset.Tkn.Value.ToUint256(),
			Tkt_category: asset.Tkt.Category,
			Tkt_value:    asset.Tkt.Value,
		}
		c_czero.GenAssetCC(&asset_desc)
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
	}
}

func (self *gen_ctx) setData() {
	self.s.Ehash = types.Ehash(*self.param.GasPrice, self.param.Gas, []byte{})
	self.setFeeData()
	self.setInsData()
	self.setOutsData()
	self.setCmdsData()
	self.s.From = self.param.From.PKr
}

func (self *gen_ctx) signTxFrom() error {
	sk := self.param.From.SKr.ToUint512()
	tk := c_czero.Sk2Tk(&sk)
	if !c_czero.IsMyPKr(&tk, &self.s.From) {
		return fmt.Errorf("sign from : sk unmatch the from (%v)", hexutil.Encode(self.s.From[:]))
	}
	if sign, err := c_czero.SignPKrBySk(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.s.From); err != nil {
		return err
	} else {
		self.s.Sign = sign
		return nil
	}
}

func (self *gen_ctx) signTxIns() error {
	for i, s_in_o := range self.O_Ins {
		g := c_czero.InputSDesc{}
		g.Ehash = self.balance_desc.Hash
		g.Sk = s_in_o.SKr.ToUint512()
		g.Pkr = s_in_o.Out.State.OS.Out_O.Addr
		g.RootCM = *s_in_o.Out.State.OS.RootCM
		if err := c_czero.GenInputSProofBySk(&g); err != nil {
			return err
		} else {
			in := stx_v1.In_S{}
			in.Sign = g.Sign_ret
			in.Nil = g.Nil_ret
			in.Root = s_in_o.Out.Root
			self.s.Tx0.Desc_O.Ins[i] = in
		}
	}
	return nil
}

func (self *gen_ctx) signTxBalance() error {
	{
		c_czero.SignBalance(&self.balance_desc)
		if self.balance_desc.Bcr == c_type.Empty_Uint256 {
			return errors.New("sign balance failed!!!")
		} else {
			self.s.Bcr = self.balance_desc.Bcr
			self.s.Bsign = self.balance_desc.Bsign
			return nil
		}
	}
}

func (self *gen_ctx) signTxCmds() error {
	if self.param.Cmds.PkgTransfer != nil {
		if sign, err := c_czero.SignPKrBySk(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.param.Cmds.PkgTransfer.Owner); err != nil {
			return err
		} else {
			self.s.Desc_Pkg.Transfer.Sign = sign
		}
	}
	if self.param.Cmds.PkgClose != nil {
		if sign, err := c_czero.SignPKrBySk(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.param.Cmds.PkgClose.Owner); err != nil {
			return err
		} else {
			self.s.Desc_Pkg.Transfer.Sign = sign
		}
	}
	return nil
}

func (self *gen_ctx) signTx() (e error) {
	self.balance_desc.Hash = self.s.ToHash_for_sign()

	if e = self.signTxFrom(); e != nil {
		return
	}
	if e = self.signTxIns(); e != nil {
		return
	}
	if e = self.signTxCmds(); e != nil {
		return
	}
	if e = self.signTxBalance(); e != nil {
		return
	}

	return
}
