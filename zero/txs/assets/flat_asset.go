package assets

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
