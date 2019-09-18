package assets

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Asset struct {
	Tkn *Token  `rlp:"nil"`
	Tkt *Ticket `rlp:"nil"`
}

func (self *Asset) ToSzkAssetDesc() (ret c_superzk.AssetDesc) {
	asset := self.ToFlatAsset()
	ret = c_superzk.AssetDesc{
		Asset: c_type.Asset{
			Tkn_currency: asset.Tkn.Currency,
			Tkn_value:    asset.Tkn.Value.ToUint256(),
			Tkt_category: asset.Tkt.Category,
			Tkt_value:    asset.Tkt.Value,
		},
	}
	return
}

func NewAssetBySzkDecInfo(info *c_superzk.DecInfoDesc) (ret Asset) {
	ret = NewAsset(
		&Token{
			info.Asset_ret.Tkn_currency,
			utils.NewU256_ByKey(&info.Asset_ret.Tkn_value),
		},
		&Ticket{
			info.Asset_ret.Tkt_category,
			info.Asset_ret.Tkt_value,
		},
	)
	return
}

func (self *Asset) HasAsset() bool {
	if self != nil {
		if self.Tkn != nil {
			if self.Tkn.Value.Cmp(&utils.U256_0) != 0 {
				return true
			}
		}
		if self.Tkt != nil {
			if self.Tkt.Value != c_type.Empty_Uint256 {
				return true
			}
		}
	}
	return false
}

func NewAsset(tkn *Token, tkt *Ticket) (ret Asset) {
	if tkn != nil {
		if tkn.Value.Cmp(&utils.U256_0) > 0 {
			ret.Tkn = tkn.Clone().ToRef()
		}
	}
	if tkt != nil {
		if tkt.Value != c_type.Empty_Uint256 {
			ret.Tkt = tkt.Clone().ToRef()
		}
	}
	return
}

func (self Asset) ToRef() (ret *Asset) {
	return &self
}

func (self *Asset) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	if self.Tkn != nil {
		d.Write(self.Tkn.ToHash().NewRef()[:])
	}
	if self.Tkt != nil {
		d.Write(self.Tkt.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Asset) Clone() (ret Asset) {
	utils.DeepCopy(&ret, self)
	return
}
