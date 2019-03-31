package utils

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type HSet struct {
	Name   string
	M      map[keys.Uint256]bool
	Orders []keys.Uint256
}

func NewHSet(name string) (ret HSet) {
	ret.Name = name
	return
}

func (self *HSet) Clear() {
	self.M = make(map[keys.Uint256]bool)
	self.Orders = []keys.Uint256{}
}

func (self *HSet) Append(item *keys.Uint256) {
	self.M[*item] = true
	self.Orders = append(self.Orders, *item)
}

func (self *HSet) List() (ret []keys.Uint256) {
	return self.Orders
}

func (self *HSet) K2Name(k *keys.Uint256) (ret []byte) {
	ret = []byte(self.Name)
	ret = append(ret, k[:]...)
	return
}

func (self *HSet) Save(tr tri.Tri) {
	for _, k := range self.Orders {
		if err := tr.TryUpdate(self.K2Name(&k), []byte{1}); err == nil {
			return
		} else {
			panic(err)
			return
		}
	}
}

func (self *HSet) Has(tr tri.Tri, k *keys.Uint256) (ret bool) {
	if _, ok := self.M[*k]; ok {
		ret = true
		return
	} else {
		if bs, err := tr.TryGet(self.K2Name(k)); err == nil {
			if len(bs) > 0 && bs[0] == 1 {
				ret = true
				self.M[*k] = true
				return
			} else {
				return
			}
		} else {
			panic(err)
			return
		}
	}
}
