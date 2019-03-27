package utils

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type HSet struct {
	Name string
}

func NewHSet(name string) (ret HSet) {
	ret.Name = name
	return
}

func (self *HSet) Clear() {}

func (self *HSet) K2Name(k *keys.Uint256) (ret []byte) {
	ret = []byte(self.Name)
	ret = append(ret, k[:]...)
	return
}

func (self *HSet) Save(tr tri.Tri, k *keys.Uint256) {
	if err := tr.TryUpdate(self.K2Name(k), []byte{1}); err == nil {
		return
	} else {
		panic(err)
		return
	}
}

func (self *HSet) Has(tr tri.Tri, k *keys.Uint256) (ret bool) {
	if bs, err := tr.TryGet(self.K2Name(k)); err == nil {
		if len(bs) > 0 && bs[0] == 1 {
			ret = true
			return
		} else {
			return
		}
	} else {
		panic(err)
		return
	}
}
