package utils

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type H2Hash struct {
	Name string
	M    map[keys.Uint256]keys.Uint256
}

func NewH2Hash(name string) (ret H2Hash) {
	ret.Name = name
	return
}

func (self *H2Hash) Clear() {
	self.M = make(map[keys.Uint256]keys.Uint256)
}

func (self *H2Hash) Add(id *keys.Uint256, hash *keys.Uint256) {
	self.M[*id] = *hash
}

func (self *H2Hash) Del(id *keys.Uint256) (ret keys.Uint256) {
	ret = self.M[*id]
	delete(self.M, *id)
	return
}

func (self *H2Hash) Get(id *keys.Uint256) (ret keys.Uint256) {
	ret = self.M[*id]
	return
}

func (self *H2Hash) K2Name(k *keys.Uint256) (ret []byte) {
	ret = []byte(self.Name)
	ret = append(ret, k[:]...)
	return
}

func (self *H2Hash) SaveByDB(putter serodb.Putter, id *keys.Uint256) {
	v := self.M[*id]
	if err := putter.Put(self.K2Name(id), v[:]); err == nil {
		return
	} else {
		panic(err)
		return
	}
}

func (self *H2Hash) GetByDB(getter serodb.Getter, id *keys.Uint256) (ret keys.Uint256) {
	var ok bool
	if ret, ok = self.M[*id]; !ok {
		if bs, err := getter.Get(self.K2Name(id)); err == nil {
			copy(ret[:], bs[:])
			return
		} else {
			panic(err)
			return
		}
	} else {
		return
	}
}

func (self *H2Hash) SaveByTri(tr tri.Tri, id *keys.Uint256) {
	v := self.M[*id]
	if err := tr.TryUpdate(self.K2Name(id), v[:]); err == nil {
		return
	} else {
		panic(err)
		return
	}
}

func (self *H2Hash) GetByTri(tr tri.Tri, id *keys.Uint256) (ret keys.Uint256) {
	var ok bool
	if ret, ok = self.M[*id]; !ok {
		if bs, err := tr.TryGet(self.K2Name(id)); err == nil {
			copy(ret[:], bs[:])
			return
		} else {
			panic(err)
			return
		}
	} else {
		return
	}
}
