package txtool

import (
	"math/big"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
)

type BlockChain interface {
	IsValid() bool
	GetCurrenHeader() *types.Header
	GetHeader(hash *common.Hash) *types.Header
	CurrentState(hash *common.Hash) *zstate.ZState
	IsContract(address common.Address) (ret bool, e error)
	GetSeroGasLimit(to *common.Address, tfee *assets.Token, gasPrice *big.Int) (gaslimit uint64, e error)
	GetTks() []c_type.Tk
	GetTkAt(tk *c_type.Tk) uint64
	GetBlockByNumber(num uint64) *types.Block
	GetHeaderByNumber(num uint64) *types.Header
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
		delay,
	)
	return
}

func (self *Ref) CurrentState() (ret *zstate.ZState) {
	num := self.GetDelayedNum(seroparam.DefaultConfirmedBlock())
	block := self.Bc.GetBlockByNumber(num)
	hash := block.Hash()
	return self.Bc.CurrentState(&hash)
}

func GetDelayNumber(current uint64, delay uint64) (num uint64) {
	if current < delay {
		return 0
	} else {
		return current - delay
	}
}
