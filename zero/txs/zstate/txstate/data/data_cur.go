package data

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

const LAST_OUTSTATE0_NAME = tri.KEY_NAME("ZState0_Cur")
const BLOCK_NAME = "ZState0_BLOCK"

func (self *Data) Name2BKey(name string, num uint64) (ret []byte) {
	key := fmt.Sprintf("%s_%d", name, num)
	ret = []byte(key)
	return
}

func (self *Data) SaveCur(tr tri.Tri) {
	return
}

func (self *Data) LoadCur(tr tri.Tri) {
	return
}
