package light

import (
	"sync/atomic"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
)

type BlockChain interface {
	GetCurrenHeader() *types.Header
	GetHeader(hash *common.Hash) *types.Header
	NewState(hash *common.Hash) *zstate.ZState
	GetTks() []keys.Uint512
	CashChose() *atomic.Value
	GetBlockByNumber(num uint64) *types.Block
	GetDB() serodb.Database
}

type Light struct {
	Bc          BlockChain
	StableState *zstate.ZState
	LogState    *zstate.ZState
}

var Light_inst Light

func (self *Light) SetBC(bc BlockChain) {
	self.Bc = bc
}

func (self *Light) GetDelayedNum(delay uint64) (ret uint64) {
	ret = self.Bc.GetCurrenHeader().Number.Uint64() - delay
	return
}

func GetDelayNumber(current uint64, chose uint64, delay uint64) (num uint64) {
	current_delayed := current
	if current < delay {
		current_delayed = current
	} else if current > delay {
		if (current - delay) < delay {
			current_delayed = delay
		} else {
			current_delayed = current - delay
		}
	}
	if chose > current_delayed {
		return chose
	} else {
		return current_delayed
	}
}
