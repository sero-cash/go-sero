package data_v1

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"
)

func (self *Data) LoadState(tr tri.Tri) {
	get := data.CurrentGet{}
	tri.GetObj(
		tr,
		data.LAST_OUTSTATE0_NAME.Bytes(),
		&get,
	)
	self.Cur = get.Out
	return
}

func (self *Data) SaveState(tr tri.Tri) {
	tri.UpdateObj(tr, data.LAST_OUTSTATE0_NAME.Bytes(), &self.Cur)
	for _, k := range self.Nils.List() {
		self.NilSet.Save(tr, &k)
	}

	for _, k := range self.Roots.List() {
		self.RootSet.Save(tr, &k)
	}
	return
}

func (self *Data) HasIn(tr tri.Tri, hash *keys.Uint256) (exists bool) {
	return self.NilSet.Has(tr, hash)
}

func (self *Data) GetOut(tr tri.Tri, root *keys.Uint256) (src *localdb.OutState) {
	if self.RootSet.Has(tr, root) {
		src = localdb.GetOut(tr.GlobalGetter(), root)
	}
	if src == nil {
		d := data.NewData(self.Num)
		d.Clear()
		src = d.GetOut(tr, root)
	}
	return
}
