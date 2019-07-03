package consensus

import (
	"errors"

	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/serodb"
)

type FakeTri struct {
	m map[string][]byte
}

func NewFakeTri() (ret FakeTri) {
	ret.m = make(map[string][]byte)
	return
}

func (self *FakeTri) Get(key []byte) ([]byte, error) {
	return self.TryGet(key)
}

func (self *FakeTri) Has(key []byte) (bool, error) {
	if _, ok := self.m[string(key)]; !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (self *FakeTri) TryGet(key []byte) ([]byte, error) {
	if v, ok := self.m[string(key)]; !ok {
		return nil, errors.New("can not find value")
	} else {
		return v, nil
	}
}

func (self *FakeTri) Put(key, value []byte) error {
	return self.TryUpdate(key, value)
}

func (self *FakeTri) TryUpdate(key, value []byte) error {
	self.m[string(key)] = value
	return nil
}

func (self *FakeTri) Delete(key []byte) error {
	return self.TryDelete(key)
}

func (self *FakeTri) TryDelete(key []byte) error {
	if _, ok := self.m[string(key)]; ok {
		delete(self.m, string(key))
		return nil
	} else {
		return errors.New("delete a empty item")
	}
}

type FakeDB struct {
	tri FakeTri
	db  FakeTri
}

func NewFakeDB() (ret FakeDB) {
	ret.tri = NewFakeTri()
	ret.db = NewFakeTri()
	return
}

func (self *FakeDB) Num() uint64 {
	return 0
}

func (self *FakeDB) CurrentTri() state.Tri {
	return &self.tri
}

func (self *FakeDB) GlobalGetter() serodb.Getter {
	return &self.db
}
