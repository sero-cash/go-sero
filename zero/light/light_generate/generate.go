package light_generate

import (
	"errors"
	"math/big"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type gen_ctx struct {
	param        light_types.GenTxParam
	O_Ins        []light_types.GIn
	Z_Ins        []light_types.GIn
	balance_desc cpt.BalanceDesc
	s            stx.T
}

func Generate(param *light_types.GenTxParam) (ret stx.T, e error) {
	ctx := gen_ctx{}
	ctx.param = *param
	ctx.prepare()
	ctx.setData()
	if err := ctx.genDesc_Zs(); err != nil {
		e = err
		return
	}
	if err := ctx.signTx(); err != nil {
		e = err
		return
	}
	return
}

func (self *gen_ctx) prepare() {
	for _, in := range self.param.Ins {
		if in.Out.Out_O != nil {
			self.O_Ins = append(self.O_Ins, in)
		} else {
			self.Z_Ins = append(self.Z_Ins, in)
		}
	}
}

func (self *gen_ctx) setData() {

	self.s.Ehash = types.Ehash(self.param.GasPrice, self.param.Gas, []byte{})

	{
		self.s.Fee = assets.Token{
			utils.StringToUint256("SERO"),
			utils.U256(*big.NewInt(0).Mul(&self.param.GasPrice, big.NewInt(int64(self.param.Gas)))),
		}
		asset_desc := cpt.AssetDesc{
			Tkn_currency: self.s.Fee.Currency,
			Tkn_value:    self.s.Fee.Value.ToUint256(),
			Tkt_category: keys.Empty_Uint256,
			Tkt_value:    keys.Empty_Uint256,
		}
		cpt.GenAssetCC(&asset_desc)
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, asset_desc.Asset_cc[:]...)
	}
	{
		for _, in := range self.param.Ins {
			if in.Out.Out_O != nil {
				s_in_o := stx.In_S{}
				s_in_o.Root = in.Out.Root
				self.s.Desc_O.Ins = append(self.s.Desc_O.Ins, s_in_o)
				{
					asset := in.Out.Out_O.Asset.ToFlatAsset()
					asset_desc := cpt.AssetDesc{
						Tkn_currency: asset.Tkn.Currency,
						Tkn_value:    asset.Tkn.Value.ToUint256(),
						Tkt_category: asset.Tkt.Category,
						Tkt_value:    asset.Tkt.Value,
					}
					cpt.GenAssetCC(&asset_desc)
					self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, asset_desc.Asset_cc[:]...)
				}
			} else {
				in := stx.In_Z{}
				self.s.Desc_Z.Ins = append(self.s.Desc_Z.Ins, in)
			}
		}
	}
	{
		for _, out_z := range self.param.Outs {
			out := stx.Out_Z{}
			out.PKr = out_z.PKr
			self.s.Desc_Z.Outs = append(self.s.Desc_Z.Outs, out)
		}
	}
	self.s.From = self.param.From.PKr
}

func (self *gen_ctx) signTx() (e error) {
	hash_z := self.s.ToHash_for_sign()
	self.balance_desc.Hash = hash_z

	if sign, err := keys.SignPKrBySk(self.param.From.SKr.ToUint512().NewRef(), &hash_z, &self.s.From); err != nil {
		e = err
		return
	} else {
		self.s.Sign = sign
	}

	for i, s_in_o := range self.O_Ins {
		g := cpt.InputSDesc{}
		g.Ehash = hash_z
		g.Sk = s_in_o.SKr.ToUint512()
		g.Pkr = s_in_o.Out.Out_O.Addr
		g.RootCM = *s_in_o.Out.RootCM
		if err := cpt.GenInputSProofBySk(&g); err != nil {
			e = err
			return
		} else {
			in := stx.In_S{}
			in.Sign = g.Sign_ret
			in.Nil = g.Nil_ret
			in.Root = s_in_o.Out.Root
			self.s.Desc_O.Ins[i] = in
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
