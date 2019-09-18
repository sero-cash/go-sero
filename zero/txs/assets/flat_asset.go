package assets

import "github.com/sero-cash/go-czero-import/c_type"

type FlatAssert struct {
	Tkn Token
	Tkt Ticket
}

func (self *Asset) ToFlatAsset() (ret FlatAssert) {
	if self.Tkt != nil {
		ret.Tkt = *self.Tkt
	}
	if self.Tkn != nil {
		ret.Tkn = *self.Tkn
	}
	return
}

func (self *Asset) ToTypeAsset() (ret c_type.Asset) {
	asset := self.ToFlatAsset()
	ret = c_type.Asset{
		Tkn_currency: asset.Tkn.Currency,
		Tkn_value:    asset.Tkn.Value.ToUint256(),
		Tkt_category: asset.Tkt.Category,
		Tkt_value:    asset.Tkt.Value,
	}
	return
}
