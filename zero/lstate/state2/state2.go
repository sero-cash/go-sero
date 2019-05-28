package state2

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/light/light_ref"
	"github.com/sero-cash/go-sero/zero/lstate"
	"github.com/sero-cash/go-sero/zero/lstate/state2/accounts"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
)

type State2 struct {
	db *accounts.DB
}

func NewState2() (ret State2) {
	return
}

func (self *State2) ZState() (ret *zstate.ZState) {
	return light_ref.Ref_inst.GetState()
}

func (self *State2) GetOut(root *keys.Uint256) (src *lstate.OutState, e error) {
	s, err := self.db.GetOut(root)
	return &s, err
}

func (self *State2) GetPkgs(tk *keys.Uint512, is_from bool) (ret []*lstate.Pkg) {
	return
}
func (self *State2) GetOuts(tk *keys.Uint512) (outs []*lstate.OutState, e error) {
	return self.db.GetOuts(tk)
}

func (self *State2) AddAccount(tk *keys.Uint512) (ret bool) {
	top_num := light_ref.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()
	return self.db.AddAccount(tk, top_num)
}

func (self *State2) GetAccount(tk *keys.Uint512) (ret accounts.Account) {
	return self.db.GetAccount(tk)
}
