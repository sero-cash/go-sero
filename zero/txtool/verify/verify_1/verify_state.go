package verify_1

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txtool/verify/verify_utils"
)

type verifyWithStateCtx struct {
	tx           *stx.T
	state        *zstate.ZState
	balance_desc c_type.BalanceDesc
	ck           assets.CKState
}

func VerifyWithState(tx *stx.T, state *zstate.ZState) (e error) {
	ctx := verifyWithStateCtx{}
	ctx.tx = tx
	ctx.state = state
	return ctx.verify()
}

func (self *verifyWithStateCtx) prepare() {
	self.balance_desc.Hash = self.tx.Tx1_Hash()
	return
}

func (self *verifyWithStateCtx) clear() {

}

func (self *verifyWithStateCtx) verifyDescO() (e error) {
	return
}

func (self *verifyWithStateCtx) verifyDescZ() (e error) {
	return
}

func (self *verifyWithStateCtx) verifyFee() (e error) {
	feeCC := self.tx.ToFeeCC_Szk()
	self.ck = assets.NewCKState(true, &self.tx.Fee)
	self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, feeCC[:]...)
	return
}

func (self *verifyWithStateCtx) verifyFrom() (e error) {
	return
}

func (self *verifyWithStateCtx) verifyPkg() (e error) {
	if self.tx.Desc_Pkg.Create != nil {
		if pg := self.state.Pkgs.GetPkgById(&self.tx.Desc_Pkg.Create.Id); pg != nil {
			e = verify_utils.ReportError(fmt.Sprintf("pkg id already exists %v", hexutil.Encode(self.tx.Desc_Pkg.Create.Id[:])), self.tx)
			return
		} else {
			self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, self.tx.Desc_Pkg.Create.Pkg.AssetCM[:]...)
		}
	}

	if self.tx.Desc_Pkg.Transfer != nil {
		if pg := self.state.Pkgs.GetPkgById(&self.tx.Desc_Pkg.Transfer.Id); pg == nil || pg.Closed {
			e = verify_utils.ReportError(fmt.Sprintf("Can not find pkg of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Transfer.Id[:])), self.tx)
			return
		} else {
			if c_superzk.VerifyPKr_X(&self.balance_desc.Hash, &self.tx.Desc_Pkg.Transfer.Sign, &pg.Pack.PKr) {
			} else {
				e = verify_utils.ReportError(fmt.Sprintf("Can not verify pkg sign of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Transfer.Id[:])), self.tx)
				return
			}
		}
	}

	if self.tx.Desc_Pkg.Close != nil {
		if pg := self.state.Pkgs.GetPkgById(&self.tx.Desc_Pkg.Close.Id); pg == nil || pg.Closed {
			e = verify_utils.ReportError(fmt.Sprintf("Can not find pkg of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Close.Id[:])), self.tx)
			return
		} else {
			if c_superzk.VerifyPKr_X(&self.balance_desc.Hash, &self.tx.Desc_Pkg.Close.Sign, &pg.Pack.PKr) {
				self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, pg.Pack.Pkg.AssetCM[:]...)
			} else {
				e = verify_utils.ReportError(fmt.Sprintf("Can not verify pkg sign of the id %v", hexutil.Encode(self.tx.Desc_Pkg.Close.Id[:])), self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyCmd() (e error) {
	if cc := self.tx.Desc_Cmd.ToAssetCC_Szk(); cc != nil {
		self.ck.AddOut(self.tx.Desc_Cmd.OutAsset())
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, cc[:]...)
	}
	return
}

func (self *verifyWithStateCtx) verifyInsP0() (e error) {
	for _, in := range self.tx.Tx1.Ins_P0 {
		if ok := self.state.State.HasIn(&in.Nil); ok {
			e = verify_utils.ReportError("txs.verify p0_in already in nils", self.tx)
			return
		}
		if ok := self.state.State.HasIn(&in.Root); ok {
			e = verify_utils.ReportError("txs.verify p0_in already in nil-roots", self.tx)
			return
		}
		if src := self.state.State.GetOut(&in.Root); src != nil {
			if e = c_superzk.VerifyNil_P0(&self.balance_desc.Hash, &in.Sign, src.ToPKr(), src.RootCM, &in.Nil); e != nil {
				return
			}
			var asset_desc c_type.Asset
			if in.Key != nil {
				if src.Out_Z != nil {
					flag := c_superzk.IsFlagSet(src.Out_Z.RPK[:])
					if asset, rsk, memo, err := c_superzk.Czero_decEInfo(in.Key, flag, &src.Out_Z.EInfo); err != nil {
						e = verify_utils.ReportError("txs.verify p0_in dec info error", self.tx)
						return
					} else {
						if out_cm, err := c_superzk.Czero_genOutCM(&asset, &memo, &src.Out_Z.PKr, &rsk); err != nil {
							e = verify_utils.ReportError("txs.verify p0_in gen out cm error", self.tx)
							return
						} else {
							if src_out_cm := src.TryGetOutCM(); src_out_cm != nil {
								if out_cm != *src.OutCM {
									e = verify_utils.ReportError("txs.verify p0_in confirm error", self.tx)
									return
								} else {
									asset_desc = asset
								}

							} else {
								e = verify_utils.ReportError("txs.verify p0_in the src out_cm is empty", self.tx)
								return
							}
						}
					}
				} else {
					e = verify_utils.ReportError("txs.verify p0_in has key but not point to Out_Z", self.tx)
					return
				}
			} else {
				if src.Out_O != nil {
					asset_desc = src.Out_O.Asset.ToTypeAsset()
				} else {
					e = verify_utils.ReportError("txs.verify p0_in no key but not point to Out_O", self.tx)
					return
				}
			}
			if cc, err := c_superzk.GenAssetCC(&asset_desc); err != nil {
				e = verify_utils.ReportError("txs.verify p0_in no key gen cc error", self.tx)
				return
			} else {
				asset := assets.NewAssetByType(&asset_desc)
				if asset.IsValid() {
					self.ck.AddIn(&asset)
					self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, cc[:]...)
				} else {
					e = verify_utils.ReportError("txs.verify p0_in but asset is invalid", self.tx)
					return
				}
			}
		} else {
			e = verify_utils.ReportError("txs.Verify: p0_in not find in the outs!", self.tx)
			return
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyInsP() (e error) {
	for _, in := range self.tx.Tx1.Ins_P {
		if ok := self.state.State.HasIn(&in.Nil); ok {
			e = verify_utils.ReportError("txs.verify p0_in already in nils", self.tx)
			return
		}
		if ok := self.state.State.HasIn(&in.Root); ok {
			e = verify_utils.ReportError("txs.verify p0_in already in nil-roots", self.tx)
			return
		}
		if src := self.state.State.GetOut(&in.Root); src != nil {
			if e = c_superzk.VerifyNil(&self.balance_desc.Hash, &in.NSign, &in.Nil, src.RootCM, src.ToPKr()); e != nil {
				e = verify_utils.ReportError("txs.verify p0_in verify nil error", self.tx)
				return
			}
			if !c_superzk.VerifyPKr_P(&self.balance_desc.Hash, &in.ASign, src.ToPKr()) {
				e = verify_utils.ReportError("txs.verify p0_in verify pkr error", self.tx)
				return
			}
			var asset_desc c_type.Asset
			if in.Key == nil {
				if src.Out_P != nil {
					asset_desc = src.Out_P.Asset.ToTypeAsset()
				} else {
					e = verify_utils.ReportError("txs.verify p_in has no key but not point to Out_P", self.tx)
					return
				}
			} else {
				if src.Out_C != nil {
					if asset, _, ar, err := c_superzk.DecEInfo(in.Key, &src.Out_C.EInfo); err != nil {
						e = err
						return
					} else {
						if cm, err := c_superzk.GenAssetCM_PC(&asset, &ar); err != nil {
							e = err
							return
						} else {
							if cm != src.Out_C.AssetCM {
								e = verify_utils.ReportError("txs.verify p_in can not confirm to Out_C", self.tx)
								return
							} else {
								asset_desc = asset
							}
						}
					}
				} else {
					e = verify_utils.ReportError("txs.verify p_in has key but not point to Out_C", self.tx)
					return
				}
			}
			if cc, err := c_superzk.GenAssetCC(&asset_desc); err != nil {
				e = err
				return
			} else {
				asset := assets.NewAssetByType(&asset_desc)
				if !asset.IsValid() {
					e = verify_utils.ReportError("txs.verify p_in asset is invalid", self.tx)
					return
				}
				self.ck.AddIn(&asset)
				self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, cc[:]...)
			}
		} else {
			e = verify_utils.ReportError("txs.Verify: in_o not find in the outs!", self.tx)
			return
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyInsC() (e error) {
	for _, in := range self.tx.Tx1.Ins_C {
		self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, in.AssetCM[:]...)
		if ok := self.state.State.HasIn(&in.Nil); ok {
			e = verify_utils.ReportError("txs.verify in already in nils", self.tx)
			return
		} else {
			if out := self.state.State.GetOut(&in.Anchor); out == nil {
				e = verify_utils.ReportError("txs.verify can not find out for anchor", self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithStateCtx) verifyOutP() (e error) {
	for _, out := range self.tx.Tx1.Outs_P {
		self.ck.AddOut(&out.Asset)
		cc := out.ToAssetCC_Szk()
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, cc[:]...)
	}
	return
}

func (self *verifyWithStateCtx) verifyOutC() (e error) {
	for _, out := range self.tx.Tx1.Outs_C {
		self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, out.AssetCM[:]...)
	}
	return
}

func (self *verifyWithStateCtx) verifyBalance() (e error) {
	if len(self.balance_desc.Zout_acms) > 0 || len(self.balance_desc.Zin_acms) > 0 {
		self.balance_desc.Bcr = self.tx.Bcr
		self.balance_desc.Bsign = self.tx.Bsign
		if err := c_superzk.VerifyBalance(&self.balance_desc); err != nil {
			e = err
			return
		}
	} else {
		if e = self.ck.Check(); e != nil {
			return
		}
	}
	return
}

func (self *verifyWithStateCtx) verify() (e error) {
	self.prepare()
	defer self.clear()
	if e = self.verifyDescO(); e != nil {
		return
	}
	if e = self.verifyDescZ(); e != nil {
		return
	}
	if e = self.verifyFee(); e != nil {
		return
	}
	if e = self.verifyFrom(); e != nil {
		return
	}
	if e = self.verifyPkg(); e != nil {
		return
	}
	if e = self.verifyCmd(); e != nil {
		return
	}
	if self.tx.Tx1.Count() > 0 {
		if e = self.verifyInsP0(); e != nil {
			return
		}
		if e = self.verifyInsP(); e != nil {
			return
		}
		if e = self.verifyInsC(); e != nil {
			return
		}
		if e = self.verifyOutP(); e != nil {
			return
		}
		if e = self.verifyOutC(); e != nil {
			return
		}
	}
	if e = self.verifyBalance(); e != nil {
		return
	}
	return
}
