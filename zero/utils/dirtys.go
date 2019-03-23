package utils

import (
	"sort"

	"github.com/sero-cash/go-czero-import/keys"
)

type Dirtys struct {
	orders []keys.Uint256
	dirtys map[keys.Uint256]bool
}

func (self *Dirtys) Clear() {
	self.orders = []keys.Uint256{}
	self.dirtys = make(map[keys.Uint256]bool)
}
func (self *Dirtys) Append(item *keys.Uint256) {
	self.orders = append(self.orders, *item)
	self.dirtys[*item] = true
}

func (self *Dirtys) List() (ret []keys.Uint256) {
	return self.orders
}

func (self *Dirtys) SortedList() (ret []keys.Uint256) {
	list := Uint256s{}
	for _, k := range self.orders {
		list = append(list, k)
	}
	sort.Sort(list)
	ret = []keys.Uint256(list)
	return
}
