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

package ethapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-sero/zero/stake"

	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-sero/zero/wallet/exchange"

	"github.com/tyler-smith/go-bip39"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero/zero/txs"

	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/davecgh/go-spew/spew"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/consensus/ethash"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/rawdb"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/core/vm"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/p2p"
	"github.com/sero-cash/go-sero/params"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/rpc"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	ztx "github.com/sero-cash/go-sero/zero/txs/tx"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	defaultGasPrice = params.Gta
)

var (
	zerobyte = string([]byte{0})
)

// PublicEthereumAPI provides an API to access Ethereum related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicEthereumAPI struct {
	b Backend
}

// NewPublicEthereumAPI creates a new Ethereum protocol API.
func NewPublicEthereumAPI(b Backend) *PublicEthereumAPI {
	return &PublicEthereumAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicEthereumAPI) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	price, err := s.b.SuggestPrice(ctx)
	return (*hexutil.Big)(price), err
}

// ProtocolVersion returns the current Ethereum protocol version this node supports
func (s *PublicEthereumAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.b.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicEthereumAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronisation already completed
	if progress.CurrentBlock >= progress.HighestBlock {
		return false, nil
	}
	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(progress.StartingBlock),
		"currentBlock":  hexutil.Uint64(progress.CurrentBlock),
		"highestBlock":  hexutil.Uint64(progress.HighestBlock),
		"pulledStates":  hexutil.Uint64(progress.PulledStates),
		"knownStates":   hexutil.Uint64(progress.KnownStates),
	}, nil
}

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicTxPoolAPI struct {
	b Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(b Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{b}
}

// Content returns the transactions contained within the transaction pool.

func (s *PublicTxPoolAPI) Content() map[string]map[string]*RPCTransaction {
	content := map[string]map[string]*RPCTransaction{
		"pending": make(map[string]*RPCTransaction),
		"queued":  make(map[string]*RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()

	// Flatten the pending transactions

	dump := make(map[string]*RPCTransaction)
	for _, tx := range pending {
		dump[tx.Hash().Hex()] = newRPCPendingTransaction(tx)
	}
	content["pending"] = dump

	// Flatten the queued transactions

	qdump := make(map[string]*RPCTransaction)
	for _, tx := range queue {
		qdump[tx.Hash().Hex()] = newRPCPendingTransaction(tx)
	}
	content["queued"] = qdump

	return content
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	pending, queue := s.b.Stats()
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(pending),
		"queued":  hexutil.Uint(queue),
	}
}

// Inspect retrieves the content of the transaction pool and flattens it into an
// easily inspectable list.

func (s *PublicTxPoolAPI) Inspect() map[string]map[string]string {
	content := map[string]map[string]string{
		"pending": make(map[string]string),
		"queued":  make(map[string]string),
	}
	/*pending, queue := s.b.TxPoolContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx *types.Transaction) string {
		if to := tx.To(); to != nil {
			return fmt.Sprintf("%s:  %v gas × %v wei", tx.To().Base58(), tx.Gas(), tx.GasPrice())
		}
		return fmt.Sprintf("contract creation: %v gas × %v wei", tx.Gas(), tx.GasPrice())
	}
	// Flatten the pending transactions

	dump := make(map[string]string)
	for _, tx := range pending {
		dump[fmt.Sprintf("%s", tx.Hash())] = format(tx)
	}
	content["pending"] = dump

	// Flatten the queued transactions
	qdump := make(map[string]string)
	for _, tx := range queue {
		qdump[fmt.Sprintf("%s", tx.Hash())] = format(tx)
	}
	content["queued"] = qdump*/
	return content
}

// PublicAccountAPI provides an API to access accounts managed by this node.
// It offers only methods that can retrieve accounts.
type PublicAccountAPI struct {
	am *accounts.Manager
}

// NewPublicAccountAPI creates a new PublicAccountAPI.
func NewPublicAccountAPI(am *accounts.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{am: am}
}

func (s *PublicAccountAPI) GetSzkAccounts() []string {
	addresses := make([]string, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			pk := c_superzk.Tk2Pk(account.Tk.ToUint512().NewRef())
			addr := utils.NewAddressByBytes(pk[:])
			addr.SetProtocol("SP")
			addresses = append(addresses, addr.ToCode())
		}
	}
	return addresses
}

// Accounts returns the collection of accounts this node manages
func (s *PublicAccountAPI) Accounts() []address.AccountAddress {
	addresses := make([]address.AccountAddress, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

func (s *PublicAccountAPI) GetTk(addr common.Address) address.AccountAddress {
	var pk address.AccountAddress
	wallets := s.am.Wallets()

	if addr.IsAccountAddress() {
		pk = common.AddrToAccountAddr(addr)
	} else {
		pkAddrr := getLocalAccountAddressByPkr(wallets, addr)
		if pkAddrr != nil {
			pk = *pkAddrr
		} else {
			return address.AccountAddress{}
		}
	}
	for _, wallet := range wallets {
		account := accounts.Account{Address: pk}
		if wallet.Contains(account) {
			return wallet.Accounts()[0].Tk
		}
	}
	return address.AccountAddress{}

}

func (s *PublicAccountAPI) IsMinePKr(PKr PKrAddress) *address.AccountAddress {
	wallets := s.am.Wallets()
	var address common.Address
	copy(address[:], PKr[:])
	return getLocalAccountAddressByPkr(wallets, address)

}

func getLocalAccountAddressByPkr(wallets []accounts.Wallet, PKr common.Address) *address.AccountAddress {

	if len(wallets) == 0 {
		return nil
	}
	if superzk.IsPKrValid(PKr.ToPKr()) {
		for _, wallet := range wallets {
			if wallet.IsMine(PKr) {
				return &wallet.Accounts()[0].Address
			}
		}
	} else {
		log.Debug("getLocalAccountAddressByPkr invalid pkr", "pkr", PKr.String())
		return nil
	}
	return nil
}

// PrivateAccountAPI provides an API to access accounts managed by this node.
// It offers methods to create, (un)lock en list accounts. Some methods accept
// passwords and are therefore considered private by default.
type PrivateAccountAPI struct {
	am        *accounts.Manager
	nonceLock *AddrLocker
	b         Backend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateAccountAPI(b Backend, nonceLock *AddrLocker) *PrivateAccountAPI {
	return &PrivateAccountAPI{
		am:        b.AccountManager(),
		nonceLock: nonceLock,
		b:         b,
	}
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (s *PrivateAccountAPI) ListAccounts() []address.AccountAddress {
	addresses := make([]address.AccountAddress, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// rawWallet is a JSON representation of an accounts.Wallet interface, with its
// data contents extracted into plain fields.
type rawWallet struct {
	URL      string             `json:"url"`
	Status   string             `json:"status"`
	Failure  string             `json:"failure,omitempty"`
	Accounts []accounts.Account `json:"accounts,omitempty"`
}

// ListWallets will return a list of wallets this node manages.
func (s *PrivateAccountAPI) ListWallets() []rawWallet {
	wallets := make([]rawWallet, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		status, failure := wallet.Status()

		raw := rawWallet{
			URL:      wallet.URL().String(),
			Status:   status,
			Accounts: wallet.Accounts(),
		}
		if failure != nil {
			raw.Failure = failure.Error()
		}
		wallets = append(wallets, raw)
	}
	return wallets
}

// OpenWallet initiates a hardware wallet opening procedure, establishing a USB
// connection and attempting to authenticate via the provided passphrase. Note,
// the method may return an extra challenge requiring a second open (e.g. the
// Trezor PIN matrix challenge).
func (s *PrivateAccountAPI) OpenWallet(url string, passphrase *string) error {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return err
	}
	pass := ""
	if passphrase != nil {
		pass = *passphrase
	}
	return wallet.Open(pass)
}

// DeriveAccount requests a HD wallet to derive a new account, optionally pinning
// it for later reuse.
func (s *PrivateAccountAPI) DeriveAccount(url string, path string, pin *bool) (accounts.Account, error) {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return accounts.Account{}, err
	}
	derivPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return accounts.Account{}, err
	}
	if pin == nil {
		pin = new(bool)
	}
	return wallet.Derive(derivPath, *pin)
}

// NewAccount will create a new account and returns the address for the new account.
func (s *PrivateAccountAPI) NewAccount(password string) (address.AccountAddress, error) {
	blockNum := uint64(0)
	if seroparam.Is_Dev() {
		blockNum = uint64(0)
	} else {
		current := s.b.CurrentBlock()
		if current != nil {
			blockNum = current.NumberU64()
		}
	}
	acc, err := fetchKeystore(s.am).NewAccount(password, blockNum)
	if err != nil {
		return address.AccountAddress{}, err
	}

	if seroparam.Is_Dev() {
		fetchKeystore(s.am).TimedUnlock(acc, password, 0)
	}
	return acc.Address, nil
}

// NewAccount will create a new account and returns the mnemonic 、address for the new account.
func (s *PrivateAccountAPI) NewAccountWithMnemonic(password string) (map[string]string, error) {
	blockNum := uint64(0)
	if seroparam.Is_Dev() {
		blockNum = uint64(0)
	} else {
		current := s.b.CurrentBlock()
		if current != nil {
			blockNum = current.NumberU64()
		}
	}
	mnemonic, acc, err := fetchKeystore(s.am).NewAccountWithMnemonic(password, blockNum)
	if err != nil {
		return nil, err
	}

	if seroparam.Is_Dev() {
		fetchKeystore(s.am).TimedUnlock(acc, password, 0)
	}
	result := map[string]string{}
	result["mnemonic"] = mnemonic
	result["address"] = acc.Address.Base58()
	return result, nil
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

// ImportRawKey stores the given hex encoded ECDSA key into the key directory,
// encrypting it with the passphrase.
func (s *PrivateAccountAPI) ImportRawKey(privkey string, password string) (address.AccountAddress, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return address.AccountAddress{}, err
	}
	acc, err := fetchKeystore(s.am).ImportECDSA(key, password)
	return acc.Address, err
}

func (s *PrivateAccountAPI) ImportTk(tk address.AccountAddress) (address.AccountAddress, error) {
	acc, err := fetchKeystore(s.am).ImportTk(tk)
	return acc.Address, err
}

func (s *PrivateAccountAPI) ImportMnemonic(mnemonic string, password string) (address.AccountAddress, error) {
	_, err := bip39.MnemonicToByteArray(mnemonic)
	if err != nil {
		return address.AccountAddress{}, err
	}
	seed, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return address.AccountAddress{}, err
	}
	if len(seed) != 32 {
		return address.AccountAddress{}, errors.New("EntropyFromMnemonic error seed not 256bits")
	}
	key, err := crypto.ToECDSA(seed[:32])
	if err != nil {
		return address.AccountAddress{}, err
	}
	acc, err := fetchKeystore(s.am).ImportECDSA(key, password)
	return acc.Address, err
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
func (s *PrivateAccountAPI) UnlockAccount(addr address.AccountAddress, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	if seroparam.Is_Dev() {
		d = 0
	}
	err := fetchKeystore(s.am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

func (s *PrivateAccountAPI) ExportMnemonic(addr address.AccountAddress, password string) (string, error) {
	return fetchKeystore(s.am).ExportMnemonic(accounts.Account{Address: addr}, password)
}

func (s *PrivateAccountAPI) ExportRawKey(addr address.AccountAddress, password string) (hexutil.Bytes, error) {
	seed, err := fetchKeystore(s.am).ExportRewKey(accounts.Account{Address: addr}, password)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(seed), nil
}

func (s *PrivateAccountAPI) GenSeed() (hexutil.Bytes, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(entropy), nil
}

// LockAccount will lock the account associated with the given address when it's unlocked.
func (s *PrivateAccountAPI) LockAccount(addr address.AccountAddress) bool {
	return fetchKeystore(s.am).Lock(addr) == nil
}

type threaded interface {
	SetThreads(threads int)
	Threads() int
}

// signTransactions sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, if applicable,
// and release it after the transaction has been submitted to the tx pool
func (s *PrivateAccountAPI) signTransaction(ctx context.Context, args SendTxArgs, passwd string) (pretx *txtool.GTxParam, tx *types.Transaction, e error) {
	s.nonceLock.mu.Lock()
	defer s.nonceLock.mu.Unlock()
	// Look up the wallet containing the requested abi
	account := accounts.Account{Address: args.From}
	wallet, err := s.am.Find(account)
	if err != nil {
		e = err
		return
	}
	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		e = err
		return
	}

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)

	if err != nil {
		e = err
		return
	}

	var txt *ztx.T
	// Assemble the transaction and sign with the wallet
	tx, txt = args.toTransaction(state)

	if th, ok := s.b.GetEngin().(threaded); ok {
		miner := s.b.GetMiner()
		if miner.CanStart() {
			miner.SetCanStart(0)
			defer miner.SetCanStart(1)
		}
		threads := th.Threads()
		if threads >= 0 {
			th.SetThreads(-1)
			defer th.SetThreads(threads)
		}
	}
	if !seroparam.IsExchange() {
		tx, e = wallet.EncryptTxWithPassphrase(account, passwd, tx, txt, state)
		return
	} else {
		if pretx, e = s.b.GenTx(args.toTxParam(state)); e != nil {
			return
		}
		log.Info("ToTxParam", "utxos", len(pretx.Ins))
		seed, err := wallet.GetSeedWithPassphrase(passwd)
		if err != nil {
			exchange.CurrentExchange().ClearTxParam(pretx)
			e = err
			return
		}
		sk := superzk.Seed2Sk(seed.SeedToUint256())
		gtx, err := flight.SignTx(&sk, pretx)
		if err != nil {
			exchange.CurrentExchange().ClearTxParam(pretx)
			e = err
			return
		}
		gasPrice := big.Int(gtx.GasPrice)
		gas := uint64(gtx.Gas)
		tx = types.NewTxWithGTx(gas, &gasPrice, &gtx.Tx)
		return

	}

}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails.
func (s *PrivateAccountAPI) SendTransaction(ctx context.Context, args SendTxArgs, passwd string) (common.Hash, error) {
	/*if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}*/
	pretx, signed, err := s.signTransaction(ctx, args, passwd)
	if err != nil {
		return common.Hash{}, err
	}

	if err := s.b.SendTx(ctx, signed); err != nil {
		if pretx != nil {
			exchange.CurrentExchange().ClearTxParam(pretx)
		}
		return common.Hash{}, err
	}
	if pretx != nil {
		log.Info("Submitted transaction", "fullhash", signed.Hash().Hex(), "recipient", args.To, "utxo", len(pretx.Ins))
	} else {
		log.Info("Submitted transaction", "fullhash", signed.Hash().Hex(), "recipient", args.To)
	}
	return signed.Hash(), nil
}

// SignAndSendTransaction was renamed to SendTransaction. This method is deprecated
// and will be removed in the future. It primary goal is to give clients time to update.
func (s *PrivateAccountAPI) SignAndSendTransaction(ctx context.Context, args SendTxArgs, passwd string) (common.Hash, error) {
	return s.SendTransaction(ctx, args, passwd)
}

// PublicBlockChainAPI provides an API to access the Ethereum blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new Ethereum blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	return &PublicBlockChainAPI{b}
}

// BlockNumber returns the block number of the chain head.
func (s *PublicBlockChainAPI) BlockNumber() hexutil.Uint64 {
	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
	return hexutil.Uint64(header.Number.Uint64())
}

func (s *PublicBlockChainAPI) CurrencyToContractAddress(ctx context.Context, cy string) (*address.AccountAddress, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	if cy == "" {
		return nil, errors.New("cy can not be empty!")
	} else {
		if cy == "sero" || cy == "SERO" {
			return nil, nil
		}
	}
	contractAddress := state.GetContrctAddressByToken(cy)
	empty := common.Address{}
	if contractAddress == empty {
		return nil, errors.New(cy + "not exists!")
	}
	contractAddr := address.BytesToAccount(contractAddress[:64])
	return &contractAddr, nil
}

type ConvertAddress struct {
	Addr      map[address.AccountAddress]common.Address `json:"addr"`
	ShortAddr map[common.Address]common.ContractAddress `json:"shortAddr"`
	Rand      *c_type.Uint128                           `json:"rand"`
}

func (s *PublicBlockChainAPI) ConvertAddressParams(ctx context.Context, rand *c_type.Uint128, addresses []address.AccountAddress, dy bool) (*ConvertAddress, error) {
	empty := &c_type.Uint128{}
	if bytes.Equal(rand[:], empty[:]) {
		randKey := c_type.RandUint128()
		rand = &randKey
	}
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}

	addrMap := map[address.AccountAddress]common.Address{}
	shortAddrMap := map[common.Address]common.ContractAddress{}

	randSeed := rand.ToUint256()

	if dy {
		randUint128 := c_type.RandUint128()
		randSeed = (&randUint128).ToUint256()
	}
	for _, addr := range addresses {
		onceAddr := common.Address{}
		if state.IsContract(common.BytesToAddress(addr[:])) {
			onceAddr = common.BytesToAddress(addr[:])
		} else {
			if !superzk.IsPKValid(addr.ToUint512()) {
				return nil, errors.New("address must be accountAddress")
			}
			pkr := superzk.Pk2PKr(addr.ToUint512(), randSeed.NewRef())
			onceAddr.SetBytes(pkr[:])
		}
		addrMap[addr] = onceAddr
		shortAddr := superzk.HashPKr(onceAddr.ToPKr())
		shortAddrMap[onceAddr] = common.BytesToContractAddress(shortAddr[:])
	}
	return &ConvertAddress{addrMap, shortAddrMap, rand}, nil
}

func (s *PublicBlockChainAPI) GetFullAddress(ctx context.Context, shortAddresses []common.ContractAddress) (map[common.ContractAddress]common.Address, error) {

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	addrMap := map[common.ContractAddress]common.Address{}
	for _, short := range shortAddresses {
		full := state.GetNonceAddress(short[:])
		//
		//wallets := s.b.AccountManager().Wallets()
		//
		//if len(wallets) > 0 {
		//	for _, wallet := range wallets {
		//		if wallet.IsMine(full) {
		//			full = common.BytesToAddress(wallet.Accounts()[0].Address[:])
		//			break
		//		}
		//	}
		//}
		addrMap[short] = full
	}
	return addrMap, nil

}

func (s *PublicBlockChainAPI) GenPKr(ctx context.Context, Pk PKAddress) (common.Address, error) {
	PKr := superzk.Pk2PKr(Pk.ToUint512().NewRef(), nil)
	result := common.Address{}
	copy(result[:], PKr[:])
	return result, nil
}

type Balance struct {
	Tkn map[string]*hexutil.Big   `json:"tkn"`
	Tkt map[string][]*common.Hash `json:"tkt"`
}

func GetBalanceFromExchange(tkns map[string]*big.Int) (result Balance) {
	tkn := map[string]*hexutil.Big{}
	if tkns != nil {
		for cy, value := range tkns {
			if tkn[cy] == nil {
				tkn[cy] = (*hexutil.Big)(value)
			} else {
				tkn[cy] = (*hexutil.Big)(new(big.Int).Add((*big.Int)(tkn[cy]), value))
			}
		}
	}
	if len(tkn) > 0 {
		result.Tkn = tkn
	}

	return
}

// GetBalance returns the amount of wei for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicBlockChainAPI) GetBalance(ctx context.Context, addr common.Address, blockNr rpc.BlockNumber) (Balance, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)

	if state == nil || err != nil {
		return Balance{}, err
	}
	var address *address.AccountAddress

	address = getAccountAddress(addr, s.b)
	if address == nil {
		return Balance{}, nil
	}

	tkn := map[string]*hexutil.Big{}
	result := Balance{}
	if size := state.GetCodeSize(common.BytesToAddress(address[:])); size > 0 {
		balances := state.Balances(common.BytesToAddress(address[:]))
		for key, value := range balances {
			tkn[key] = (*hexutil.Big)(value)
		}
		if len(tkn) > 0 {
			result.Tkn = tkn
		}
		return result, state.Error()
	} else {

		if seroparam.IsExchange() {
			exchangBalance := s.b.GetBalances(*address.ToUint512())
			return GetBalanceFromExchange(exchangBalance), nil
		} else {
			return result, errors.New("lstate.balance is no longer supported")
		}
	}

}

func getAccountAddress(addr common.Address, b Backend) *address.AccountAddress {
	//accountAddr = &address.AccountAddress{}
	if addr.IsAccountAddress() {
		accountAddr := common.AddrToAccountAddr(addr)
		return &accountAddr
	} else {
		wallets := b.AccountManager().Wallets()
		return getLocalAccountAddressByPkr(wallets, addr)

	}
	return nil
}

/**
func (s *PublicBlockChainAPI) GetPkg(ctx context.Context, addr common.Address, packed bool, id *c_type.Uint256) (interface{}, error) {

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	wallets := s.b.AccountManager().Wallets()
	accountAddress := getAccountAddress(addr, s.b)
	if accountAddress == nil {
		return nil, nil
	}

	if state.IsContract(common.BytesToAddress(accountAddress[:])) {
		return nil, errors.New("does not support contract address!")
	}
	// Look up the wallet containing the requested abi
	account := accounts.Account{Address: *accountAddress}
	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	seed := wallet.Accounts()[0].Tk
	pkgs := lstate.CurrentLState().GetPkgs(seed.ToUint512(), packed)
	if len(pkgs) > 0 {
		result := []map[string]interface{}{}
		for _, p := range pkgs {
			pkg := map[string]interface{}{}

			pkg["id"] = p.Pkg.Z.Pack.Id
			pkg["packed"] = packed
			to := getLocalAccountAddressByPkr(wallets, common.BytesToAddress(p.Pkg.Z.Pack.PKr[:]))
			if to != nil {
				pkg["to_addr"] = to
			} else {
				pkg["to"] = common.BytesToAddress(p.Pkg.Z.Pack.PKr[:]).String()
			}
			if (p.Key != c_type.Uint256{}) {
				pkg["key"] = p.Key
				asset := map[string]interface{}{}
				if p.Pkg.O.Asset.Tkn != nil {
					tkn := map[string]interface{}{}
					tkn["currency"] = strings.Trim(string(p.Pkg.O.Asset.Tkn.Currency[:]), zerobyte)
					tkn["value"] = p.Pkg.O.Asset.Tkn.Value
					asset["tkn"] = tkn
				}
				if p.Pkg.O.Asset.Tkt != nil {
					tkt := map[string]interface{}{}
					tkt["category"] = strings.Trim(string(p.Pkg.O.Asset.Tkt.Category[:]), zerobyte)
					tkt["value"] = p.Pkg.O.Asset.Tkt.Value
					asset["tkt"] = tkt
				}

				pkg["asset"] = asset

			}
			if id != nil {
				if p.Pkg.Z.Pack.Id == *id {
					return pkg, nil
				} else {
					continue
				}
			} else {
				result = append(result, pkg)
			}

		}
		return result, nil
	}
	return nil, nil
}
**/
func (s *PublicBlockChainAPI) WatchPkg(ctx context.Context, id c_type.Uint256, key c_type.Uint256) (map[string]interface{}, error) {

	pkg_o, pkr, err := txs.WatchPkg(&id, &key)
	if err != nil {
		return nil, err
	}
	wallets := s.b.AccountManager().Wallets()
	pkg := map[string]interface{}{}
	pkg["id"] = id
	pkg["key"] = key
	to := getLocalAccountAddressByPkr(wallets, common.BytesToAddress(pkr[:]))
	if to != nil {
		pkg["to_addr"] = to
	} else {
		pkg["to"] = common.BytesToAddress(pkr[:]).String()
	}
	asset := map[string]interface{}{}
	if pkg_o.Asset.Tkn != nil {
		tkn := map[string]interface{}{}
		tkn["currency"] = strings.Trim(string(pkg_o.Asset.Tkn.Currency[:]), zerobyte)
		tkn["value"] = pkg_o.Asset.Tkn.Value
		asset["tkn"] = tkn
	}
	if pkg_o.Asset.Tkt != nil {
		tkt := map[string]interface{}{}
		tkt["category"] = strings.Trim(string(pkg_o.Asset.Tkt.Category[:]), zerobyte)
		tkt["value"] = pkg_o.Asset.Tkt.Value
		asset["tkt"] = tkt
	}

	pkg["asset"] = asset

	return pkg, nil
}

func (s *PublicBlockChainAPI) GetBlockInfo(ctx context.Context, start hexutil.Uint64, count hexutil.Uint64) ([]txtool.Block, error) {
	block, err := s.b.GetBlocksInfo(uint64(start), uint64(count))
	if err != nil {
		return nil, err
	}
	return block, err
}

func (s *PublicBlockChainAPI) GetAnchor(ctx context.Context, roots []c_type.Uint256) ([]txtool.Witness, error) {
	witness, err := s.b.GetAnchor(roots)
	if err != nil {
		return nil, err
	}
	return witness, err
}

// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response, err := s.rpcOutputBlock(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

//pow reward
func (s *PublicBlockChainAPI) GetBlockRewardByNumber(ctx context.Context, blockNr rpc.BlockNumber) [3]hexutil.Big {
	var res [3]hexutil.Big
	zero := big.NewInt(0)
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		rewards := GetBlockReward(block)
		res[0] = hexutil.Big(*rewards[0])
		res[1] = hexutil.Big(*rewards[1])
		res[2] = hexutil.Big(*zero)
	}
	return res
}

//block reward
func (s *PublicBlockChainAPI) GetBlockTotalRewardByNumber(ctx context.Context, blockNr rpc.BlockNumber) hexutil.Big {

	reward := big.NewInt(0)
	block, _ := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		pows := GetBlockReward(block)
		for _, p := range pows {
			reward.Add(reward, p)
		}
	}
	if block != nil && blockNr >= 1300000 {
		state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)

		if err != nil {
			return hexutil.Big(*reward)
		}
		stakeState := stake.NewStakeState(state)
		solo, total := stakeState.StakeCurrentReward(big.NewInt(int64(blockNr)))
		for _, v := range header.CurrentVotes {
			if v.IsPool {
				reward.Add(reward, total)
			} else {
				reward.Add(reward, solo)
			}
		}
		for _, v := range header.ParentVotes {
			if v.IsPool {
				reward.Add(reward, total)
			} else {
				reward.Add(reward, solo)
			}
		}
	}
	return hexutil.Big(*reward)

}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(block, true, fullTx)
	}
	return nil, err
}

// GetCode returns the code stored at the given address in the state for the given block number.
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	code := state.GetCode(address)
	return code, state.Error()
}

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address common.Address, key string, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	res := state.GetState(address, common.HexToHash(key))
	return res[:], state.Error()
}

type Smbol string

// MarshalText implements encoding.TextMarshaler.
func (s Smbol) MarshalText() ([]byte, error) {
	return []byte(strings.ToUpper(string(s))), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (s *Smbol) UnmarshalText(input []byte) error {
	*s = Smbol(strings.ToUpper(string(input)))
	return nil
}

func (s *Smbol) IsEmpty() bool {
	return (strings.TrimSpace(string(*s)) == "")
}

func (s *Smbol) IsNotEmpty() bool {
	return !s.IsEmpty()
}

func (s *Smbol) IsSero() bool {
	return (strings.ToUpper(strings.TrimSpace(string(*s))) == params.DefaultCurrency)
}

func (s *Smbol) IsNotSero() bool {
	return !s.IsSero()
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From        common.Address  `json:"from"`
	To          *common.Address `json:"to"`
	GasCurrency Smbol           `json:"gasCy"` //default SERO
	Gas         hexutil.Uint64  `json:"gas"`
	GasPrice    hexutil.Big     `json:"gasPrice"`
	Value       hexutil.Big     `json:"value"`
	Data        hexutil.Bytes   `json:"data"`
	Currency    Smbol           `json:"cy"`
	Dynamic     bool            `json:"dy"` //contract address parameters are dynamically generated.
	Category    Smbol           `json:"catg"`
	Tkt         *common.Hash    `json:"tkt"`
}

func (s *PublicBlockChainAPI) doCall(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) ([]byte, uint64, bool, error) {
	defer func(start time.Time) { log.Debug("Executing EVM call finished", "runtime", time.Since(start)) }(time.Now())

	state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)

	if state == nil || err != nil {
		return nil, 0, false, err
	}
	// Set sender address or use a default if none specified
	addr := args.From
	if addr == (common.Address{}) {
		if wallets := s.b.AccountManager().Wallets(); len(wallets) > 0 {
			if accounts := wallets[0].Accounts(); len(accounts) > 0 {
				fromAddr := accounts[0].Address
				addr = common.BytesToAddress(fromAddr[:])
			}
		}
	}
	// Set default gas & gas price if none were set
	gas, gasPrice := uint64(args.Gas), args.GasPrice.ToInt()
	if gas == 0 {
		gas = math.MaxUint64 / 2
	}
	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}

	if args.GasCurrency.IsEmpty() {
		args.GasCurrency = Smbol(params.DefaultCurrency)
	}

	// Create new call message
	//args.Data = args.Data[2:]
	if args.Currency.IsEmpty() {
		args.Currency = Smbol(params.DefaultCurrency)
	}

	var token *assets.Token
	var ticket *assets.Ticket
	if args.Value.ToInt() != nil {
		token = &assets.Token{
			Currency: *(common.BytesToHash(common.LeftPadBytes([]byte(args.Currency), 32)).HashToUint256()),
			Value:    *utils.U256(*args.Value.ToInt()).ToRef(),
		}
	}
	if args.Tkt != nil {
		ticket = &assets.Ticket{
			Category: *(common.BytesToHash(common.LeftPadBytes([]byte(args.Category), 32)).HashToUint256()),
			Value:    *args.Tkt.HashToUint256(),
		}

	}
	asset := assets.Asset{
		Tkn: token,
		Tkt: ticket,
	}
	rand := c_type.RandUint128()
	var to *common.Address
	if args.To != nil && state.IsContract(common.BytesToAddress(args.To[:])) && !args.Dynamic {
		copy(rand[:], args.To[:16])
	}
	if args.To != nil {
		to = &common.Address{}
		if state.IsContract(common.BytesToAddress(args.To[:])) {
			t := common.BytesToAddress(args.To[:])
			to = &t
		} else {
			if args.To.IsAccountAddress() {
				pk := common.AddrToAccountAddr(*args.To)
				pkr := (superzk.Pk2PKr(pk.ToUint512(), nil))
				t := common.BytesToAddress(pkr[:])
				to = &t
			} else {
				to = args.To
			}
		}
	}

	fee := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gas))
	if args.To != nil && state.IsContract(common.BytesToAddress(args.To[:])) && args.GasCurrency.IsNotSero() {
		m, d := state.GetTokenRate(common.BytesToAddress(args.To[:]), string(args.GasCurrency))
		if m.Sign() == 0 || d.Sign() == 0 {
			return nil, 0, false, errors.New("gasCurrency must be SERO or nil")
		}
		state.AddBalance(common.BytesToAddress(args.To[:]), "SERO", fee)
		fee = new(big.Int).Div(fee.Mul(fee, m), d)
	}
	feeToken := assets.Token{
		utils.CurrencyToUint256(string(args.GasCurrency)),
		utils.U256(*fee),
	}
	var fromPkr c_type.PKr
	if addr.IsAccountAddress() {
		fromPkr = superzk.Pk2PKr(addr.ToUint512(), rand.ToUint256().NewRef())
	} else {
		fromPkr = *addr.ToPKr()
	}

	msg := types.NewMessage(common.BytesToAddress(fromPkr[:]), to, 0, asset, feeToken, gasPrice, args.Data)

	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// Get a new instance of the EVM.
	evm, vmError, err := s.b.GetEVM(ctx, msg, state, header, vmCfg)
	if err != nil {
		return nil, 0, false, err
	}
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	res, gas, failed, err := core.ApplyMessage(evm, msg, gp)

	if err := vmError(); err != nil {
		return nil, 0, false, err
	}
	return res, gas, failed, err

}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicBlockChainAPI) Call(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	result, _, _, err := s.doCall(ctx, args, blockNr, vm.Config{}, 5*time.Second)
	return (hexutil.Bytes)(result), err
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicBlockChainAPI) EstimateGas(ctx context.Context, args CallArgs) (hexutil.Uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)
	if uint64(args.Gas) >= params.TxGas {
		hi = uint64(args.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		block, err := s.b.BlockByNumber(ctx, rpc.LatestBlockNumber)
		if err != nil {
			return 0, err
		}
		hi = block.GasLimit()
	}
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) bool {
		args.Gas = hexutil.Uint64(gas)

		_, _, failed, err := s.doCall(ctx, args, rpc.LatestBlockNumber, vm.Config{}, 0)
		if err != nil || failed {
			return false
		}
		return true
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		if !executable(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		if !executable(hi) {
			return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
		}
	}
	return hexutil.Uint64(hi), nil
}

func (s *PublicBlockChainAPI) GetDecimal(ctx context.Context, tokenName string) (*hexutil.Uint, error) {

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	if tokenName == "" {
		return nil, errors.New("tokenName can not be empty!")
	} else {
		if tokenName == "sero" || tokenName == "SERO" {
			return nil, errors.New("tokenName can not be sero!")

		}
	}
	contractAddress := state.GetContrctAddressByToken(tokenName)
	empty := common.Address{}
	if contractAddress == empty {
		return nil, errors.New(tokenName + "not exists!")
	}
	contractAddr := address.BytesToAccount(contractAddress[:64])
	callArgs := CallArgs{
		To: &contractAddress,
	}
	decimals := NewSRC20Decimal(contractAddr, tokenName)
	for _, d := range decimals {
		data, err := d.Pack()
		if err != nil {
			log.Info("SRC20Decimal", "pack", d.method, err)
			continue
		}
		callArgs.Data = data
		res, _, failed, err := s.doCall(ctx, callArgs, rpc.LatestBlockNumber, vm.Config{}, 0)
		if err != nil || failed {
			log.Info("SRC20Decimal", "docall", err)
			continue
		}
		decimal, err := d.Unpack(res)
		if err != nil {
			log.Info("SRC20Decimal", "unpack", err)
			continue
		}
		result := hexutil.Uint(*decimal)
		log.Info("GetDecimal", "contract", contractAddr.String(), "method", d.method, "decimal", *decimal)
		return &result, nil

	}
	return nil, errors.New("contract not support SER20 decimals")
}

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   error              `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

// formatLogs formats EVM returned structured logs for json output
func FormatLogs(logs []vm.StructLog) []StructLogRes {
	formatted := make([]StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = StructLogRes{
			Pc:      trace.Pc,
			Op:      trace.Op.String(),
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Depth:   trace.Depth,
			Error:   trace.Err,
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = fmt.Sprintf("%x", math.PaddedBigBytes(stackValue, 32))
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func RPCMarshalBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             b.Hash(),
		"licr":             hexutil.Bytes(head.Licr.Proof[:]),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(b.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        (*hexutil.Big)(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
		"currentVotes":     head.CurrentVotes,
		"parentVotes":      head.ParentVotes,
	}

	if inclTx {
		formatTx := func(tx *types.Transaction) (interface{}, error) {
			return tx.Hash(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Transaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}
		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range txs {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}

	return fields, nil
}

// rpcOutputBlock uses the generalized output filler, then adds the total difficulty field, which requires
// a `PublicBlockchainAPI`.
func (s *PublicBlockChainAPI) rpcOutputBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	fields, err := RPCMarshalBlock(b, inclTx, fullTx)
	if err != nil {
		return nil, err
	}
	miner_addr := getLocalAccountAddressByPkr(s.b.AccountManager().Wallets(), fields["miner"].(common.Address))
	if miner_addr != nil {
		fields["miner_addr"] = miner_addr
	}
	fields["totalDifficulty"] = (*hexutil.Big)(s.b.GetTd(b.Hash()))
	return fields, err
}

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	Stx              *stx.T          `json:"stx"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	//var abi types.Signer = types.FrontierSigner{}

	//from, _ := types.Sender(abi, tx)

	to := tx.To()

	if to != nil && bytes.Equal(to[:], (&common.Address{})[:]) {
		to = nil
	}
	result := &RPCTransaction{
		From:     tx.From(),
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		To:       to,
		Stx:      tx.Stxt(),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}
	return result
}

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *types.Transaction) *RPCTransaction {
	return newRPCTransaction(tx, common.Hash{}, 0, 0)
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCRawTransactionFromBlockIndex returns the bytes of a transaction given a block and a transaction index.
func newRPCRawTransactionFromBlockIndex(b *types.Block, index uint64) hexutil.Bytes {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	blob, _ := rlp.EncodeToBytes(txs[index])
	return blob
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicTransactionPoolAPI struct {
	b         Backend
	nonceLock *AddrLocker
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicTransactionPoolAPI(b Backend, nonceLock *AddrLocker) *PublicTransactionPoolAPI {
	return &PublicTransactionPoolAPI{b, nonceLock}
}

func (s *PublicTransactionPoolAPI) AddressUnlocked(accountAddr address.AccountAddress) (bool, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: accountAddr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return false, err
	}
	return wallet.AddressUnlocked(account)

}

// GetBlockTransactionCountByNumber returns the number of transactions in the block with the given block number.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block with the given hash.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetTransactionByBlockNumberAndIndex returns the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockNumberAndIndex returns the bytes of the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockHashAndIndex returns the bytes of the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionCount returns the number of transactions the given address has sent for the given block number
func (s *PublicTransactionPoolAPI) GetTransactionCount(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*hexutil.Uint64, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	//nonce := state.GetNonce(address)
	u := hexutil.Uint64(0)
	return &u, state.Error()
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicTransactionPoolAPI) GetTransactionByHash(ctx context.Context, hash common.Hash) *RPCTransaction {
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		return newRPCTransaction(tx, blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		return newRPCPendingTransaction(tx)
	}
	// Transaction unknown, return as such
	return nil
}

// GetRawTransactionByHash returns the bytes of the transaction for the given hash.
func (s *PublicTransactionPoolAPI) GetRawTransactionByHash(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	var tx *types.Transaction

	// Retrieve a finalized transaction, or a pooled otherwise
	if tx, _, _, _ = rawdb.ReadTransaction(s.b.ChainDb(), hash); tx == nil {
		if tx = s.b.GetPoolTransaction(hash); tx == nil {
			// Transaction not found anywhere, abort
			return nil, nil
		}
	}
	// Serialize to RLP and return
	return rlp.EncodeToBytes(tx)
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicTransactionPoolAPI) GetTransactionReceipt(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	receipt := receipts[index]

	//var abi types.Signer = types.FrontierSigner{}
	//
	//from, _ := types.Sender(abi, tx)

	to := tx.To()

	if to != nil && bytes.Equal(to[:], (&common.Address{})[:]) {
		to = nil
	}

	fields := map[string]interface{}{
		"blockHash":         blockHash,
		"blockNumber":       hexutil.Uint64(blockNumber),
		"transactionHash":   hash,
		"transactionIndex":  hexutil.Uint64(index),
		"from":              tx.From(),
		"to":                to,
		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"contractAddress":   nil,
		"logs":              receipt.Logs,
		"logsBloom":         receipt.Bloom,
		"shareId":           receipt.ShareId,
		"poolId":            receipt.PoolId,
	}

	// Assign receipt status or post state.
	if len(receipt.PostState) > 0 {
		fields["root"] = hexutil.Bytes(receipt.PostState)
	}
	fields["status"] = hexutil.Uint(receipt.Status)
	if receipt.Logs == nil {
		fields["logs"] = [][]*types.Log{}
	}
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if receipt.ContractAddress != (common.Address{}) {
		fields["contractAddress"] = address.BytesToAccount(receipt.ContractAddress[:64])
	}
	return fields, nil
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From        address.AccountAddress `json:"from"`
	To          *common.Address        `json:"to"`
	Gas         *hexutil.Uint64        `json:"gas"`
	GasCurrency Smbol                  `json:"gasCy"` //default SERO
	GasPrice    *hexutil.Big           `json:"gasPrice"`
	Value       *hexutil.Big           `json:"value"`
	Data        *hexutil.Bytes         `json:"data"`
	Currency    Smbol                  `json:"cy"`
	Dynamic     bool                   `json:"dy"` //contract address parameters are dynamically generated.
	Category    Smbol                  `json:"catg"`
	Tkt         *common.Hash           `json:"tkt"`
	Memo        string                 `json:"Memo"`
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 90000
	}

	if args.GasCurrency.IsEmpty() {
		args.GasCurrency = Smbol(params.DefaultCurrency)
	}

	if strings.Trim(args.Memo, "") != "" {
		b := []byte(args.Memo)
		if len(b) > 64 {
			return errors.New("args memo is too long,it's limited 64 bytes")
		}
	}

	state, _, err := b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return err
	}

	if args.To != nil && !state.IsContract(common.BytesToAddress(args.To[:])) {
		var input []byte
		if args.Data != nil {
			input = *args.Data
		}

		if len(input) > 0 {
			return errors.New(`not create or call contract data must be nil`)
		}
		if args.To.IsAccountAddress() {
			if !superzk.IsPKValid(args.To.ToUint512()) {
				return errors.New("invalid to address")
			}
		}
	}

	if args.To == nil || !state.IsContract(common.BytesToAddress(args.To[:])) {
		if args.GasCurrency.IsNotEmpty() && args.GasCurrency.IsNotSero() {
			return errors.New(`GasCurrency must be null or SERO`)
		}
	} else {
		if args.GasCurrency.IsNotSero() {
			m, d := state.GetTokenRate(common.BytesToAddress(args.To[:]), string(args.GasCurrency))
			if m.Sign() == 0 || d.Sign() == 0 {
				return errors.New("the smart contract dose not support alternative payment!")
			}
		}
	}

	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	} else {
		if args.GasPrice.ToInt().Sign() == 0 {
			return errors.New(`gasPrice can not be zero`)
		}
	}

	if args.Currency.IsEmpty() {
		args.Currency = Smbol(params.DefaultCurrency)
	}

	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Category.IsEmpty() {
		if args.Tkt != nil {
			return errors.New(fmt.Sprintf("tx without tkt:%s catg", args.Tkt))
		}
	} else {
		if args.Tkt == nil {
			return errors.New(fmt.Sprintf("tx without %s tkt", args.Category))
		}
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		}

		if len(input) < 18 {
			return errors.New(`contract creation without any data provided`)
		}
	}
	return nil
}

func (args *SendTxArgs) toAsset() assets.Asset {
	var token *assets.Token
	var ticket *assets.Ticket
	if args.Value.ToInt().Sign() > 0 {
		token = &assets.Token{
			Currency: *(common.BytesToHash(common.LeftPadBytes([]byte(args.Currency), 32)).HashToUint256()),
			Value:    *utils.U256(*args.Value.ToInt()).ToRef(),
		}
	}
	if args.Tkt != nil {
		ticket = &assets.Ticket{
			Category: *(common.BytesToHash(common.LeftPadBytes([]byte(args.Category), 32)).HashToUint256()),
			Value:    *args.Tkt.HashToUint256(),
		}
	}
	return assets.Asset{
		Tkn: token,
		Tkt: ticket,
	}
}

func (args *SendTxArgs) toTxParam(state *state.StateDB) (txParam prepare.PreTxParam) {

	account, _ := Backend_Instance.AccountManager().FindByKey(&args.From)

	var refundPkr c_type.PKr
	txParam.GasPrice = (*big.Int)(args.GasPrice)
	txParam.From = *args.From.ToUint512()

	feevalue := new(big.Int).Mul(((*big.Int)(args.GasPrice)), new(big.Int).SetUint64(uint64(*args.Gas)))
	asset := args.toAsset()
	if args.To == nil {
		from := account.Accounts()[0].GetPKByHeight(Backend_Instance.CurrentBlock().NumberU64())
		fromRand := c_type.Uint256{}
		copy(fromRand[:16], (*args.Data)[:16])
		txParam.Cmds = prepare.Cmds{}
		contractCmd := stx.ContractCmd{asset, nil, *args.Data}
		txParam.Cmds.Contract = &contractCmd
		refundPkr = superzk.Pk2PKr(from.ToUint512(), &fromRand)
	} else if state.IsContract(common.BytesToAddress(args.To[:])) {
		from := account.Accounts()[0].GetPKByHeight(Backend_Instance.CurrentBlock().NumberU64())
		fromRand := c_type.Uint256{}
		copy(fromRand[:16], args.To[:16])
		if args.Dynamic {
			refundPkr = superzk.Pk2PKr(from.ToUint512(), nil)
		} else {
			refundPkr = superzk.Pk2PKr(from.ToUint512(), &fromRand)
		}
		if args.GasCurrency.IsNotSero() {
			m, d := state.GetTokenRate(common.BytesToAddress(args.To[:]), string(args.GasCurrency))
			feevalue = new(big.Int).Div(feevalue.Mul(feevalue, m), d)
		}
		txParam.Cmds = prepare.Cmds{}
		var data []byte
		if args.Data != nil {
			data = *args.Data
		}
		contractCmd := stx.ContractCmd{asset, args.To.ToPKr(), data}
		txParam.Cmds.Contract = &contractCmd
		refundPkr = superzk.Pk2PKr(from.ToUint512(), &fromRand)
	} else {
		var from address.AccountAddress
		var topkr c_type.PKr
		if args.To.IsAccountAddress() {
			topkr = superzk.Pk2PKr(args.To.ToUint512(), nil)
			from = account.Accounts()[0].GetPKByPK(args.To)
		} else {
			topkr = *args.To.ToPKr()
			from = account.Accounts()[0].GetPKByPKr(args.To)
		}
		refundPkr = superzk.Pk2PKr(from.ToUint512(), nil)
		receptions := []prepare.Reception{{Addr: topkr, Asset: asset}}
		txParam.Receptions = receptions
	}
	feeAsset := assets.Token{
		utils.CurrencyToUint256(string(args.GasCurrency)),
		utils.U256(*feevalue),
	}
	txParam.Fee = feeAsset
	txParam.RefundTo = &refundPkr
	return

}

func defaultFee(gasPrice *hexutil.Big, gas *hexutil.Uint64) *big.Int {
	return new(big.Int).Mul(((*big.Int)(gasPrice)), new(big.Int).SetUint64(uint64(*gas)))
}

func (args *SendTxArgs) toTransaction(state *state.StateDB) (*types.Transaction, *ztx.T) {
	var input []byte
	var Pkr c_type.PKr
	var isZ bool
	to := args.To
	fromRand := c_type.Uint256{}
	feevalue := defaultFee(args.GasPrice, args.Gas)
	if to == nil {
		copy(fromRand[:16], (*args.Data)[:16])
		isZ = false
	} else if state.IsContract(common.BytesToAddress(args.To[:])) {
		isZ = false
		Pkr = *to.ToPKr()
		if !args.Dynamic {
			copy(fromRand[:16], args.To[:16])
		}
		if args.GasCurrency.IsNotSero() {
			m, d := state.GetTokenRate(common.BytesToAddress(args.To[:]), string(args.GasCurrency))
			feevalue = new(big.Int).Div(feevalue.Mul(feevalue, m), d)
		}
	} else {
		Pkr = common.AddrToPKr(*args.To)
		fromRand = c_type.RandUint256()
		isZ = true
	}
	if args.Data != nil {
		input = *args.Data
	}
	tx := types.NewTransaction((*big.Int)(args.GasPrice), uint64(*args.Gas), input)
	ehash := tx.Ehash()
	fee := assets.Token{
		utils.CurrencyToUint256(string(args.GasCurrency)),
		utils.U256(*feevalue),
	}
	outData := types.NewTxtOut(Pkr, string(args.Currency), (*big.Int)(args.Value), string(args.Category), args.Tkt, args.Memo, isZ)
	txt := types.NewTxt(fromRand.NewRef(), ehash, fee, outData, nil, nil, nil)
	return tx, txt
}

func (args *SendTxArgs) toPkg(state *state.StateDB) (*types.Transaction, *ztx.T) {
	var Pkr c_type.PKr
	if state.IsContract(common.BytesToAddress(args.To[:])) {
		Pkr = *(args.To.ToPKr())
	} else {
		Pkr = common.AddrToPKr(*args.To)
	}
	tx := types.NewTransaction((*big.Int)(args.GasPrice), uint64(*args.Gas), nil)
	fromRand := c_type.RandUint256().NewRef()
	ehash := tx.Ehash()
	fee := assets.Token{
		utils.CurrencyToUint256(string(args.GasCurrency)),
		utils.U256(*new(big.Int).Mul(((*big.Int)(args.GasPrice)), new(big.Int).SetUint64(uint64(*args.Gas)))),
	}
	pkgCreate := types.NewCreatePkg(Pkr, string(args.Currency), (*big.Int)(args.Value), string(args.Category), args.Tkt, args.Memo)
	txt := types.NewTxt(fromRand, ehash, fee, nil, pkgCreate, nil, nil)
	txt.FromRnd = c_type.RandUint256().NewRef()
	return tx, txt
}

func stringToUint512(str string) c_type.Uint512 {
	var ret c_type.Uint512
	b := []byte(str)
	if len(b) > len(ret) {
		b = b[len(b)-len(ret):]
	}
	copy(ret[len(ret)-len(b):], b)
	return ret
}

func (args *SendTxArgs) toCreatePkg(state *state.StateDB) (txParam prepare.PreTxParam) {
	var toPkr c_type.PKr
	txParam.GasPrice = (*big.Int)(args.GasPrice)
	txParam.From = *args.From.ToUint512()
	feevalue := defaultFee(args.GasPrice, args.Gas)
	asset := args.toAsset()
	if state.IsContract(common.BytesToAddress(args.To[:])) {
		toPkr = *(args.To.ToPKr())
	} else {
		toPkr = common.AddrToPKr(*args.To)
	}
	feeToken := assets.Token{
		utils.CurrencyToUint256(string(args.GasCurrency)),
		utils.U256(*feevalue),
	}
	txParam.From = *args.From.ToUint512()
	txParam.RefundTo = superzk.Pk2PKr(args.From.ToUint512(), c_type.RandUint256().NewRef()).NewRef()
	txParam.Fee = feeToken
	pkgCreateCmd := prepare.PkgCreateCmd{c_type.RandUint256(), toPkr, asset, stringToUint512(args.Memo)}
	txParam.Cmds.PkgCreate = &pkgCreateCmd
	return

}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *types.Transaction, to *common.Address) (common.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	log.Info("Submitted transaction", "fullhash", tx.Hash().Hex(), "recipient", to)
	return tx.Hash(), nil
}

// SendTransaction creates a transaction for the given argument, sign it and submit it to the
// transaction pool.
func (s *PublicTransactionPoolAPI) SendTransaction(ctx context.Context, args SendTxArgs) (common.Hash, error) {
	s.nonceLock.mu.Lock()
	defer s.nonceLock.mu.Unlock()
	// Look up the wallet containing the requested abi

	if th, ok := s.b.GetEngin().(threaded); ok {
		miner := s.b.GetMiner()
		if miner.CanStart() {
			miner.SetCanStart(0)
			defer miner.SetCanStart(1)
		}
		threads := th.Threads()
		if threads >= 0 {
			th.SetThreads(-1)
			defer th.SetThreads(threads)
		}
	}
	return commitSendTxArgs(ctx, s.b, args)

}

func commitSendTxArgs(ctx context.Context, b Backend, args SendTxArgs) (common.Hash, error) {

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, b); err != nil {
		return common.Hash{}, err
	}

	state, _, err := b.StateAndHeaderByNumber(ctx, -1)

	if err != nil {
		return common.Hash{}, err
	}

	if !seroparam.IsExchange() {

		account := accounts.Account{Address: args.From}

		wallet, err := b.AccountManager().Find(account)
		if err != nil {
			return common.Hash{}, err
		}
		// Assemble the transaction and sign with the wallet
		tx, txt := args.toTransaction(state)
		encrypted, err := wallet.EncryptTx(account, tx, txt, state)
		if err != nil {
			return common.Hash{}, err
		}
		return submitTransaction(ctx, b, encrypted, args.To)
	} else {
		txParam := args.toTxParam(state)
		txhash, err := commitPreTx(txParam, b, args.To)
		if err != nil {
			return common.Hash{}, err
		} else {
			return txhash, nil
		}
	}
}

func commitPreTx(txParam prepare.PreTxParam, b Backend, to *common.Address) (common.Hash, error) {
	pretx, gtx, err := exchange.CurrentExchange().GenTxWithSign(txParam)
	if err != nil {
		return common.Hash{}, err
	}
	err = b.CommitTx(gtx)
	if err != nil {
		exchange.CurrentExchange().ClearTxParam(pretx)
		return common.Hash{}, err
	}
	txhash := common.BytesToHash(gtx.Hash[:])
	if to == nil {
		log.Info("create contract  transaction", "fullhash", txhash.Hex())
	} else {
		log.Info("Submitted transaction", "fullhash", txhash.Hex(), "recipient", to)
	}
	return txhash, nil

}

func (s *PublicTransactionPoolAPI) CommitTx(ctx context.Context, args *txtool.GTx) error {
	return s.b.CommitTx(args)
}

func (s *PublicTransactionPoolAPI) ReSendTransaction(ctx context.Context, txhash common.Hash) (common.Hash, error) {

	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return common.Hash{}, err
	}
	var tx *types.Transaction

	for _, ptx := range pending {
		if ptx.Hash() == txhash {
			tx = ptx
			break
		}
	}
	if tx == nil {
		return common.Hash{}, errors.New("can not find tx " + txhash.Hex() + " in local txpool!")
	}
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, tx, nil)
}

/**
func (s *PublicTransactionPoolAPI) CreatePkg(ctx context.Context, args SendTxArgs) (common.Hash, error) {
	s.nonceLock.mu.Lock()
	defer s.nonceLock.mu.Unlock()
	// Look up the wallet containing the requested abi
	account := accounts.Account{Address: args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	if args.To == nil {
		return common.Hash{}, errors.New("to can not be nil")
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}

	if args.GasCurrency.IsNotSero() {
		return common.Hash{}, errors.New("create pkg gasCurrency must be sero")
	}

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)

	if err != nil {
		return common.Hash{}, err
	}

	if seroparam.IsExchange() {
		txParam := args.toCreatePkg(state)
		pretx, gtx, err := exchange.CurrentExchange().GenTxWithSign(txParam)
		if err != nil {
			return common.Hash{}, err
		}
		err = s.b.CommitTx(gtx)
		if err != nil {
			exchange.CurrentExchange().ClearTxParam(pretx)
			return common.Hash{}, err
		}
		return common.BytesToHash(gtx.Hash[:]), nil
	} else {
		// Assemble the transaction and sign with the wallet
		tx, txt := args.toPkg(state)
		encrypted, err := wallet.EncryptTx(account, tx, txt, state)
		if err != nil {
			return common.Hash{}, err
		}
		return submitTransaction(ctx, s.b, encrypted, args.To)
	}

}
**/
type ClosePkgArgs struct {
	From     *address.AccountAddress `json:"from"`
	Gas      *hexutil.Uint64         `json:"gas"`
	GasPrice *hexutil.Big            `json:"gasPrice"`
	PkgId    *c_type.Uint256         `json:"id"`
	Key      *c_type.Uint256         `json:"key"`
}

func (args *ClosePkgArgs) setDefaults(ctx context.Context, b Backend) error {

	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 90000
	}

	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	} else {
		if args.GasPrice.ToInt().Sign() == 0 {
			return errors.New(`gasPrice can not be zero`)
		}
	}
	if args.PkgId == nil {
		return errors.New("id can not be nil")
	}

	if args.Key == nil {
		return errors.New("key can not be nil")
	}

	return nil
}

func (args *ClosePkgArgs) toTxParam() (txParam prepare.PreTxParam) {
	txParam.GasPrice = (*big.Int)(args.GasPrice)
	txParam.From = *args.From.ToUint512()
	feevalue := defaultFee(args.GasPrice, args.Gas)
	feeToken := assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*feevalue),
	}
	txParam.From = *args.From.ToUint512()
	txParam.RefundTo = superzk.Pk2PKr(args.From.ToUint512(), c_type.RandUint256().NewRef()).NewRef()
	txParam.Fee = feeToken
	pkgCloseCmd := prepare.PkgCloseCmd{*args.PkgId, *args.Key}
	txParam.Cmds.PkgClose = &pkgCloseCmd
	return

}

func (args *ClosePkgArgs) toTransaction(state *state.StateDB) (*types.Transaction, *ztx.T, error) {
	tx := types.NewTransaction((*big.Int)(args.GasPrice), uint64(*args.Gas), nil)
	fee := new(big.Int).Mul(((*big.Int)(args.GasPrice)), new(big.Int).SetUint64(uint64(*args.Gas)))
	ehash := tx.Ehash()
	txt := &ztx.T{
		Fee: assets.Token{
			utils.CurrencyToUint256(params.DefaultCurrency),
			utils.U256(*fee),
		},
		PkgClose: &ztx.PkgClose{*args.PkgId, *args.Key},
	}
	txt.Ehash = ehash
	txt.FromRnd = c_type.RandUint256().NewRef()
	return tx, txt, nil
}

func (s *PublicTransactionPoolAPI) ClosePkg(ctx context.Context, args ClosePkgArgs) (common.Hash, error) {
	s.nonceLock.mu.Lock()
	defer s.nonceLock.mu.Unlock()
	// Look up the wallet containing the requested abi
	account := accounts.Account{Address: *args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)

	if err != nil {
		return common.Hash{}, err
	}

	// Assemble the transaction and sign with the wallet
	tx, txt, err := args.toTransaction(state)
	if err != nil {
		return common.Hash{}, err
	}
	if th, ok := s.b.GetEngin().(threaded); ok {
		miner := s.b.GetMiner()
		if miner.CanStart() {
			miner.SetCanStart(0)
			defer miner.SetCanStart(1)
		}
		threads := th.Threads()
		if threads >= 0 {
			th.SetThreads(-1)
			defer th.SetThreads(threads)
		}
	}
	encrypted, err := wallet.EncryptTx(account, tx, txt, state)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, encrypted, nil)
}

type TransferPkgArgs struct {
	From     *address.AccountAddress `json:"from"`
	Gas      *hexutil.Uint64         `json:"gas"`
	GasPrice *hexutil.Big            `json:"gasPrice"`
	PkgId    *c_type.Uint256         `json:"id"`
	To       *common.Address         `json:"To"`
}

func (args *TransferPkgArgs) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 90000
	}

	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	} else {
		if args.GasPrice.ToInt().Sign() == 0 {
			return errors.New(`gasPrice can not be zero`)
		}
	}
	if args.PkgId == nil {
		return errors.New("id can not be nil")
	}

	if args.To == nil {
		return errors.New("to can not be nil")
	}

	return nil
}

func (args *TransferPkgArgs) toTransaction(state *state.StateDB) (*types.Transaction, *ztx.T, error) {
	tx := types.NewTransaction((*big.Int)(args.GasPrice), uint64(*args.Gas), nil)
	fee := new(big.Int).Mul(((*big.Int)(args.GasPrice)), new(big.Int).SetUint64(uint64(*args.Gas)))
	ehash := tx.Ehash()
	var Pkr c_type.PKr
	if state.IsContract(common.BytesToAddress(args.To[:])) {
		Pkr = *(args.To.ToPKr())
	} else {
		Pkr = common.AddrToPKr(*args.To)
	}
	txt := &ztx.T{
		Fee: assets.Token{
			utils.CurrencyToUint256(params.DefaultCurrency),
			utils.U256(*fee),
		},
		PkgTransfer: &ztx.PkgTransfer{*args.PkgId, Pkr},
	}
	txt.Ehash = ehash
	txt.FromRnd = c_type.RandUint256().NewRef()
	return tx, txt, nil
}

func (s *PublicTransactionPoolAPI) TransferPkg(ctx context.Context, args TransferPkgArgs) (common.Hash, error) {
	s.nonceLock.mu.Lock()
	defer s.nonceLock.mu.Unlock()
	// Look up the wallet containing the requested abi
	account := accounts.Account{Address: *args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}

	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)

	if err != nil {
		return common.Hash{}, err
	}

	// Assemble the transaction and sign with the wallet
	tx, txt, err := args.toTransaction(state)
	if err != nil {
		return common.Hash{}, err
	}
	if th, ok := s.b.GetEngin().(threaded); ok {
		miner := s.b.GetMiner()
		if miner.CanStart() {
			miner.SetCanStart(0)
			defer miner.SetCanStart(1)
		}
		threads := th.Threads()
		if threads >= 0 {
			th.SetThreads(-1)
			defer th.SetThreads(threads)
		}
	}
	encrypted, err := wallet.EncryptTx(account, tx, txt, state)
	if err != nil {
		return common.Hash{}, err
	}
	return submitTransaction(ctx, s.b, encrypted, args.To)
}

// EncryptTransactionResult represents a RLP encoded signed transaction.
type EncryptTransactionResult struct {
	Raw hexutil.Bytes      `json:"raw"`
	Tx  *types.Transaction `json:"tx"`
}

// EncryptTransaction will sign the given transaction with the from account.
// The node needs to have the private key of the account corresponding with
// the given from address and it needs to be unlocked.
func (s *PublicTransactionPoolAPI) EncryptTransaction(ctx context.Context, args SendTxArgs) (*EncryptTransactionResult, error) {
	s.nonceLock.mu.Lock()
	defer s.nonceLock.mu.Unlock()
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)

	if err != nil {
		return nil, err
	}

	if th, ok := s.b.GetEngin().(threaded); ok {
		miner := s.b.GetMiner()
		if miner.CanStart() {
			miner.SetCanStart(0)
			defer miner.SetCanStart(1)
		}
		threads := th.Threads()
		if threads >= 0 {
			th.SetThreads(-1)
			defer th.SetThreads(threads)
		}
	}
	var signed *types.Transaction

	if !seroparam.IsExchange() {
		// Assemble the transaction and sign with the wallet
		tx, txt := args.toTransaction(state)

		account := accounts.Account{Address: args.From}

		wallet, err := s.b.AccountManager().Find(account)
		if err != nil {
			return nil, err
		}
		signed, err = wallet.EncryptTx(account, tx, txt, state)
		if err != nil {
			return nil, err
		}
	} else {
		_, gtx, err := exchange.CurrentExchange().GenTxWithSign(args.toTxParam(state))
		if err != nil {
			return nil, err
		}
		gasPrice := big.Int(gtx.GasPrice)
		gas := uint64(gtx.Gas)
		signed = types.NewTxWithGTx(gas, &gasPrice, &gtx.Tx)
	}

	data, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return nil, err
	}
	return &EncryptTransactionResult{data, signed}, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return nil, err
	}
	transactions := make([]*RPCTransaction, 0, len(pending))
	for _, tx := range pending {
		if fromAddr := getLocalAccountAddressByPkr(s.b.AccountManager().Wallets(), tx.From()); fromAddr != nil {
			transactions = append(transactions, newRPCPendingTransaction(tx))
		}
	}
	return transactions, nil
}

// PublicDebugAPI is the collection of Ethereum APIs exposed over the public
// debugging endpoint.
type PublicDebugAPI struct {
	b Backend
}

// NewPublicDebugAPI creates a new API definition for the public debug methods
// of the Ethereum service.
func NewPublicDebugAPI(b Backend) *PublicDebugAPI {
	return &PublicDebugAPI{b: b}
}

// GetBlockRlp retrieves the RLP encoded for of a single block.
func (api *PublicDebugAPI) GetBlockRlp(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	encoded, err := rlp.EncodeToBytes(block)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", encoded), nil
}

// PrintBlock retrieves a block and returns its pretty printed form.
func (api *PublicDebugAPI) PrintBlock(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return spew.Sdump(block), nil
}

// SeedHash retrieves the seed hash of a block.
func (api *PublicDebugAPI) SeedHash(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return fmt.Sprintf("0x%x", ethash.SeedHash(number)), nil
}

// PrivateDebugAPI is the collection of Ethereum APIs exposed over the private
// debugging endpoint.
type PrivateDebugAPI struct {
	b Backend
}

// NewPrivateDebugAPI creates a new API definition for the private debug methods
// of the Ethereum service.
func NewPrivateDebugAPI(b Backend) *PrivateDebugAPI {
	return &PrivateDebugAPI{b: b}
}

// ChaindbProperty returns leveldb properties of the chain database.
func (api *PrivateDebugAPI) ChaindbProperty(property string) (string, error) {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return "", fmt.Errorf("chaindbProperty does not work for memory databases")
	}
	if property == "" {
		property = "leveldb.stats"
	} else if !strings.HasPrefix(property, "leveldb.") {
		property = "leveldb." + property
	}
	return ldb.LDB().GetProperty(property)
}

func (api *PrivateDebugAPI) ChaindbCompact() error {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return fmt.Errorf("chaindbCompact does not work for memory databases")
	}
	for b := byte(0); b < 255; b++ {
		log.Info("Compacting chain database", "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
		err := ldb.LDB().CompactRange(util.Range{Start: []byte{b}, Limit: []byte{b + 1}})
		if err != nil {
			log.Error("Database compaction failed", "err", err)
			return err
		}
	}
	return nil
}

// SetHead rewinds the head of the blockchain to a previous block.
func (api *PrivateDebugAPI) SetHead(number hexutil.Uint64) {
	api.b.SetHead(uint64(number))
}

// PublicNetAPI offers network related RPC methods
type PublicNetAPI struct {
	net            *p2p.Server
	networkVersion uint64
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(net *p2p.Server, networkVersion uint64) *PublicNetAPI {
	return &PublicNetAPI{net, networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
func (s *PublicNetAPI) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (s *PublicNetAPI) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.net.PeerCount())
}

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}
