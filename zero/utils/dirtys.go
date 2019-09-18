package utils

import "github.com/sero-cash/go-czero-import/c_type"

type Dirtys struct {
	Orders []c_type.Uint256
}

func (self *Dirtys) Clear() {
	self.Orders = []c_type.Uint256{}
}
func (self *Dirtys) Append(item *c_type.Uint256) {
	self.Orders = append(self.Orders, *item)
}

func (self *Dirtys) List() (ret []c_type.Uint256) {
	return self.Orders
}

/*
func (self *Dirtys) SortedList() (ret []c_type.Uint256) {
	list := Uint256s{}
	for _, k := range self.Orders {
		list = append(list, k)
	}
	sort.Sort(list)
	ret = []c_type.Uint256(list)
	return
}
*/
