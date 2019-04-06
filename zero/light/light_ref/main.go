package light_ref

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

type Ref struct {
	Bc BlockChain
}

var Ref_inst Ref

func (self *Ref) SetBC(bc BlockChain) {
	self.Bc = bc
}

func (self *Ref) GetDelayedNum(delay uint64) (ret uint64) {
	ret = GetDelayNumber(
		self.Bc.GetCurrenHeader().Number.Uint64(),
		self.Bc.CashChose().Load().(uint64),
		delay,
	)
	return
}

func (self *Ref) GetState() (ret *zstate.ZState) {
	hash := self.Bc.GetCurrenHeader().Hash()
	return self.Bc.NewState(&hash)
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
