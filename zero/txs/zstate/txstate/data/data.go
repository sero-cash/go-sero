package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type Data struct {
	Num    uint64
	Cur    Current
	Block  StateBlock
	G2outs map[keys.Uint256]*localdb.OutState
	H2tx   map[keys.Uint256]*keys.Uint256

	Dirty_G2ins  map[keys.Uint256]bool
	Dirty_G2outs map[keys.Uint256]bool
}

func NewData(num uint64) (ret *Data) {
	ret = &Data{}
	ret.Num = num
	return
}

func (state *Data) clear_dirty() {
	state.Dirty_G2ins = make(map[keys.Uint256]bool)
	state.Dirty_G2outs = make(map[keys.Uint256]bool)
}

func (state *Data) Clear() {
	state.Cur = NewCur()
	state.G2outs = make(map[keys.Uint256]*localdb.OutState)
	state.H2tx = make(map[keys.Uint256]*keys.Uint256)
	state.Block = StateBlock{}
	state.clear_dirty()
}

func (self *Data) appendDel(del *keys.Uint256) {
	if del == nil {
		panic("set_last_out but del is nil")
	}
	self.Block.Dels = append(self.Block.Dels, *del)
}

func (self *Data) appendRoot(root *keys.Uint256) {
	if root == nil {
		panic("set_last_out but root is nil")
	}
	self.Cur.Index = self.Cur.Index + int64(1)
	self.Block.Roots = append(self.Block.Roots, *root)
}

func (self *Data) addInByNilOrRoot(in *keys.Uint256) {
	self.Dirty_G2ins[*in] = true
}

func (self *Data) addOutByRoot(k *keys.Uint256, out *localdb.OutState) {
	self.G2outs[*k] = out
	self.Dirty_G2outs[*k] = true
}

func (self *Data) AddTxOut(pkr *keys.PKr) int {
	return 0
}

func (self *Data) AddOut(root *keys.Uint256, out *localdb.OutState, txhash *keys.Uint256) {
	self.addOutByRoot(root, out)
	self.appendRoot(root)
	if txhash != nil {
		th := *txhash
		self.H2tx[*root] = &th
	} else {
		self.H2tx[*root] = nil
	}
	if self.Cur.Index != int64(out.Index) {
		panic("add out but cur.index != current_index")
	}
	if self.Cur.Index < 0 {
		panic("add out but cur.index < 0")
	}
	return
}

func (self *Data) AddNil(in *keys.Uint256) {
	self.addInByNilOrRoot(in)
	self.appendDel(in)
}

func (self *Data) AddDel(in *keys.Uint256) {
	self.appendDel(in)
}

func (self *Data) GetRoots() (roots []keys.Uint256) {
	return self.Block.Roots
}

func (self *Data) GetDels() (dels []keys.Uint256) {
	return self.Block.Dels
}

func (self *Data) GetIndex() (index int64) {
	return self.Cur.Index
}
