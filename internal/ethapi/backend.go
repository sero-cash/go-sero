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

// Package ethapi implements the general Ethereum API functions.
package ethapi

import (
	"context"
	"math/big"

	"github.com/sero-cash/go-sero/zero/exchange"

	"github.com/sero-cash/go-sero/zero/light/light_types"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/miner"

	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/consensus"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/core/vm"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/params"
	"github.com/sero-cash/go-sero/rpc"
	"github.com/sero-cash/go-sero/sero/downloader"
	"github.com/sero-cash/go-sero/serodb"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Ethereum API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() serodb.Database
	EventMux() *event.TypeMux
	AccountManager() *accounts.Manager

	// BlockChain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
	GetTd(blockHash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription

	// TxPool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	//GetPoolNonce(ctx context.Context, addr common.Data) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (types.Transactions, types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription

	ChainConfig() *params.ChainConfig
	CurrentBlock() *types.Block
	GetEngin() consensus.Engine
	GetMiner() *miner.Miner

	GetBlocksInfo(start uint64, count uint64) ([]light_types.Block, error)
	GetAnchor(roots []keys.Uint256) ([]light_types.Witness, error)
	CommitTx(tx *light_types.GTx) error

	GetPkr(address keys.Uint512, index uint64) (keys.PKr, error)
	GetBalances(address keys.Uint512) (balances map[string]*big.Int)
	GenTx(param exchange.TxParam) (*light_types.GenTxParam, error)
	GenTxWithSign(param exchange.TxParam) (*light_types.GTx, error)
	GetRecords(address keys.PKr, begin, end uint64) (records []exchange.Utxo, err error)
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "ssi",
			Version:   "1.0",
			Service:   &PublicSSIAPI{apiBackend},
			Public:    true,
		},
		{
			Namespace: "exchange",
			Version:   "1.0",
			Service:   &PublicExchangeAPI{apiBackend},
			Public:    true,
		},
		{
			Namespace: "sero",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "sero",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "sero",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "sero",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		},
	}
}
