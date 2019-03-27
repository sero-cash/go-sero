package data_v1

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Data struct {
	Num     uint64
	NilSet  utils.HSet
	RootSet utils.HSet

	Cur      data.Current
	Root2Out map[keys.Uint256]*localdb.OutState

	Dels  utils.Dirtys
	Nils  utils.Dirtys
	Roots utils.Dirtys
}

func NewData(num uint64) (ret *Data) {
	ret = &Data{}
	ret.Num = num
	ret.NilSet = utils.NewHSet("ZState0_InName")
	ret.RootSet = utils.NewHSet("$ZState0$ROOT-OUT$")
	return
}

func (state *Data) Clear() {
	state.Cur = data.NewCur()
	state.Root2Out = make(map[keys.Uint256]*localdb.OutState)
	state.Dels.Clear()
	state.Nils.Clear()
	state.Roots.Clear()
}

func (self *Data) AddOut(root *keys.Uint256, out *localdb.OutState) {
	self.Roots.Append(root)
	self.Root2Out[*root] = out
	self.Cur.Index++
	if self.Cur.Index != int64(out.Index) {
		panic("add out but cur.index != current_index")
	}
	if self.Cur.Index < 0 {
		panic("add out but cur.index < 0")
	}
	return
}

func (self *Data) AddNil(in *keys.Uint256) {
	self.Nils.Append(in)
	self.Dels.Append(in)
}

func (self *Data) AddDel(in *keys.Uint256) {
	self.Dels.Append(in)
}

func (self *Data) GetRoots() (roots []keys.Uint256) {
	return self.Roots.List()
}

func (self *Data) GetDels() (dels []keys.Uint256) {
	return self.Dels.List()
}

func (self *Data) GetIndex() (index int64) {
	return self.Cur.Index
}
