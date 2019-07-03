package consensus

import "github.com/sero-cash/go-czero-import/keys"

type Bytes []byte

func (self *Bytes) CopyTo() CItem {
	ret := append(Bytes{}, (*self)[:]...)
	return &ret
}

func (self *Bytes) CopyFrom(from CItem) {
	*self = append(Bytes{}, *from.(*Bytes)...)
	return
}

func (self *Bytes) State() (ret *keys.Uint256) {
	return nil
}
