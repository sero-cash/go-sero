// copyright 2018 The go-ethereum Authors
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

package sero

import (
	"context"
	"errors"
	"github.com/sero-cash/go-sero/common/hexutil"
	"math/big"

	"github.com/sero-cash/go-sero/zero/exchange"

	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-sero/zero/light"

	"github.com/sero-cash/go-sero/zero/light/light_types"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/consensus"
	"github.com/sero-cash/go-sero/miner"

	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/bloombits"
	"github.com/sero-cash/go-sero/core/rawdb"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/core/vm"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/params"
	"github.com/sero-cash/go-sero/rpc"
	"github.com/sero-cash/go-sero/sero/downloader"
	"github.com/sero-cash/go-sero/sero/gasprice"
	"github.com/sero-cash/go-sero/serodb"
)

// SeroAPIBackend implements ethapi.Backend for full nodes
type SeroAPIBackend struct {
	sero *Sero
	gpo  *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *SeroAPIBackend) ChainConfig() *params.ChainConfig {
	return b.sero.chainConfig
}

func (b *SeroAPIBackend) CurrentBlock() *types.Block {
	return b.sero.blockchain.CurrentBlock()
}

func (b *SeroAPIBackend) GetEngin() consensus.Engine {
	return b.sero.engine
}

func (b *SeroAPIBackend) GetMiner() *miner.Miner {
	return b.sero.miner
}

func (b *SeroAPIBackend) SetHead(number uint64) {
	b.sero.protocolManager.downloader.Cancel()
	b.sero.blockchain.SetHead(number, core.DelFn)
}

func (b *SeroAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.sero.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.sero.blockchain.CurrentBlock().Header(), nil
	}
	return b.sero.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *SeroAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.sero.blockchain.GetHeaderByHash(hash), nil
}

func (b *SeroAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.sero.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.sero.blockchain.CurrentBlock(), nil
	}
	return b.sero.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *SeroAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.sero.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.sero.BlockChain().StateAt(header.Root, header.Number.Uint64())
	return stateDb, header, err
}

func (b *SeroAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.sero.blockchain.GetBlockByHash(hash), nil
}

func (b *SeroAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.sero.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.sero.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *SeroAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.sero.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.sero.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *SeroAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.sero.blockchain.GetTdByHash(blockHash)
}

func (b *SeroAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.sero.BlockChain(), nil)
	return vm.NewEVM(context, state, b.sero.chainConfig, vmCfg), vmError, nil
}

func (b *SeroAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.sero.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *SeroAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.sero.BlockChain().SubscribeChainEvent(ch)
}

func (b *SeroAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.sero.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *SeroAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.sero.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *SeroAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.sero.BlockChain().SubscribeLogsEvent(ch)
}

func (b *SeroAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.sero.txPool.AddLocal(signedTx)
}

func (b *SeroAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.sero.txPool.Pending()
	if err != nil {
		return nil, err
	}

	return pending, nil
}

func (b *SeroAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.sero.txPool.Get(hash)
}

//func (b *SeroAPIBackend) GetPoolNonce(ctx context.Context, addr common.Data) (uint64, error) {
//	return b.sero.txPool.State().GetNonce(addr), nil
//}

func (b *SeroAPIBackend) Stats() (pending int, queued int) {
	return b.sero.txPool.Stats()
}

func (b *SeroAPIBackend) TxPoolContent() (types.Transactions, types.Transactions) {
	return b.sero.TxPool().Content()
}

func (b *SeroAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.sero.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *SeroAPIBackend) Downloader() *downloader.Downloader {
	return b.sero.Downloader()
}

func (b *SeroAPIBackend) ProtocolVersion() int {
	return b.sero.EthVersion()
}

func (b *SeroAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *SeroAPIBackend) ChainDb() serodb.Database {
	return b.sero.ChainDb()
}

func (b *SeroAPIBackend) EventMux() *event.TypeMux {
	return b.sero.EventMux()
}

func (b *SeroAPIBackend) AccountManager() *accounts.Manager {
	return b.sero.AccountManager()
}

func (b *SeroAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.sero.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *SeroAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.sero.bloomRequests)
	}
}

func (b *SeroAPIBackend) GetBlocksInfo(start uint64, count uint64) ([]light_types.Block, error) {
	return light.SRI_Inst.GetBlocksInfo(start, count)

}
func (b *SeroAPIBackend) GetAnchor(roots []keys.Uint256) ([]light_types.Witness, error) {
	return light.SRI_Inst.GetAnchor(roots)

}
func (b *SeroAPIBackend) CommitTx(tx *light_types.GTx) error {
	gasPrice := big.Int(tx.GasPrice)
	gas := uint64(tx.Gas)
	signedTx := types.NewTxWithGTx(gas, &gasPrice, &tx.Tx)
	log.Info("commitTx", "txhash", signedTx.Hash().String())
	return b.sero.txPool.AddLocal(signedTx)
}

func (b *SeroAPIBackend) GetPkNumber(pk keys.Uint512) (number uint64, e error) {
	if b.sero.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.sero.exchange.GetCurrencyNumber(pk), nil
}

func (b *SeroAPIBackend) GetPkr(address *keys.Uint512, index *keys.Uint256) (pkr keys.PKr, e error) {
	if b.sero.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.sero.exchange.GetPkr(address, index)
}

func (b *SeroAPIBackend) GetLockedBalances(address keys.Uint512) (balances map[string]*big.Int) {
	if b.sero.exchange == nil {
		return
	}
	return b.sero.exchange.GetLockedBalances(address)
}

func (b *SeroAPIBackend) GetMaxAvailable(pk keys.Uint512, currency string) (amount *big.Int) {
	if b.sero.exchange == nil {
		return
	}
	return b.sero.exchange.GetMaxAvailable(pk, currency)
}

func (b *SeroAPIBackend) GetBalances(address keys.Uint512) (balances map[string]*big.Int) {
	if b.sero.exchange == nil {
		return
	}
	return b.sero.exchange.GetBalances(address)
}

func (b *SeroAPIBackend) GenTx(param exchange.TxParam) (txParam *light_types.GenTxParam, e error) {
	if b.sero.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.sero.exchange.GenTx(param)
}

func (b *SeroAPIBackend) GenTxWithSign(param exchange.TxParam) (gtx *light_types.GTx, e error) {
	if b.sero.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.sero.exchange.GenTxWithSign(param)
}

func (b *SeroAPIBackend) GetRecords(address hexutil.Bytes, begin, end uint64) (records []exchange.Utxo, err error) {
	if b.sero.exchange == nil {
		err = errors.New("not start exchange")
		return
	}
	pkr := keys.PKr{}
	copy(pkr[:], address)
	return b.sero.exchange.GetRecords(pkr, begin, end)
}
