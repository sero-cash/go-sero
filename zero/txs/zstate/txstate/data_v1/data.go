package data_v1

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/txstate/data"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Data struct {
	Num    uint64
	Cur    data.Current
	G2outs map[keys.Uint256]*localdb.OutState

	DirtyIns  utils.Dirtys
	DirtyOuts utils.Dirtys
}

func NewData(num uint64) (ret Data) {
	ret.Num = num
	return
}

func (state *Data) Clear() {
	state.Cur = data.NewCur()
	state.G2outs = make(map[keys.Uint256]*localdb.OutState)
	state.DirtyIns.Clear()
	state.DirtyOuts.Clear()
}
