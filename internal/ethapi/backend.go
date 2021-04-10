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

	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-sero/zero/wallet/exchange"

	"github.com/sero-cash/go-czero-import/c_type"

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
	"github.com/sero-cash/go-sero/zero/wallet/light"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Ethereum API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	PeerCount() uint
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
	Stats() (pending int, queued int, all int, total int)
	TxPoolContent() (types.Transactions, types.Transactions, types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription

	ChainConfig() *params.ChainConfig
	CurrentBlock() *types.Block
	GetEngin() consensus.Engine
	GetMiner() *miner.Miner

	GetBlocksInfo(start uint64, count uint64) ([]txtool.Block, error)
	GetAnchor(roots []c_type.Uint256) ([]txtool.Witness, error)
	CommitTx(tx *txtool.GTx) error
	GetCommittedTx(txHash c_type.Uint256) (*txtool.GTx, error)
	ReSendCommittedTx(txHash c_type.Uint256) error

	GetPkNumber(pk c_type.Uint512) (number uint64, e error)
	GetPkr(pk *c_type.Uint512, index *c_type.Uint256) (c_type.PKr, error)
	GetBalances(pk c_type.Uint512) (balances map[string]*big.Int, tickets map[string][]*common.Hash)
	GenTx(param prepare.PreTxParam) (*txtool.GTxParam, error)
	GetRecordsByPk(pk *c_type.Uint512, begin, end uint64) (records []exchange.Utxo, err error)
	GetRecordsByPkr(pkr c_type.PKr, begin, end uint64) (records []exchange.Utxo, err error)
	GetLockedBalances(pk c_type.Uint512) (balances map[string]*big.Int)
	GetMaxAvailable(pk c_type.Uint512, currency string) (amount *big.Int)
	GetRecordsByTxHash(txHash c_type.Uint256) (records []exchange.Utxo, err error)

	//Light node api
	GetOutByPKr(pkrs []c_type.PKr, start, end uint64) (br light.BlockOutResp, e error)
	CheckNil(Nils []c_type.Uint256) (nilResps []light.NilValue, e error)
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "proof",
			Version:   "1.0",
			Service:   NewProofServiceApi(),
			Public:    true,
		},
		{
			Namespace: "stake",
			Version:   "1.0",
			Service:   NewPublicStakeApI(apiBackend, nonceLock),
			Public:    true,
		},
		{
			Namespace: "sero",
			Version:   "1.0",
			Service:   &PublicAbiAPI{},
			Public:    true,
		},
		{
			Namespace: "light",
			Version:   "1.0",
			Service:   &PublicLightNodeApi{apiBackend},
			Public:    true,
		},
		{
			Namespace: "ssi",
			Version:   "1.0",
			Service:   &PublicSSIAPI{apiBackend},
			Public:    true,
		},
		{
			Namespace: "local",
			Version:   "1.0",
			Service:   &PublicLocalAPI{},
			Public:    true,
		},
		{
			Namespace: "flight",
			Version:   "1.0",
			Service:   &PublicFlightAPI{&PublicExchangeAPI{apiBackend}},
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
