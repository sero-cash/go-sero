// Copyright 2014 The go-ethereum Authors
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

package core

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/stake"
	"github.com/sero-cash/go-sero/zero/txs/stx"

	"github.com/sero-cash/go-sero/zero/txtool/verify"

	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/metrics"
	"github.com/sero-cash/go-sero/params"
)

const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
)

var (
	ErrVerifyError = errors.New("stx Verify error")

	// ErrUnderpriced is returned if a transaction's gas priced is below the minimum
	// configured for the transaction pool.
	ErrUnderpriced = errors.New("transaction underpriced")

	// ErrIntrinsicGas is returned if the transaction is specified to use less gas
	// than required to start the invocation.
	ErrIntrinsicGas = errors.New("intrinsic gas too low")

	// ErrGasLimit is returned if a transaction's requested gas limit exceeds the
	// maximum allowance of the current block.
	ErrGasLimit = errors.New("exceeds block gas limit")

	// ErrOversizedData is returned if the input data of a transaction is greater
	// than some meaningful limit a user might use. This is not a consensus error
	// making the transaction invalid, rather a DOS protection.
	ErrOversizedData = errors.New("oversized data")

	ErrCurrencyError = errors.New("currency error")
)

var (
	evictionInterval    = 5 * time.Minute // Time interval to check for evictable transactions
	statsReportInterval = 8 * time.Second // Time interval to report transaction pool stats
)

var (
	// General tx metrics
	invalidTxCounter     = metrics.NewRegisteredCounter("txpool/invalid", nil)
	underpricedTxCounter = metrics.NewRegisteredCounter("txpool/underpriced", nil)
)

// TxStatus is the current status of a transaction as seen by the pool.
type TxStatus uint

const (
	TxStatusUnknown TxStatus = iota
	TxStatusQueued
	TxStatusPending
)

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and event subscribers.
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(header *types.Header) (*state.StateDB, error)

	SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
}

// TxPoolConfig are the configuration parameters of the transaction pool.
type TxPoolConfig struct {
	NoLocals bool // Whether local transaction handling should be disabled

	PriceLimit uint64 // Minimum gas priced to enforce for acceptance into the pool

	AccountSlots uint64 // Number of executable transaction slots guaranteed per account
	GlobalSlots  uint64 // Maximum number of executable transaction slots for all accounts
	AccountQueue uint64 // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue  uint64 // Maximum number of non-executable transaction slots for all accounts

	Lifetime time.Duration // Maximum amount of time non-executable transaction are queued

	StartLight bool
}

// DefaultTxPoolConfig contains the default configurations for the transaction
// pool.
var DefaultTxPoolConfig = TxPoolConfig{

	PriceLimit:   params.Gta,
	AccountSlots: 16,
	GlobalSlots:  2048,
	AccountQueue: 64,
	GlobalQueue:  32,

	Lifetime: 3 * time.Hour,
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *TxPoolConfig) sanitize() TxPoolConfig {
	conf := *config

	if conf.PriceLimit < 1 {
		log.Warn("Sanitizing invalid txpool priced limit", "provided", conf.PriceLimit, "updated", DefaultTxPoolConfig.PriceLimit)
		conf.PriceLimit = DefaultTxPoolConfig.PriceLimit
	}
	return conf
}

func TxToOut(tx types.Transaction) (result []txtool.Out, txHash c_type.Uint256) {
	txCommonHash := tx.Hash()
	copy(txHash[:], txCommonHash[:])
	ztxs := tx.GetZZSTX()
	if ztxs != nil {
		tx1 := tx.GetZZSTX().Tx1
		for index := range tx1.Outs_P {
			out := txtool.Out{}
			rootState := localdb.RootState{}
			os := localdb.OutState{}
			out_p := &tx1.Outs_P[index]
			if c_superzk.IsSzkPKr(&tx1.Outs_P[index].PKr) {
				os.Out_P = out_p
				os.GenRootCM()
			} else {
				os.Out_O = &stx_v0.Out_O{
					Addr:  out_p.PKr,
					Asset: out_p.Asset,
					Memo:  out_p.Memo,
				}
				os.GenRootCM()
			}
			rootState.OS = os
			rootState.TxHash = txHash
			out.State = rootState
			result = append(result, out)

		}

		for index := range tx1.Outs_C {
			out := txtool.Out{}
			rootState := localdb.RootState{}
			os := localdb.OutState{}
			os.Out_C = &tx1.Outs_C[index]
			os.GenRootCM()
			rootState.OS = os
			rootState.TxHash = txHash
			out.State = rootState
			result = append(result, out)
		}
	}
	return
}

type PKrTxOuts map[c_type.PKr]map[c_type.Uint256]*TxOutInfo

func (p PKrTxOuts) AddPendingTxOut(tx types.Transaction) {

	txOuts, _ := TxToOut(tx)
	nowUnix := uint64(time.Now().Unix())
	for index := range txOuts {
		p.AddOut(tx.From(), tx.Gas(), tx.Gas(), tx.GasPrice(), 0, common.Hash{}, nowUnix, txOuts[index])
	}

}

func (p PKrTxOuts) AddImmatureTxOut(tx types.Transaction, blockNumber uint64, blockHash common.Hash, time uint64) {

	txOuts, _ := TxToOut(tx)
	for index := range txOuts {
		p.AddOut(tx.From(), tx.Gas(), tx.Gas(), tx.GasPrice(), blockNumber, blockHash, time, txOuts[index])
	}

}

func (p PKrTxOuts) AddOut(
	from common.Address,
	gas uint64, gasUsed uint64,
	gasPrice *big.Int,
	blockNumber uint64,
	blockHash common.Hash, time uint64,
	out txtool.Out) {
	pkr := out.State.OS.ToPKr()
	txHash := out.State.TxHash
	if pkr != nil {
		if _, ok := p[*pkr]; ok {
			if _, ok := p[*pkr][txHash]; ok {
				p[*pkr][txHash].addOut(txHash, from, gas, gasUsed, gasPrice, blockNumber, blockHash, time, out)
			} else {
				txOutInfo := &TxOutInfo{OutExists: map[c_type.Uint256]bool{}}
				txOutInfo.addOut(txHash, from, gas, gasUsed, gasPrice, blockNumber, blockHash, time, out)
				p[*pkr][txHash] = txOutInfo
			}

		} else {
			txOutInfo := &TxOutInfo{OutExists: map[c_type.Uint256]bool{}}
			txOutInfo.addOut(txHash, from, gas, gasUsed, gasPrice, blockNumber, blockHash, time, out)
			txOutMap := make(map[c_type.Uint256]*TxOutInfo)
			txOutMap[txHash] = txOutInfo
			p[*pkr] = txOutMap
		}
	}

}

func (p PKrTxOuts) delPendintTxOut(tx types.Transaction) {
	txOuts, txHash := TxToOut(tx)
	for index := range txOuts {
		out := txOuts[index]
		pkr := out.State.OS.ToPKr()
		if pkr != nil {
			delete(p[*pkr], txHash)
			if len(p[*pkr]) == 0 {
				delete(p, *pkr)
			}

		}
	}
}

func (p PKrTxOuts) delPkrTxOut(pkr c_type.PKr, txHash c_type.Uint256) {

	delete(p[pkr], txHash)
	if len(p[pkr]) == 0 {
		delete(p, pkr)
	}
}

type TxOutInfo struct {
	TxHash      c_type.Uint256
	BlockNumber uint64
	BlockHash   common.Hash
	GasUsed     uint64
	Gas         uint64
	GasPrice    *big.Int
	From        common.Address
	Time        uint64
	Outs        []txtool.Out
	OutExists   map[c_type.Uint256]bool
}

func (t *TxOutInfo) addOut(txHash c_type.Uint256, from common.Address, gas, gasUsed uint64, gasPrice *big.Int, blockNumber uint64, blockHash common.Hash, time uint64, out txtool.Out) {
	t.TxHash = txHash
	t.BlockNumber = blockNumber
	t.BlockHash = blockHash
	t.Gas = gas
	t.GasUsed = gasUsed
	t.GasPrice = gasPrice
	t.From = from
	t.Time = time

	var existsKey c_type.Uint256

	if out.State.OS.Out_O != nil {
		existsKey = out.State.OS.Out_O.ToHash()
	}

	if out.State.OS.Out_P != nil {
		existsKey = out.State.OS.Out_P.ToHash()
	}

	if out.State.OS.Out_C != nil {
		existsKey = out.State.OS.Out_C.Tx1_Hash()
	}

	if _, ok := t.OutExists[existsKey]; !ok {
		t.Outs = append(t.Outs, out)
		t.OutExists[existsKey] = true

	}

}

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
//
// The pool separates processable transactions (which can be applied to the
// current state) and future transactions. Transactions move between those
// two states over time as they are received and processed.
type TxPool struct {
	config       TxPoolConfig
	chainconfig  *params.ChainConfig
	chain        blockChain
	gasPrice     *big.Int
	txFeed       event.Feed
	scope        event.SubscriptionScope
	chainHeadCh  chan ChainHeadEvent
	chainHeadSub event.Subscription
	//abi       types.Signer
	mu sync.RWMutex

	currentState  *state.StateDB      // Current state in the blockchain head
	pendingState  *state.ManagedState // Pending state tracking virtual nonces
	currentMaxGas uint64              // Current gas limit for transaction caps

	//locals *accountSet // Set of local transaction to exempt from eviction rules
	//journal *txJournal  // Journal of local transaction to back up to disk

	all        *txLookup     // All transactions to allow lookups
	priced     *txPricedList // All transactions sorted by priced
	newQueue   *txPricedList
	newPending *txPricedList
	beats      hashTime
	faileds    hashTime

	wg sync.WaitGroup // for shutdown sync

	pkrTxOuts PKrTxOuts

	homestead bool
	mining    int32
}

// NewTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain) *TxPool {
	// Sanitize the input to ensure no vulnerable gas prices are set
	config = (&config).sanitize()

	// Create the transaction pool with its initial settings
	pool := &TxPool{
		config:      config,
		chainconfig: chainconfig,
		chain:       chain,
		beats:       make(map[common.Hash]time.Time),
		faileds:     make(map[common.Hash]time.Time),
		all:         newTxLookup(),
		chainHeadCh: make(chan ChainHeadEvent, chainHeadChanSize),
		gasPrice:    new(big.Int).SetUint64(config.PriceLimit),
	}
	pool.pkrTxOuts = make(map[c_type.PKr]map[c_type.Uint256]*TxOutInfo)
	//pool.locals = newAccountSet()
	pool.priced = newTxPricedList(pool.all)
	pool.newQueue = newTxPricedList(newTxLookup())
	pool.newPending = newTxPricedList(newTxLookup())
	pool.reset(nil, chain.CurrentBlock().Header())

	// Subscribe events from blockchain
	pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadCh)

	// Start the event loop and return
	pool.wg.Add(1)
	go pool.loop()

	return pool
}

func (pool *TxPool) SetMining(m int32) {
	atomic.StoreInt32(&pool.mining, m)
}

func (pool *TxPool) Mining() bool {
	return atomic.LoadInt32(&pool.mining) > 0
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (pool *TxPool) loop() {
	defer pool.wg.Done()

	// Start the stats reporting and transaction eviction tickers
	//var prevPending, prevQueued, prevStales int

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	// Track the previous head headers for transaction reorgs
	head := pool.chain.CurrentBlock()

	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle ChainHeadEvent
		case ev := <-pool.chainHeadCh:
			if ev.Block != nil {
				pool.mu.Lock()

				pool.reset(head.Header(), ev.Block.Header())
				head = ev.Block

				pool.mu.Unlock()
			}
			// Be unsubscribed due to system stopped
		case <-pool.chainHeadSub.Err():
			return

			// Handle stats reporting ticks
		case <-report.C:
			log.Debug("Transaction pool status report", "queued", pool.all.Count())

		case <-evict.C:
			drop := types.Transactions{}
			pool.mu.RLock()
			pendingTxs := pool.newPending.Flatten()
			queuedTxs := pool.newQueue.Flatten()
			beats := pool.beats.Flatten()
			faileds := pool.faileds.Flatten()
			pool.mu.RUnlock()

			dropFaileds := []common.Hash{}
			for k, v := range faileds {
				if time.Since(v) > pool.config.Lifetime {
					dropFaileds = append(dropFaileds, k)
				}
			}

			for _, tx := range pendingTxs {
				if err := pool.validateTx(tx, false); err != nil {
					drop = append(drop, tx)
				} else {
					if t, ok := beats[tx.Hash()]; ok && time.Since(t) > pool.config.Lifetime {
						drop = append(drop, tx)
					}
				}
			}

			for _, tx := range queuedTxs {
				if err := pool.validateTx(tx, false); err != nil {
					drop = append(drop, tx)
				}
			}

			pool.mu.Lock()

			for _, tx := range drop {
				pool.faileds[tx.Hash()] = time.Now()
				pool.removeAllTx(tx.Hash())
				if pool.canAddPkrTx() {
					pool.pkrTxOuts.delPendintTxOut(*tx)
				}
			}
			for _, h := range dropFaileds {
				delete(pool.faileds, h)
			}
			pool.mu.Unlock()

		}
	}
}

func (pool *TxPool) RemoveTxs(txs types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, tx := range txs {
		pool.removeAllTx(tx.Hash())

	}
}

// lockedReset is a wrapper around reset to allow calling it in a thread safe
// manner. This method is only ever used in the tester!
func (pool *TxPool) lockedReset(oldHead, newHead *types.Header) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.reset(oldHead, newHead)
}

func (pool *TxPool) canAddPkrTx() bool {
	difference := time.Now().Unix() - pool.chain.CurrentBlock().Time().Int64()
	if difference > 10*60 {
		return false
	}
	return pool.config.StartLight
}

// reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
func (pool *TxPool) reset(oldHead, newHead *types.Header) {
	// If we're reorging an old state, reinject all dropped transactions
	var reinject types.Transactions
	// Reorg seems shallow enough to pull in all transactions into memory
	var discarded, included types.Transactions

	if oldHead != nil && oldHead.Hash() != newHead.ParentHash {
		// If the reorg is too deep, avoid doing it (will happen during fast sync)
		oldNum := oldHead.Number.Uint64()
		newNum := newHead.Number.Uint64()

		if depth := uint64(math.Abs(float64(oldNum) - float64(newNum))); depth > 64 {
			log.Debug("Skipping deep transaction reorg", "depth", depth)
		} else {

			var (
				rem = pool.chain.GetBlock(oldHead.Hash(), oldHead.Number.Uint64())
				add = pool.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
			)
			for rem.NumberU64() > add.NumberU64() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
			}
			for add.NumberU64() > rem.NumberU64() {
				included = append(included, add.Transactions()...)
				if pool.canAddPkrTx() {
					for _, tx := range add.Transactions() {
						pool.pkrTxOuts.AddImmatureTxOut(*tx, add.Number().Uint64(), add.Hash(), add.Time().Uint64())
					}
				}
				if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			for rem.Hash() != add.Hash() {
				discarded = append(discarded, rem.Transactions()...)
				if rem = pool.chain.GetBlock(rem.ParentHash(), rem.NumberU64()-1); rem == nil {
					log.Error("Unrooted old chain seen by tx pool", "block", oldHead.Number, "hash", oldHead.Hash())
					return
				}
				included = append(included, add.Transactions()...)
				if pool.canAddPkrTx() {
					for _, tx := range add.Transactions() {
						pool.pkrTxOuts.AddImmatureTxOut(*tx, add.Number().Uint64(), add.Hash(), add.Time().Uint64())
					}
				}

				if add = pool.chain.GetBlock(add.ParentHash(), add.NumberU64()-1); add == nil {
					log.Error("Unrooted new chain seen by tx pool", "block", newHead.Number, "hash", newHead.Hash())
					return
				}
			}
			reinject = types.TxDifference(discarded, included)
		}
	}

	// Initialize the internal state to the current head
	if newHead == nil {
		newHead = pool.chain.CurrentBlock().Header() // Special case during testing
	}
	statedb, err := pool.chain.StateAt(newHead)
	if err != nil {
		log.Error("Failed to reset txpool state", "err", err)
		return
	}
	pool.currentState = statedb
	pool.pendingState = state.ManageState(statedb)
	pool.currentMaxGas = newHead.GasLimit

	if len(included) == 0 {
		add := pool.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
		if pool.canAddPkrTx() {
			for _, tx := range add.Transactions() {
				pool.pkrTxOuts.AddImmatureTxOut(*tx, add.Number().Uint64(), add.Hash(), add.Time().Uint64())
			}
		}
		included = append(included, add.Transactions()...)
	}

	for _, tx := range included {
		pool.removeAllTx(tx.Hash())
		log.Debug("confirm removeTx tx", "hash", tx.Hash())
	}
	// Inject any transactions discarded due to reorgs
	log.Debug("Reinjecting stale transactions", "count", len(reinject))
	if len(reinject) > 0 {
		pool.addTxsLocked(reinject, false, false)
		for _, tx := range reinject {
			log.Info("reinject tx", "hash", tx.Hash())
		}
	}

	pool.promoteExecutables(true)
}

// Stop terminates the transaction pool.
func (pool *TxPool) Stop() {
	// Unsubscribe all subscriptions registered from txpool
	pool.scope.Close()

	// Unsubscribe subscriptions registered from blockchain
	pool.chainHeadSub.Unsubscribe()
	pool.wg.Wait()

	log.Info("Transaction pool stopped")
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}

// SetGasPrice updates the minimum priced required by the transaction pool for a
// new transaction, and drops all transactions below this threshold.
func (pool *TxPool) SetGasPrice(price *big.Int) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.gasPrice = price
	pool.priced.RemoveWithPrice(pool.gasPrice)
	pool.newQueue.RemoveWithPrice(pool.gasPrice)
	pool.newPending.RemoveWithPrice(pool.gasPrice)

	log.Info("Transaction pool priced threshold updated", "priced", pool.gasPrice)
}

// State returns the virtual managed state of the transaction pool.
func (pool *TxPool) State() *state.ManagedState {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.pendingState
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) Stats() (int, int, int, int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.newPending.Len(), pool.newQueue.Len(), pool.priced.Len(), pool.all.Count()
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (pool *TxPool) Content() (types.Transactions, types.Transactions, types.Transactions) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pending := types.Transactions{}
	if pool.newPending != nil && pool.newPending.items != nil {
		pending = append(pending, pool.newPending.Flatten()...)
	}

	queued := types.Transactions{}

	if pool.newQueue != nil && pool.newQueue.items != nil {
		queued = append(queued, pool.newQueue.Flatten()...)
	}

	all := types.Transactions{}

	if pool.priced != nil && pool.priced.items != nil {
		all = append(all, pool.priced.Flatten()...)
	}

	return pending, queued, all
}

func (pool *TxPool) PendingOuts(pkr c_type.PKr) map[c_type.Uint256]*TxOutInfo {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.pkrTxOuts[pkr]
}

func (pool *TxPool) DelMaturedOuts(pkr c_type.PKr, txHash c_type.Uint256, currentNum uint64) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if txHashMap, ok := pool.pkrTxOuts[pkr]; ok {
		for k, v := range txHashMap {
			if v.BlockNumber != 0 && v.BlockNumber < currentNum {
				pool.pkrTxOuts.delPkrTxOut(pkr, k)
			}
		}
	}
	pool.pkrTxOuts.delPkrTxOut(pkr, txHash)

}

// Pending retrieves all currently processable transactions, groupped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *TxPool) Pending() (types.Transactions, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.newPending.Flatten(), nil
}

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (priced and size).
func (pool *TxPool) validateTx(tx *types.Transaction, local bool) (e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("validateTx error : ", "hash", tx.Hash().Hex(), "recover", r)
			debug.PrintStack()
			e = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	if tx.Size() > 3200*1024 {
		return ErrOversizedData
	}

	// Ensure the transaction doesn't exceed the current block limit gas.
	var gaslimit uint64
	if gaslimit, e = pool.currentState.GetTxGasLimit(tx); e != nil {
		return
	}
	if pool.currentMaxGas < gaslimit {
		return ErrGasLimit
	}

	num := pool.chain.CurrentBlock().NumberU64()
	if err := verify.VerifyWithoutState(tx.Ehash().NewRef(), tx.GetZZSTX(), num); err != nil {
		log.Trace("validateTx verify without state failed", "hash", tx.Hash().Hex(), "verify stx err", err)
		//return ErrVerifyError
		return err
	}

	copyState := pool.currentState.CopyWithNoZState()
	if err := pool.checkDescCmd(tx.GetZZSTX(), copyState); err != nil {
		return err
	}

	state := copyState.NextZState()
	err := verify.VerifyWithState(tx.GetZZSTX(), state, num)
	//err := verify.Verify(tx.GetZZSTX(), pool.currentState.Copy().GetZState())
	if err != nil {
		log.Trace("validateTx error", "hash", tx.Hash().Hex(), "verify stx err", err)
		//pool.faileds[tx.Hash()] = time.Now()
		//return ErrVerifyError
		return err
	}

	// Drop non-local transactions under our own minimal accepted gas priced
	if !local && pool.gasPrice.Cmp(tx.GasPrice()) > 0 {
		return ErrUnderpriced
	}

	if !tx.IsOpContract() {
		if len(tx.Data()) > 0 {
			return errors.New(`not create or call crontract tx playdata must be nil`)
		}
	}

	intrGas, err := IntrinsicGas(tx.Data(), tx.To() == nil)
	if err != nil {
		return err
	}
	if gaslimit < intrGas {
		return ErrIntrinsicGas
	}
	return nil
}

func (pool *TxPool) checkDescCmd(tx *stx.T, state *state.StateDB) (err error) {
	cmd := tx.Desc_Cmd
	stakeState := stake.NewStakeState(state)
	if cmd.BuyShare != nil {
		if cmd.BuyShare.Pool != nil {
			stakePool := stakeState.GetStakePool(common.BytesToHash(cmd.BuyShare.Pool[:]))
			if stakePool == nil || stakePool.Closed {
				err = errors.New("pool is not exist or closed")
				return
			}
		}
	} else if cmd.RegistPool != nil {
		id := crypto.Keccak256Hash(tx.From[:])
		stakePool := stakeState.GetStakePool(id)
		if stakePool == nil {
			if cmd.RegistPool.Value.ToInt().Cmp(stake.GetPoolValueThreshold()) != 0 {
				err = errors.New("registPool value error")
				return
			}
		} else {
			if stakePool.Closed {
				err = errors.New("registPool but stakePool is closed")
				return
			}
		}
		if !superzk.IsPKrValid(&cmd.RegistPool.Vote) {
			err = errors.New("registPool Vote is invalid")
			return
		}
		if cmd.RegistPool.FeeRate > seroparam.HIGHEST_STAKING_NODE_FEE_RATE {
			err = fmt.Errorf("registPool Vote fee must <= %v%%", seroparam.HIGHEST_STAKING_NODE_FEE_RATE/100)
			return
		}
		if cmd.RegistPool.FeeRate < seroparam.LOWEST_STAKING_NODE_FEE_RATE {
			err = fmt.Errorf("registPool Vote fee must >= %v%%", seroparam.LOWEST_STAKING_NODE_FEE_RATE/100)
			return
		}
	} else if cmd.ClosePool != nil {
		id := crypto.Keccak256Hash(tx.From[:])
		stakePool := stakeState.GetStakePool(id)
		if stakePool == nil {
			err = errors.New("pool is not exist")
			return
		}
		if stakePool.Closed {
			err = errors.New("pool is closed")
			return
		}
		if stakePool.BlockNumber+stake.GetLockingBlockNum() > pool.chain.CurrentBlock().NumberU64() {
			err = errors.New("pool locking in")
			return
		}
	}
	return
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution. If the transaction is a replacement for
// an already pending or queued one, it overwrites the previous and returns this
// so outer code doesn't uselessly call promote.
//
// If a newly added transaction is marked as local, its sending account will be
// whitelisted, preventing any associated transaction from being dropped out of
// the pool due to pricing constraints.
func (pool *TxPool) add(tx *types.Transaction, local bool) (bool, error) {
	// If the transaction is already known, discard it
	hash := tx.Hash()
	if pool.all.Get(hash) != nil && !local {
		log.Trace("Discarding already known transaction", "hash", hash.Hex())
		return false, fmt.Errorf("known transaction: %x", hash)
	}

	if _, ok := pool.faileds[hash]; ok {
		log.Trace("Discarding already known failed transaction", "hash", hash.Hex())
		return false, fmt.Errorf("known failed transaction: %x", hash)
	}

	currentBlockNum := pool.chain.CurrentBlock().NumberU64()

	if true && (!seroparam.Is_Dev()) {
		if (seroparam.SIP10()-25) < currentBlockNum && currentBlockNum < (seroparam.SIP10()+25) {
			return false, fmt.Errorf("protect SIP10:%v", seroparam.SIP10())
		}
	}

	// If the transaction fails basic validation, discard it
	if err := pool.validateTx(tx, local); err != nil {
		log.Info("Discarding invalid transaction", "hash", hash.Hex(), "err", err)
		pool.faileds[tx.Hash()] = time.Now()
		invalidTxCounter.Inc(1)
		return false, err
	}
	// If the transaction pool is full, discard underpriced transactions
	if uint64(pool.all.Count()) >= pool.config.GlobalSlots+pool.config.GlobalQueue {
		// If the new transaction is underpriced, don't accept it
		if !local && pool.newQueue.Underpriced(tx) {
			log.Info("Discarding underpriced transaction", "hash", hash.Hex(), "priced", tx.GasPrice())
			underpricedTxCounter.Inc(1)
			return false, ErrUnderpriced
		}
		// New transaction is better than our worse ones, make room for it
		drop := pool.priced.Discard(pool.all.Count() - int(pool.config.GlobalSlots+pool.config.GlobalQueue-1))
		for _, tx := range drop {
			pool.removeWorkQueue(tx)
			if pool.canAddPkrTx() {
				pool.pkrTxOuts.delPendintTxOut(*tx)
			}
			log.Info("Discarding freshly underpriced transaction", "hash", tx.Hash().Hex(), "priced", tx.GasPrice())
		}
	}

	flag, err := pool.enqueueTx(hash, tx)
	if err != nil {
		return false, err
	}
	if pool.canAddPkrTx() {
		pool.pkrTxOuts.AddPendingTxOut(*tx)
	}
	if flag {
		log.Info("Pooled new future transaction", "hash", hash.Hex())
	} else {
		log.Info("Discard new future transaction", "hash", hash.Hex())

	}
	return flag, nil
}

// Note, this method assumes the pool lock is held!
func (pool *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction) (bool, error) {
	// Try to insert the transaction into the future queue

	if pool.newQueue.Add(tx, pool.gasPrice) {
		if pool.all.Get(hash) == nil {
			flag := pool.priced.Add(tx, pool.gasPrice)
			if !flag {
				log.Info("txPool enqueueTx error", "tx.gasPrice", tx.GasPrice().String())
			}
		}
	} else {
		return false, errors.New("gas price too low")
	}

	return true, nil
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one in the mean time, ensuring it goes around the local
// pricing constraints.
func (pool *TxPool) AddLocal(tx *types.Transaction) error {
	return pool.addTx(tx, !pool.config.NoLocals)
}

// AddRemote enqueues a single transaction into the pool if it is valid. If the
// sender is not among the locally tracked ones, full pricing constraints will
// apply.
func (pool *TxPool) AddRemote(tx *types.Transaction) error {
	return pool.addTx(tx, false)
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones in the mean time, ensuring they go around
// the local pricing constraints.
func (pool *TxPool) AddLocals(txs []*types.Transaction) []error {
	return pool.addTxs(txs, !pool.config.NoLocals, true)
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid.
// If the senders are not among the locally tracked ones, full pricing constraints
// will apply.
func (pool *TxPool) AddRemotes(txs []*types.Transaction) []error {
	for _, tx := range txs {
		log.Debug("AddRemotes tx", "hash", tx.Hash().Hex())
	}
	return pool.addTxs(txs, false, false)
}

// addTx enqueues a single transaction into the pool if it is valid.
func (pool *TxPool) addTx(tx *types.Transaction, local bool) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Try to inject the transaction and update any state
	_, err := pool.add(tx, local)
	if err != nil {
		return err
	}
	broadCast := false
	if local {
		if pool.Mining() || pool.config.StartLight {
			pool.broadCastLocalTx(tx)
		} else {
			broadCast = true
		}
	}
	pool.promoteExecutables(broadCast)

	return nil
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (pool *TxPool) addTxs(txs []*types.Transaction, local bool, broadcast bool) []error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTxsLocked(txs, local, broadcast)
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (pool *TxPool) addTxsLocked(txs []*types.Transaction, local bool, broadcast bool) []error {
	// Add the batch of transaction, tracking the accepted ones
	errs := []error{}
	errCount := 0
	knowErrCount := 0
	failedErrCount := 0
	added := 0
	for _, tx := range txs {
		_, err := pool.add(tx, local)
		if err != nil {
			if strings.Contains(err.Error(), "known transaction") {
				knowErrCount++
			} else if strings.Contains(err.Error(), "known failed transaction") {
				failedErrCount++
			} else {
				errCount++
			}

		} else {
			added++
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	log.Trace("txpool", "addTxs", len(txs), "knowErr", knowErrCount, "failedErr", failedErrCount, "validateErr", errCount)
	if added > 0 {
		pool.promoteExecutables(broadcast)
	}
	return errs
}

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (pool *TxPool) Status(hashes []common.Hash) []TxStatus {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	status := make([]TxStatus, len(hashes))
	for i, hash := range hashes {
		if tx := pool.all.Get(hash); tx != nil {
			if pool.newQueue.Get(hash) != nil {
				status[i] = TxStatusQueued
			} else if pool.newPending.Get(hash) != nil {
				status[i] = TxStatusPending
			} else {
				status[i] = TxStatusUnknown
			}
		}
	}
	return status
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (pool *TxPool) Get(hash common.Hash) *types.Transaction {
	return pool.all.Get(hash)
}

// removeTx removes a single transaction from the queue, moving all subsequent
// transactions back to the future queue.
func (pool *TxPool) removeAllTx(hash common.Hash) {
	// Fetch the transaction we wish to delete
	tx := pool.all.Get(hash)
	if tx == nil {
		return
	}

	pool.priced.Remove(tx)
	delete(pool.beats, hash)
	//Remove it from the list of known transactions
	if pool.newQueue.Remove(tx) {
		return
	}
	if pool.newPending.Remove(tx) {
		return
	}
}

func (pool *TxPool) removeWorkQueue(tx *types.Transaction) {

	delete(pool.beats, tx.Hash())
	//Remove it from the list of known transactions
	if pool.newQueue.Remove(tx) {
		return
	}
	if pool.newPending.Remove(tx) {
		return
	}
}

func (pool *TxPool) promoteTx(tx *types.Transaction) bool {
	// Try to insert the transaction into the pending queue
	if pool.newPending.Add(tx, new(big.Int).Set(pool.gasPrice)) {
		pool.beats[tx.Hash()] = time.Now()
	}
	return true
}

func (pool *TxPool) broadCastLocalTx(tx *types.Transaction) {
	go pool.txFeed.Send(NewTxsEvent{[]*types.Transaction{tx}})
}

func (pool *TxPool) promoteExecutables(broadcast bool) {
	// Track the promoted transactions to broadcast them at once
	var promoted []*types.Transaction
	//var invalidTx []common.Hash
	for _, tx := range pool.newQueue.Ready() {
		//if err := pool.validateTx(tx, false); err != nil {
		//	invalidTx = append(invalidTx, tx.Hash())
		//	continue
		//}
		if pool.promoteTx(tx) {
			log.Trace("Promoting queued transaction", "hash", tx.Hash())
			if (pool.config.StartLight || !pool.Mining()) && broadcast {
				promoted = append(promoted, tx)
			}
		}
	}
	//if len(invalidTx) > 0 {
	//	for _, tx := range invalidTx {
	//		pool.removeAllTx(tx)
	//	}
	//
	//}

	// Notify subsystem for new promoted transactions.
	if len(promoted) > 0 {
		log.Debug("txpool promoted and broadcast txs", "txs", len(promoted))
		//subLen := 100
		//if len(promoted) > subLen {
		//	promoted = promoted[:subLen]
		//}
		go pool.txFeed.Send(NewTxsEvent{promoted})
	}

	// If we've queued more transactions than the hard limit, drop oldest ones
	if uint64(pool.newPending.Len()) > pool.config.GlobalQueue {
		drop := uint64(pool.newPending.Len()) - pool.config.GlobalQueue
		if drop > 0 {
			transactions := pool.newPending.Discard(int(drop))
			for _, tx := range transactions {
				pool.newQueue.Add(tx, big.NewInt(0))
				log.Trace("Removed fairness-exceeding pending transaction", "hash", tx.Hash())
			}
		}
	}
}

// txLookup is used internally by TxPool to track transactions while allowing lookup without
// mutex contention.
//
// Note, although this type is properly protected against concurrent access, it
// is **not** a type that should ever be mutated or even exposed outside of the
// transaction pool, since its internal state is tightly coupled with the pools
// internal mechanisms. The sole purpose of the type is to permit out-of-bound
// peeking into the pool in TxPool.Get without having to acquire the widely scoped
// TxPool.mu mutex.
type txLookup struct {
	all  map[common.Hash]*types.Transaction
	lock sync.RWMutex
}

// newTxLookup returns a new txLookup structure.
func newTxLookup() *txLookup {
	return &txLookup{
		all: make(map[common.Hash]*types.Transaction),
	}
}

// Range calls f on each key and value present in the map.
func (t *txLookup) Range(f func(hash common.Hash, tx *types.Transaction) bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for key, value := range t.all {
		if !f(key, value) {
			break
		}
	}
}

// Get returns a transaction if it exists in the lookup, or nil if not found.
func (t *txLookup) Get(hash common.Hash) *types.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.all[hash]
}

// Count returns the current number of items in the lookup.
func (t *txLookup) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.all)
}

// Add adds a transaction to the lookup.
func (t *txLookup) Add(tx *types.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.all[tx.Hash()] = tx
}

// Remove removes a transaction from the lookup.
func (t *txLookup) Remove(hash common.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.all, hash)
}
