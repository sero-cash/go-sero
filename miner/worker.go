// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package miner

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/accounts"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/zero/stake"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/consensus"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/core/vm"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/params"
	"github.com/sero-cash/go-sero/serodb"
)

const (
	resultQueueSize  = 10
	miningLogAtDepth = 5

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
	// chainSideChanSize is the size of channel listening to ChainSideEvent.
	chainSideChanSize = 10

	chainVoteSize = 30

	//delayBlock = 1
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *Result)
	Stop()
	Start()
	GetHashRate() int64
}

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig

	state   *state.StateDB // apply state changes here
	tcount  int            // tx count in cycle
	gasPool *core.GasPool  // available gas used to pack transactions

	Block *types.Block // the new block

	header   *types.Header
	txs      []*types.Transaction
	receipts []*types.Receipt

	createdAt time.Time

	handledTxs    []*types.Transaction
	errHandledTxs []*types.Transaction

	gasReward uint64
}

func (self *Work) Copy() (ret *Work) {
	ret = &Work{}
	*ret = *self
	ret.state = self.state.Copy()
	return
}

type Result struct {
	Work  *Work
	Block *types.Block
}

// worker is the main object which takes care of applying messages to the new state
type worker struct {
	config *params.ChainConfig
	engine consensus.Engine

	mu sync.Mutex

	// update loop
	mux          *event.TypeMux
	txsCh        chan core.NewTxsEvent
	txsSub       event.Subscription
	voteCh       chan core.NewVoteEvent
	voteSub      event.Subscription
	chainHeadCh  chan core.ChainHeadEvent
	chainHeadSub event.Subscription
	chainSideCh  chan core.ChainSideEvent
	chainSideSub event.Subscription
	wg           sync.WaitGroup

	voter voter

	agents    map[Agent]struct{}
	recv      chan *Result
	powRecv   chan *Result
	posTaskCh chan *Result

	eth     Backend
	chain   *core.BlockChain
	proc    core.Validator
	chainDb serodb.Database

	coinbase accounts.Account
	extra    []byte

	currentMu sync.Mutex
	current   *Work

	snapshotMu    sync.RWMutex
	snapshotBlock *types.Block
	snapshotState *state.StateDB

	unconfirmed *unconfirmedBlocks // set of locally mined blocks pending canonicalness confirmations

	// atomic status counters
	mining int32
	atWork int32

	pendingVote pendingVote

	stopVote  chan struct{}
	stopPow   chan struct{}
	stopWrite chan struct{}
	loopwg    sync.WaitGroup
	stopMu    sync.Mutex
	//pendingVoteMu sync.RWMutex
	//pendingVote   map[voteKey]voteSet
	//pendingVoteTime sync.Map
	//pendingPosMu  sync.RWMutex
	//pendingPos    map[common.Hash]time.Time
	//pendingVoteMu sync.RWMutex
	//pendingVote   map[common.Hash]mapset.Set
}

func newWorker(config *params.ChainConfig, engine consensus.Engine, account accounts.Account, voter voter, sero Backend, mux *event.TypeMux) *worker {
	worker := &worker{
		config:      config,
		engine:      engine,
		eth:         sero,
		mux:         mux,
		txsCh:       make(chan core.NewTxsEvent, txChanSize),
		voteCh:      make(chan core.NewVoteEvent, chainVoteSize),
		chainHeadCh: make(chan core.ChainHeadEvent, chainHeadChanSize),
		chainSideCh: make(chan core.ChainSideEvent, chainSideChanSize),
		chainDb:     sero.ChainDb(),
		recv:        make(chan *Result, resultQueueSize),
		powRecv:     make(chan *Result),
		posTaskCh:   make(chan *Result),
		chain:       sero.BlockChain(),
		proc:        sero.BlockChain().Validator(),
		coinbase:    account,
		agents:      make(map[Agent]struct{}),
		unconfirmed: newUnconfirmedBlocks(sero.BlockChain(), miningLogAtDepth),
		voter:       voter,
		pendingVote: newPendingVote(),
		stopVote:    make(chan struct{}),
		stopPow:     make(chan struct{}),
		stopWrite:   make(chan struct{}),
	}
	// Subscribe NewTxsEvent for tx pool
	worker.txsSub = sero.TxPool().SubscribeNewTxsEvent(worker.txsCh)
	worker.voteSub = worker.voter.SubscribeWorkerVoteEvent(worker.voteCh)
	// Subscribe events for blockchain
	worker.chainHeadSub = sero.BlockChain().SubscribeChainHeadEvent(worker.chainHeadCh)
	worker.chainSideSub = sero.BlockChain().SubscribeChainSideEvent(worker.chainSideCh)
	go worker.update()
	go worker.voteLoop()
	go worker.powResultLoop()
	go worker.resultLoop()
	worker.commitNewWork()

	return worker
}

func (self *worker) setSerobase(account accounts.Account) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.coinbase = account
}

func (self *worker) setExtra(extra []byte) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.extra = extra
}

func (self *worker) pending() (*types.Block, *state.StateDB) {
	if atomic.LoadInt32(&self.mining) == 0 {
		// return a snapshot to avoid contention on currentMu mutex
		self.snapshotMu.RLock()
		defer self.snapshotMu.RUnlock()
		return self.snapshotBlock, self.snapshotState.Copy()
	}

	self.currentMu.Lock()
	defer self.currentMu.Unlock()
	return self.current.Block, self.current.state.Copy()
}

func (self *worker) pendingBlock() *types.Block {
	if atomic.LoadInt32(&self.mining) == 0 {
		// return a snapshot to avoid contention on currentMu mutex
		self.snapshotMu.RLock()
		defer self.snapshotMu.RUnlock()
		return self.snapshotBlock
	}

	self.currentMu.Lock()
	defer self.currentMu.Unlock()
	return self.current.Block
}

func (self *worker) start() {
	self.mu.Lock()
	defer self.mu.Unlock()

	atomic.StoreInt32(&self.mining, 1)

	// spin up agents
	for agent := range self.agents {
		agent.Start()
	}
}

func (self *worker) stop() {
	self.wg.Wait()

	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.mining) == 1 {
		for agent := range self.agents {
			agent.Stop()
		}
	}
	atomic.StoreInt32(&self.mining, 0)
	atomic.StoreInt32(&self.atWork, 0)
}

func (self *worker) register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.powRecv)
}

func (self *worker) unregister(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.agents, agent)
	agent.Stop()
}

func (self *worker) update() {

	defer self.txsSub.Unsubscribe()
	defer self.chainHeadSub.Unsubscribe()
	defer self.chainSideSub.Unsubscribe()

	for {
		// A real event arrived, process interesting content
		select {
		// Handle ChainHeadEvent
		case block := <-self.chainHeadCh:
			header := block.Block.Header()

			self.pendingVote.deleteBefore(header.Number.Uint64() - 1)

			self.commitNewWork()

			// Handle ChainSideEvent
		case <-self.chainSideCh:

			//Handle NewTxsEvent
		case ev := <-self.txsCh:
			//Apply transactions to the pending state if we're not mining.

			//Note all transactions received may not be continuous with transactions
			//already included in the current mining block. These transactions will
			//be automatically eliminated.
			if atomic.LoadInt32(&self.mining) == 0 && self.current != nil {
				self.currentMu.Lock()
				txset := types.NewTransactionsByPrice(ev.Txs)
				addr := common.Address{}
				pkr := self.coinbase.GetPkr(nil)
				addr.SetBytes(pkr[:])

				self.current.commitTransactions(self.mux, txset, self.chain, addr)
				self.updateSnapshot()
				self.currentMu.Unlock()
			}
			// System stopped
		case <-self.txsSub.Err():
			return
		case <-self.chainHeadSub.Err():
			return
		case <-self.chainSideSub.Err():
			return
		}
	}
}

func (self *worker) powResultLoop() {
	self.loopwg.Add(1)
	defer self.loopwg.Done()
	for {
		select {
		case result := <-self.powRecv:
			if result == nil {
				continue
			}
			if result.Block.Header().Number.Uint64() < seroparam.SIP4() {
				self.recv <- result
			} else {
				lotter := newLotter(self, result.Block, result.Work.state)
				self.voter.AddLottery(&types.Lottery{result.Block.ParentHash(), result.Block.Number().Uint64() - 1, result.Block.HashPos()})

				log.Info("Broadcast Lottery", "poshash", result.Block.HashPos(), "block", result.Block.Number().Uint64())

				go func() {
					if lotter.wait() {
						result.Block.SetVotes(lotter.currentHeaderVotes, lotter.parentHeaderVotes)
						self.recv <- result
					}
				}()
			}
		case <-self.stopPow:
			log.Info("stop worker pow")
			return
		}
	}
}

func (self *worker) voteLoop() {
	self.loopwg.Add(1)
	defer self.loopwg.Done()
	defer self.voteSub.Unsubscribe()
	for {
		select {
		case voteResult := <-self.voteCh:
			if atomic.LoadInt32(&self.mining) == 0 {
				continue
			}

			vote := voteResult.Vote
			if vote == nil {
				continue
			}

			if vote.ParentNum+1 < self.pendingBlock().NumberU64()-1 {
				continue
			}

			log.Trace("worker voteLoop", "posHash", vote.PosHash, "block", vote.ParentNum+1, "share", vote.ShareId, "idx", vote.Idx)

			self.pendingVote.add(vote)

		case <-self.voteSub.Err():
			return
		case <-self.stopVote:
			log.Info("stop worker vote loop...")
			return
		}

	}
}

func (self *worker) resultLoop() {
	self.loopwg.Add(1)
	defer self.loopwg.Done()
	for {
		select {
		case result := <-self.recv:

			self.stopMu.Lock()
			atomic.AddInt32(&self.atWork, -1)

			if result == nil {
				self.stopMu.Unlock()
				continue
			}
			block := result.Block
			work := result.Work

			// Update the block hash in all logs since it is now available and not when the
			// receipt/log of individual transactions were created.
			for _, r := range work.receipts {
				for _, l := range r.Logs {
					l.BlockHash = block.Hash()
				}
			}
			for _, log := range work.state.Logs() {
				log.BlockHash = block.Hash()
			}
			self.currentMu.Lock()
			stat, err := self.chain.WriteBlockWithState(block, work.receipts, work.state)
			self.currentMu.Unlock()
			if err != nil {
				log.Error("Failed writing block to chain", "err", err)
				self.stopMu.Unlock()
				continue
			}
			// Broadcast the block and announce chain insertion event
			self.mux.Post(core.NewMinedBlockEvent{Block: block})
			var (
				events []interface{}
				logs   = work.state.Logs()
			)
			events = append(events, core.ChainEvent{Block: block, Hash: block.Hash(), Logs: logs})
			self.eth.TxPool().RemoveTxs(work.handledTxs)
			if stat == core.CanonStatTy {
				events = append(events, core.ChainHeadEvent{Block: block})
			}
			self.chain.PostChainEvents(events, logs)

			// Insert the block into the set of pending ones to resultLoop for confirmations
			self.unconfirmed.Insert(block.NumberU64(), block.Hash())
			log.Info(fmt.Sprintf("mined new block done in %v, number = %v, txs = %v", time.Since(work.createdAt), block.NumberU64(), len(block.Body().Transactions)))
			self.stopMu.Unlock()
		case <-self.stopWrite:
			self.stopMu.Lock()
			self.stopMu.Unlock()
			log.Info("stop worker result....")
			return
		}
	}
}

// push sends a new work task to currently live miner agents.
func (self *worker) push(work *Work) {
	if atomic.LoadInt32(&self.mining) != 1 {
		return
	}
	self.eth.TxPool().RemoveTxs(work.errHandledTxs)
	for agent := range self.agents {
		atomic.AddInt32(&self.atWork, 1)
		if ch := agent.Work(); ch != nil {
			ch <- work
		}
	}
}

// makeCurrent creates a new environment for the current cycle.
func (self *worker) makeCurrent(parent *types.Block, header *types.Header) error {
	state, err := self.chain.StateAt(parent.Header())
	if err != nil {
		return err
	}
	work := &Work{
		config:    self.config,
		state:     state,
		header:    header,
		createdAt: time.Now(),
	}
	// Keep track of transactions which return errors so they can be removed
	work.tcount = 0
	self.current = work
	return nil
}

func (self *worker) commitNewWork() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	tstart := time.Now()
	parent := self.chain.CurrentBlock()

	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info("Mining too far in the future", "resultLoop", common.PrettyDuration(wait))
		time.Sleep(wait)
	}

	num := parent.Number()
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		GasLimit:   core.CalcGasLimit(parent),
		Extra:      self.extra,
		Time:       big.NewInt(tstamp),
	}
	// Only set the coinbase if we are mining (avoid spurious block rewards)
	if atomic.LoadInt32(&self.mining) == 1 {
		//pkr :=  superzk.Pk2PKr(self.coinbase.ToUint512(), nil)
		pkr, licr, ret := superzk.Pk2PKrAndLICr(self.coinbase.Address.ToUint512().NewRef(), header.Number.Uint64())
		if !ret {
			log.Error("Failed to Addr2PKrAndLICr")
			return
		}
		header.Licr = licr
		copy(header.Coinbase[:], pkr[:])
	}

	if err := self.engine.Prepare(self.chain, header); err != nil {
		log.Error("Failed to prepare header for mining", "err", err)
		return
	}

	// Could potentially happen if starting to mine in an odd state.
	err := self.makeCurrent(parent, header)
	if err != nil {
		log.Error("Failed to create mining context", "err", err)
		return
	}
	// Create the current work task and check any fork transitions needed
	work := self.current

	pending, err := self.eth.TxPool().Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return
	}
	txs := types.NewTransactionsByPrice(pending)

	if header.Number.Uint64() >= seroparam.SIP4() {
		stakeState := stake.NewStakeState(work.state)
		err := stakeState.ProcessBeforeApply(self.chain, header)
		if err != nil {
			log.Error("ProcessBeforeApply", "err", err)
			return
		}
	}

	work.commitTransactions(self.mux, txs, self.chain, header.Coinbase)

	log.Debug(fmt.Sprintf("commitTransactions %v tx done in %v", len(pending), time.Since(tstart)))

	// Create the new block to seal with the consensus engine
	if work.Block, err = self.engine.Finalize(self.chain, header, work.state, work.txs, work.receipts, work.gasReward); err != nil {
		log.Error("Failed to finalize block for sealing", "err", err)
		return
	}
	// We only care about logging if we're actually mining.
	if atomic.LoadInt32(&self.mining) == 1 {
		log.Info("Commit new mining work", "number", work.Block.Number(), "txs", work.tcount, "elapsed", common.PrettyDuration(time.Since(tstart)))
		self.unconfirmed.Shift(work.Block.NumberU64() - 1)
	}
	self.push(work)
	self.updateSnapshot()
}

func (self *worker) updateSnapshot() {
	self.snapshotMu.Lock()
	defer self.snapshotMu.Unlock()

	self.snapshotBlock = types.NewBlock(
		self.current.header,
		self.current.txs,
		self.current.receipts,
	)
	self.snapshotState = self.current.state.Copy()
}
func (self *worker) Close() {
	close(self.stopPow)
	close(self.stopVote)
	close(self.stopWrite)
	self.loopwg.Wait()

}

func (env *Work) commitTransactions(mux *event.TypeMux, txs *types.TransactionsByPrice, bc *core.BlockChain, coinbase common.Address) {
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(env.header.GasLimit)
	}

	var coalescedLogs []*types.Log

	for {
		// If we don't have enough gas for any further transactions then we're done
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			break
		}
		// Retrieve the next transaction and abort if all done
		tx := txs.Peek()
		if tx == nil {
			break
		}

		if true && (!seroparam.Is_Dev()) {
			if env.header.Number.Uint64() == seroparam.SIP10() {
				txs.Shift()
				env.errHandledTxs = append(env.errHandledTxs, tx)
				continue
			}
		}

		// Start executing the transaction
		env.state.Prepare(tx.Hash(), common.Hash{}, env.tcount)

		err, logs := env.commitTransaction(tx, bc, coinbase, env.gasPool)
		switch err {
		case core.ErrGasLimitReached:
			log.Info("Gas limit exceeded for current block", "block", bc.CurrentBlock().Header().Number.Uint64())
			// Pop the current out-of-gas transaction without shifting in the next from the account
			log.Trace("Gas limit exceeded for current block", "sender", tx.From())
			txs.Pop()
			break

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			coalescedLogs = append(coalescedLogs, logs...)
			env.tcount++
			txs.Shift()
			env.handledTxs = append(env.handledTxs, tx)
		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", tx.Hash(), "err", err)
			txs.Shift()
			env.errHandledTxs = append(env.errHandledTxs, tx)
		}
	}

	if len(coalescedLogs) > 0 || env.tcount > 0 {
		// make a copy, the state caches the logs and these logs get "upgraded" from pending to mined
		// logs by filling in the block hash when the block was mined by the local miner. This can
		// cause a race condition if a log was "upgraded" before the PendingLogsEvent is processed.
		cpy := make([]*types.Log, len(coalescedLogs))
		for i, l := range coalescedLogs {
			cpy[i] = new(types.Log)
			*cpy[i] = *l
		}
		go func(logs []*types.Log, tcount int) {
			if len(logs) > 0 {
				mux.Post(core.PendingLogsEvent{Logs: logs})
			}
			if tcount > 0 {
				mux.Post(core.PendingStateEvent{})
			}
		}(cpy, env.tcount)
	}
}

func (env *Work) commitTransaction(tx *types.Transaction, bc *core.BlockChain, coinbase common.Address, gp *core.GasPool) (error, []*types.Log) {
	snap := env.state.Snapshot()

	receipt, gas, err := core.ApplyTransaction(env.config, bc, &coinbase, gp, env.state, env.header, tx, &env.header.GasUsed, vm.Config{})
	if err != nil {
		env.state.RevertToSnapshot(snap)
		return err, nil
	}

	env.gasReward += new(big.Int).Mul(new(big.Int).SetUint64(gas), tx.GasPrice()).Uint64()
	env.txs = append(env.txs, tx)
	env.receipts = append(env.receipts, receipt)

	return err, receipt.Logs
}
