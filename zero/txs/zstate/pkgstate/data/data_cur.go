package data

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

func pkgBlockName(num uint64) (ret []byte) {
	ret = []byte(fmt.Sprintf("PKGSTATE_BLOCK_NAME_%d", num))
	return
}

func (self *Data) SaveCur(tr tri.Tri) {
	return
}

func (self *Data) LoadCur(tr tri.Tri) {
	return
}
