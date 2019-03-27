package utils

import (
	"github.com/sero-cash/go-czero-import/keys"
)

type Dirtys struct {
	Orders []keys.Uint256
}

func (self *Dirtys) Clear() {
	self.Orders = []keys.Uint256{}
}
func (self *Dirtys) Append(item *keys.Uint256) {
	self.Orders = append(self.Orders, *item)
}

func (self *Dirtys) List() (ret []keys.Uint256) {
	return self.Orders
}

/*
func (self *Dirtys) SortedList() (ret []keys.Uint256) {
	list := Uint256s{}
	for _, k := range self.Orders {
		list = append(list, k)
	}
	sort.Sort(list)
	ret = []keys.Uint256(list)
	return
}
*/
