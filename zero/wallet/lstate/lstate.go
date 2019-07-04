package lstate

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/wallet/lstate/balance"
	"github.com/sero-cash/go-sero/zero/wallet/lstate/lstate_types"
)

var current_lstate LState

type LState struct {
	b *balance.Balance
}

func (self *LState) ZState() *zstate.ZState {
	return txtool.Ref_inst.GetState()
}

func (self *LState) GetOut(root *keys.Uint256) (src *lstate_types.OutState, e error) {
	if self.b != nil {
		return self.b.GetOut(root)
	} else {
		return nil, errors.New("lstate.b is nil")
	}
}

func (self *LState) GetPkgs(tk *keys.Uint512, is_from bool) (ret []*lstate_types.Pkg) {
	if self.b != nil {
		return self.b.GetPkgs(tk, is_from)
	} else {
		return
	}
}

func (self *LState) GetOuts(tk *keys.Uint512) (outs []*lstate_types.OutState, e error) {
	if self.b != nil {
		return self.b.GetOuts(tk)
	} else {
		return nil, errors.New("lstate.b is nil")
	}
}

func (self *LState) AddAccount(tk *keys.Uint512) (ret bool) {
	if self.b != nil {
		return self.b.AddAccount(tk)
	} else {
		return
	}
}

func (self *LState) GetAccount(tk *keys.Uint512) (tkn map[keys.Uint256]*utils.U256, tkt map[keys.Uint256][]keys.Uint256) {
	if self.b != nil {
		a := self.b.GetAccount(tk)
		tkn = a.Token
		tkt = a.Ticket
		return
	} else {
		return
	}
}

func CurrentLState() *LState {
	return &current_lstate
}

func InitLState() {
	current_lstate.b = balance.NewBalance()
	return
}
