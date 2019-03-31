package data_v1

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Data struct {
	Num uint64

	Cur      data.Current
	Root2Out map[keys.Uint256]localdb.RootState

	Dels  utils.Dirtys
	Nils  utils.HSet
	Roots utils.HSet
}

func NewData(num uint64) (ret *Data) {
	ret = &Data{}
	ret.Num = num
	ret.Nils = utils.NewHSet(data.ZSTATE0_INNAME)
	ret.Roots = utils.NewHSet("$ZState0$ROOT-OUT$")
	return
}

func (state *Data) Clear() {
	state.Cur = data.NewCur()
	state.Root2Out = make(map[keys.Uint256]localdb.RootState)
	state.Dels.Clear()
	state.Nils.Clear()
	state.Roots.Clear()
}

func (self *Data) AddOut(root *keys.Uint256, out *localdb.OutState, txhash *keys.Uint256) {
	self.Roots.Append(root)
	rs := localdb.RootState{}
	rs.Num = self.Num
	rs.OS = *out
	if txhash != nil {
		rs.TxHash = *txhash
	}
	self.Root2Out[*root] = rs
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
