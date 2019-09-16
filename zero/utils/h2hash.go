package utils

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type H2Hash struct {
	Name string
	M    map[c_type.Uint256]c_type.Uint256
}

func NewH2Hash(name string) (ret H2Hash) {
	ret.Name = name
	return
}

func (self *H2Hash) Clear() {
	self.M = make(map[c_type.Uint256]c_type.Uint256)
}

func (self *H2Hash) Add(id *c_type.Uint256, hash *c_type.Uint256) {
	self.M[*id] = *hash
}

func (self *H2Hash) Del(id *c_type.Uint256) (ret c_type.Uint256) {
	ret = self.M[*id]
	delete(self.M, *id)
	return
}

func (self *H2Hash) Get(id *c_type.Uint256) (ret c_type.Uint256) {
	ret = self.M[*id]
	return
}

func (self *H2Hash) K2Name(k *c_type.Uint256) (ret []byte) {
	ret = []byte(self.Name)
	ret = append(ret, k[:]...)
	return
}

func (self *H2Hash) SaveByDB(putter serodb.Putter, id *c_type.Uint256) {
	v := self.M[*id]
	if err := putter.Put(self.K2Name(id), v[:]); err == nil {
		return
	} else {
		panic(err)
		return
	}
}

func (self *H2Hash) GetByDB(getter serodb.Getter, id *c_type.Uint256) (ret c_type.Uint256) {
	var ok bool
	if ret, ok = self.M[*id]; !ok {
		if bs, err := getter.Get(self.K2Name(id)); err == nil {
			copy(ret[:], bs[:])
			self.M[*id] = ret
			return
		} else {
			panic(err)
			return
		}
	} else {
		return
	}
}

func (self *H2Hash) SaveByTri(tr tri.Tri, id *c_type.Uint256) {
	v := self.M[*id]
	if err := tr.TryUpdate(self.K2Name(id), v[:]); err == nil {
		return
	} else {
		panic(err)
		return
	}
}

func (self *H2Hash) GetByTri(tr tri.Tri, id *c_type.Uint256) (ret c_type.Uint256) {
	var ok bool
	if ret, ok = self.M[*id]; !ok {
		if bs, err := tr.TryGet(self.K2Name(id)); err == nil {
			copy(ret[:], bs[:])
			self.M[*id] = ret
			return
		} else {
			panic(err)
			return
		}
	} else {
		return
	}
}
