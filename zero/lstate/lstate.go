package lstate

import (
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
)

type BlockChain interface {
	GetCurrenHeader() *types.Header
	GetHeader(hash *common.Hash) *types.Header
	NewState(hash *common.Hash) *zstate.ZState
	GetTks() []keys.Uint512
	CashChose() *atomic.Value
	GetDB() serodb.Database
}

type LState interface {
	Parse(last_chose uint64) (chose uint64)

	ZState() *zstate.ZState
	GetOut(root *keys.Uint256) (src *OutState, e error)
	GetPkgs(tk *keys.Uint512, is_from bool) (ret []*Pkg)
	GetOuts(tk *keys.Uint512) (outs []*OutState, e error)
	AddAccount(tk *keys.Uint512) (ret bool)
}

var current_lstate LState

func SetCurrentLState(lst LState) {
	current_lstate = lst
}

func CurrentLState() LState {
	return current_lstate
}

var current_bc BlockChain

func BC() BlockChain {
	return current_bc
}

func Run(bc BlockChain, lst LState) {
	current_bc = bc
	current_lstate = lst
	go run()
}

func Parse(chose uint64) uint64 {
	defer func() {
		if r := recover(); r != nil {
			log.Error("parse block chain error : ", "number", BC().GetCurrenHeader().Number, "recover", r)
			debug.PrintStack()
		}
	}()
	return current_lstate.Parse(chose)
}

func run() {
	chose := uint64(0)
	for {
		last_chose := chose

		chose = Parse(last_chose)

		cashChose := BC().CashChose()
		cashChose.Store(chose)

		if chose-last_chose <= 1 {
			time.Sleep(1000 * 1000 * 1000 * 8)
		} else {
			time.Sleep(1000 * 1000 * 10)
		}
	}
}
