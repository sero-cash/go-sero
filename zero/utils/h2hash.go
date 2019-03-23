package utils

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type H2Hash struct {
	name string
	m    map[keys.Uint256]keys.Uint256
}

func NewH2Hash(name string) (ret H2Hash) {
	return H2Hash{}
}

func (self *H2Hash) Clear() {
	self.m = make(map[keys.Uint256]keys.Uint256)
}

func (self *H2Hash) Add(id *keys.Uint256, hash *keys.Uint256) {
	self.m[*id] = *hash
}

func (self *H2Hash) Del(id *keys.Uint256) {
	self.m[*id] = keys.Empty_Uint256
}

func (self *H2Hash) K2Name(k *keys.Uint256) (ret []byte) {
	ret = []byte(self.name)
	ret = append(ret, k[:]...)
	return
}

func (self *H2Hash) Save(tr tri.Tri, id *keys.Uint256) {
	v := self.m[*id]
	tr.TryUpdate(self.K2Name(id), v[:])
}

func (self *H2Hash) Get(tr tri.Tri, id *keys.Uint256) (ret *keys.Uint256) {
	var hash keys.Uint256
	var ok bool
	if hash, ok = self.m[*id]; !ok {
		if bs, err := tr.TryGet(self.K2Name(id)); err == nil {
			copy(hash[:], bs[:])
		} else {
			panic(err)
		}
	}
	if hash != keys.Empty_Uint256 {
		ret = &hash
		return
	} else {
		return
	}
}
