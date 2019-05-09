package state2

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/lstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/zconfig"
	"github.com/syndtr/goleveldb/leveldb"
)

type State2 struct {
	db *leveldb.DB
}

func NewState2(bc lstate.BlockChain) (ret State2) {
	if db, err := leveldb.OpenFile(zconfig.State2_dir(), nil); err != nil {
		panic(err)
	} else {
		ret.db = db
	}
	return
}

func (self *State2) ZState() (ret *zstate.ZState) {
	return
}

func (self *State2) GetOut(root *keys.Uint256) (src *lstate.OutState, e error) {
	return
}

func (self *State2) GetPkgs(tk *keys.Uint512, is_from bool) (ret []*lstate.Pkg) {
	return
}
func (self *State2) GetOuts(tk *keys.Uint512) (outs []*lstate.OutState, e error) {
	return
}
