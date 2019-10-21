package data_v1

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"
	"github.com/sero-cash/go-sero/zero/utils"
)

const ZSTATE0_ROOT_OUT = "$ZState0$ROOT-OUT$"

type Data struct {
	Num uint64

	Root2Out map[c_type.Uint256]localdb.RootState

	Dels  utils.Dirtys
	Nils  utils.HSet
	Roots utils.HSet

	PKr2Count map[c_type.PKr]int
}

func NewData(num uint64) (ret *Data) {
	ret = &Data{}
	ret.Num = num
	ret.Nils = utils.NewHSet(data.ZSTATE0_INNAME)
	ret.Roots = utils.NewHSet(ZSTATE0_ROOT_OUT)
	return
}

func (state *Data) Clear() {
	state.Root2Out = make(map[c_type.Uint256]localdb.RootState)
	state.PKr2Count = make(map[c_type.PKr]int)
	state.Dels.Clear()
	state.Nils.Clear()
	state.Roots.Clear()
}

func (self *Data) AddTxOut(pkr *c_type.PKr) int {
	if count, ok := self.PKr2Count[*pkr]; !ok {
		self.PKr2Count[*pkr] = 1
		return 1
	} else {
		count++
		self.PKr2Count[*pkr] = count
		return count
	}
}

func (self *Data) AddOut(root *c_type.Uint256, out *localdb.OutState, txhash *c_type.Uint256) {
	self.Roots.Append(root)
	rs := localdb.RootState{}
	rs.Num = self.Num
	rs.OS = *out
	if txhash != nil {
		rs.TxHash = *txhash
	}
	self.Root2Out[*root] = rs
	return
}

func (self *Data) AddNil(in *c_type.Uint256) {
	self.Nils.Append(in)
	self.Dels.Append(in)
}

func (self *Data) AddDel(in *c_type.Uint256) {
	self.Dels.Append(in)
}

func (self *Data) GetRoots() (roots []c_type.Uint256) {
	return self.Roots.List()
}

func (self *Data) GetDels() (dels []c_type.Uint256) {
	return self.Dels.List()
}
