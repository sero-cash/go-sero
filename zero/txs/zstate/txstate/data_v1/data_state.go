package data_v1

import (
	"errors"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"
)

func (self *Data) RecordState(putter serodb.Putter, root *c_type.Uint256) {
	if out, ok := self.Root2Out[*root]; ok {
		localdb.PutRoot(putter, root, &out)
	} else {
		panic(errors.New("data_v1.recordstate can not find root"))
	}
	return
}

func (self *Data) LoadState(tr tri.Tri) {
	return
}

func (self *Data) SaveState(tr tri.Tri) {
	self.Nils.Save(tr)
	self.Roots.Save(tr)
	self.PKr2Count = make(map[c_type.PKr]int)
	return
}

func (self *Data) HasIn(tr tri.Tri, hash *c_type.Uint256) (exists bool) {
	return self.Nils.Has(tr, hash)
}

func (self *Data) HashRoot(tr tri.Tri, root *c_type.Uint256) bool {
	return self.Roots.Has(tr, root)
}

func (self *Data) GetOut(tr tri.Tri, root *c_type.Uint256) (src *localdb.OutState) {
	if self.Roots.Has(tr, root) {
		var rt *localdb.RootState
		if r, ok := self.Root2Out[*root]; !ok {
			rt = localdb.GetRoot(tr.GlobalGetter(), root)
			self.Root2Out[*root] = *rt
		} else {
			rt = &r
		}
		if rt != nil {
			src = &rt.OS
		}
	}
	if src == nil {
		d := data.NewData(self.Num)
		d.Clear()
		src = d.GetOut(tr, root)
	}
	return
}
