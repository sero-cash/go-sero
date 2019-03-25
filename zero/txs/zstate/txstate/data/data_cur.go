package data

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

const LAST_OUTSTATE0_NAME = tri.KEY_NAME("ZState0_Cur")
const BLOCK_NAME = "ZState0_BLOCK"

func Name2BKey(name string, num uint64) (ret []byte) {
	key := fmt.Sprintf("%s_%d", name, num)
	ret = []byte(key)
	return
}

func (self *Data) SaveCur(tr tri.Tri) {
	if self.Dirty_last_out {
		tri.UpdateObj(tr, LAST_OUTSTATE0_NAME.Bytes(), &self.Cur)
		tri.UpdateObj(
			tr,
			Name2BKey(BLOCK_NAME, self.Num),
			&self.Block,
		)
	}
	return
}

func (self *Data) LoadCur(tr tri.Tri) {
	get := CurrentGet{}
	tri.GetObj(
		tr,
		LAST_OUTSTATE0_NAME.Bytes(),
		&get,
	)
	self.Cur = get.Out

	blockget := State0BlockGet{}
	tri.GetObj(
		tr,
		Name2BKey(BLOCK_NAME, self.Num),
		&blockget,
	)
	self.Block = blockget.Out
	return
}

func (self *Data) AppendDel(del *keys.Uint256) {
	if del == nil {
		panic("set_last_out but del is nil")
	}
	self.Block.Dels = append(self.Block.Dels, *del)
	self.Dirty_last_out = true
}

func (self *Data) appendRoot(root *keys.Uint256) {
	if root == nil {
		panic("set_last_out but root is nil")
	}
	self.Cur.Index = self.Cur.Index + int64(1)
	self.Block.Roots = append(self.Block.Roots, *root)
	self.Dirty_last_out = true
}
