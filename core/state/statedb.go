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

// Package state provides a caching layer atop the Ethereum state trie.
package state

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"strings"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/trie"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	// emptyState is the known hash of an empty state trie entry.
	emptyState = crypto.Keccak256Hash(nil)

	// emptyCode is the known hash of the empty EVM bytecode.
	emptyCode = crypto.Keccak256Hash(nil)

	EmptyAddress = common.BytesToAddress(crypto.Keccak512(nil))

	currency_value = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
)

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	root common.Hash
	db   Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}
	zstate            *zstate.State

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash common.Hash
	txIndex      int
	logs         map[common.Hash][]*types.Log
	logSize      uint

	preimages map[common.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	seeds  []keys.Uint512
	number uint64
	lock   sync.Mutex
}

func (self *StateDB) IsContract(addr common.Address) bool {
	return self.GetCode(addr) != nil
}

func (self *StateDB) GetSeeds() []keys.Uint512 {
	return self.seeds
}

func (self *StateDB) SetSeeds(seeds []keys.Uint512) {
	self.seeds = seeds
}

type StateDbGet interface {
	Get(key []byte) ([]byte, error)
}
type StateDbPut interface {
	Put(key, value []byte) error
}
type StateTri struct {
	Tri   Trie
	Dbget StateDbGet
	Dbput StateDbPut
}

func (self *StateTri) TryGet(key []byte) ([]byte, error) {
	return self.Tri.TryGet(key)
}

func (self *StateTri) TryUpdate(key, value []byte) error {
	return self.Tri.TryUpdate(key, value)
}

func (self *StateTri) TryGlobalGet(key []byte) ([]byte, error) {
	return self.Dbget.Get(key)
}

func (self *StateTri) TryGlobalPut(key, value []byte) error {
	return self.Dbput.Put(key, value)
}

func (self *StateDB) GetZState() *zstate.State {

	if self.zstate == nil {
		st := StateTri{self.trie, self.db.TrieDB().DiskDB(), self.db.TrieDB().WDiskDB()}
		self.zstate = zstate.NewState(&st, self.number)
	}
	return self.zstate
}

func NewGenesis(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)

	if err != nil {
		return nil, err
	}

	return &StateDB{
		root:              root,
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		logs:              make(map[common.Hash][]*types.Log),
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}, nil
}

// Create a new state from a given trie.
func New(root common.Hash, db Database, number uint64) (*StateDB, error) {
	tr, err := db.OpenTrie(root)

	if err != nil {
		return nil, err
	}

	return &StateDB{
		root:              root,
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		logs:              make(map[common.Hash][]*types.Log),
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
		number:            number + 1,
	}, nil
}

func (self *StateDB) GetAvgUsedGas(number uint64) *big.Int {
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		value := stateObject.GetState(self.db, self.usedGasKey(number))
		return big.NewInt(0).SetBytes(value.Bytes())
	}
	return new(big.Int)
}

func (self *StateDB) SetAvgUsedGas(number uint64, avgGsedGas *big.Int) {
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		stateObject.SetState(self.db, self.usedGasKey(number), common.BigToHash(avgGsedGas))
	}
}

func (self *StateDB) usedGasKey(number uint64) common.Hash {
	bytes, _ := rlp.EncodeToBytes([]interface{}{"usedGas", number})
	key := crypto.Keccak256Hash(bytes)
	return key
}

//register
func (self *StateDB) RegisterCurrency(contractAddr common.Address, coinName string) bool {
	coinName = strings.ToLower(coinName)
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		key := crypto.Keccak256Hash([]byte(coinName))
		value := stateObject.GetState(self.db, key)
		hash := crypto.Keccak256Hash(contractAddr.Bytes())
		if value == (common.Hash{}) {
			stateObject.SetState(self.db, key, hash)
			return true
		} else {
			return value == hash
		}
	}
	return false
}

func (self *StateDB) ExistsCurrency(coinName string) bool {
	coinName = strings.ToLower(coinName)
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		value := stateObject.GetState(self.db, crypto.Keccak256Hash([]byte(coinName)))
		if value != (common.Hash{}) {
			return true
		}
	}
	return false
}

func (self *StateDB) setCurrency(contractAddr common.Address, coinName string) common.Hash {
	coinName = strings.ToLower(coinName)
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		//bytes, _ := rlp.EncodeToBytes([]interface{}{contractAddr, coinName})
		//hash := crypto.Keccak256Hash(bytes)
		stateObject.SetState(self.db, crypto.Keccak256Hash([]byte(coinName)), currency_value)
		return common.BytesToHash(common.LeftPadBytes([]byte(coinName), 32))
	}
	return common.Hash{}
}

func (self *StateDB) getCurrency(coinName string) common.Hash {
	if strings.TrimSpace(coinName) == "" {
		return common.Hash{}
	}
	coinName = strings.ToLower(coinName)
	stateObject := self.getStateObject(EmptyAddress)
	if stateObject != nil {
		return stateObject.GetState(self.db, crypto.Keccak256Hash([]byte(coinName)))
	}
	return common.Hash{}
}

func (self *StateDB) AddNonceAddress(key []byte, nonceAddr common.Address) {
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		key0 := crypto.Keccak256Hash(append([]byte("nonceAddr0"), key[:]...))
		if stateObject.GetState(self.db, key0) != (common.Hash{}) {
			return
		}
		key1 := crypto.Keccak256Hash(append([]byte("nonceAddr1"), key[:]...))
		stateObject.SetState(self.db, key0, common.BytesToHash(nonceAddr.Bytes()[:32]))
		stateObject.SetState(self.db, key1, common.BytesToHash(nonceAddr.Bytes()[32:]))
	}
}

func (self *StateDB) GetNonceAddress(key []byte) common.Address {
	stateObject := self.GetOrNewStateObject(EmptyAddress)
	if stateObject != nil {
		key0 := crypto.Keccak256Hash(append([]byte("nonceAddr0"), key[:]...))
		key1 := crypto.Keccak256Hash(append([]byte("nonceAddr1"), key[:]...))
		h := stateObject.GetState(self.db, key0)
		if h == (common.Hash{}) {
			return common.Address{}
		}
		l := stateObject.GetState(self.db, key1)
		if l == (common.Hash{}) {
			return common.Address{}
		}
		return common.BytesToAddress(append(h[:], l[:]...))
	}
	return common.Address{}
}

// setError remembers the first non-nil error it is called with.
func (self *StateDB) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *StateDB) Error() error {
	return self.dbErr
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset(root common.Hash) error {
	tr, err := self.db.OpenTrie(root)
	if err != nil {
		return err
	}
	self.trie = tr
	self.stateObjects = make(map[common.Address]*stateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash][]*types.Log)
	self.logSize = 0
	self.preimages = make(map[common.Hash][]byte)
	self.clearJournalAndRefund()
	return nil
}

func (self *StateDB) AddLog(log *types.Log) {
	self.journal.append(addLogChange{txhash: self.thash})

	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return self.logs[hash]
}

func (self *StateDB) Logs() []*types.Log {
	var logs []*types.Log
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (self *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	if _, ok := self.preimages[hash]; !ok {
		self.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		self.preimages[hash] = pi
	}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (self *StateDB) Preimages() map[common.Hash][]byte {
	return self.preimages
}

func (self *StateDB) AddRefund(gas uint64) {
	self.journal.append(refundChange{prev: self.refund})
	self.refund += gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (self *StateDB) Exist(addr common.Address) bool {
	return self.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (self *StateDB) Empty(addr common.Address) bool {
	so := self.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address, coinName string) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance(coinName)
	}
	return common.Big0
}

func (self *StateDB) Balances(addr common.Address) map[string]*big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balances()
	}
	return map[string]*big.Int{}
}

func (self *StateDB) GetCode(addr common.Address) []byte {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code(self.db)
	}
	return nil
}

func (self *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	size, err := self.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		self.setError(err)
	}
	return size
}

func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

func (self *StateDB) GetState(addr common.Address, bhash common.Hash) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(self.db, bhash)
	}
	return common.Hash{}
}

// Database retrieves the low level database supporting the lower level trie ops.
func (self *StateDB) Database() Database {
	return self.db
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (self *StateDB) StorageTrie(addr common.Address) Trie {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(self)
	return cpy.updateTrie(self.db)
}

func (self *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (self *StateDB) AddBalance(addr common.Address, coinName string, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(coinName, amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (self *StateDB) SubBalance(addr common.Address, coinName string, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(coinName, amount)
	}
}

func (self *StateDB) SetBalance(addr common.Address, coinName string, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(coinName, amount)
	}
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (self *StateDB) SetState(addr common.Address, key, value common.Hash) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(self.db, key, value)
	}
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (self *StateDB) Suicide(addr common.Address, toAddr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	if self.IsContract(toAddr) {
		toStateObject := self.getStateObject(addr)
		for coinName, book := range stateObject.data.bookMap {
			self.journal.append(suicideChange{
				account:     &addr,
				prev:        stateObject.suicided,
				prevbalance: new(big.Int).Set(book.Balance),
			})
			toStateObject.AddBalance(coinName, book.Balance)
		}
	} else {
		for coinName, book := range stateObject.data.bookMap {
			self.journal.append(suicideChange{
				account:     &addr,
				prev:        stateObject.suicided,
				prevbalance: new(big.Int).Set(book.Balance),
			})
			currency := common.BytesToHash(common.LeftPadBytes([]byte(coinName), 32))
			self.GetZState().AddTxOut(toAddr, book.Balance, currency.HashToUint256())
		}
	}
	stateObject.markSuicided()
	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
func (self *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr.Bytes(), err))
	}
	self.setError(self.trie.TryUpdate(addr.Bytes(), data))
}

// deleteStateObject removes the given object from the state trie.
func (self *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	self.setError(self.trie.TryDelete(addr.Bytes()))
}

// Retrieve a state object given by the address. Returns nil if not found.
func (self *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	enc, err := self.trie.TryGet(addr.Bytes())
	if len(enc) == 0 {
		self.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *stateObject) {
	self.stateObjects[object.Address()] = object
}

// Retrieve a state object or create a new state object if nil.
func (self *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := self.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObject(addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (self *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = self.getStateObject(addr)
	newobj = newObject(self, addr, Account{})
	//newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		self.journal.append(createObjectChange{account: &addr})
	} else {
		self.journal.append(resetObjectChange{prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

// Carrying over the balance ensures that Ether doesn't disappear.
func (self *StateDB) CreateAccount(addr common.Address) {
	self.createObject(addr)

}

func (db *StateDB) ForEachOuts(cb func(key, value common.Hash) bool) {
	it := trie.NewIterator(db.trie.NodeIterator(nil))
	for it.Next() {
		// ignore cached values
		key := common.BytesToHash(db.trie.GetKey(it.Key))
		cb(key, common.BytesToHash(it.Value))
	}
}

func (db *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {
	so := db.getStateObject(addr)
	if so == nil {
		return
	}

	// When iterating over the storage check the cache first
	for h, value := range so.cachedStorage {
		cb(h, value)
	}

	it := trie.NewIterator(so.getTrie(db.db).NodeIterator(nil))
	for it.Next() {
		// ignore cached values
		key := common.BytesToHash(db.trie.GetKey(it.Key))
		if _, ok := so.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (self *StateDB) Copy() *StateDB {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		root:              self.root,
		db:                self.db,
		trie:              self.db.CopyTrie(self.trie),
		stateObjects:      make(map[common.Address]*stateObject, len(self.journal.dirties)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(self.journal.dirties)),
		zstate:            self.zstate.Copy(),
		refund:            self.refund,
		logs:              make(map[common.Hash][]*types.Log, len(self.logs)),
		logSize:           self.logSize,
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range self.journal.dirties {
		// As documented [here](https://github.com/sero-cash/go-sero/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to AutumnTwilight. Thus, we need to check for
		// nil
		if object, exist := self.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range self.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	for hash, logs := range self.logs {
		state.logs[hash] = make([]*types.Log, len(logs))
		copy(state.logs[hash], logs)
	}
	for hash, preimage := range self.preimages {
		state.preimages[hash] = preimage
	}
	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (self *StateDB) Snapshot() int {
	id := self.nextRevisionId
	self.nextRevisionId++
	self.validRevisions = append(self.validRevisions, revision{id, self.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (self *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(self.validRevisions), func(i int) bool {
		return self.validRevisions[i].id >= revid
	})
	if idx == len(self.validRevisions) || self.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := self.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	self.journal.revert(self, snapshot)
	self.validRevisions = self.validRevisions[:idx]

	self.GetZState().Revert()
}

// GetRefund returns the current value of the refund counter.
func (self *StateDB) GetRefund() uint64 {
	return self.refund
}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (s *StateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range s.journal.dirties {
		stateObject, exist := s.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}

		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			s.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(s.db)
			s.updateStateObject(stateObject)
		}
		s.stateObjectsDirty[addr] = struct{}{}
	}
	// Invalidate journal because reverting across transactions is not allowed.
	s.clearJournalAndRefund()
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	s.GetZState().Update()
	s.Finalise(deleteEmptyObjects)
	return s.trie.Hash()
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (self *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (s *StateDB) clearJournalAndRefund() {
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.refund = 0
}

// Commit writes the state to the underlying in-memory trie database.
func (s *StateDB) Commit(deleteEmptyObjects bool) (root common.Hash, err error) {
	defer s.clearJournalAndRefund()

	for addr := range s.journal.dirties {
		s.stateObjectsDirty[addr] = struct{}{}
	}
	start := time.Now()
	// Commit objects to the trie.
	for addr, stateObject := range s.stateObjects {
		_, isDirty := s.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.deleteStateObject(stateObject)
		case isDirty:
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				s.db.TrieDB().InsertBlob(common.BytesToHash(stateObject.CodeHash()), stateObject.code)
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.CommitTrie(s.db); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			s.updateStateObject(stateObject)
		}
		delete(s.stateObjectsDirty, addr)
	}

	log.Info(fmt.Sprintf("Commit objects to the trie done in %v,num=%v\n", time.Since(start), s.GetZState().Num()))

	// Write trie changes.
	root, err = s.trie.Commit(s.leafCallback)

	log.Debug("Trie cache stats after commit", "misses", trie.CacheMisses(), "unloads", trie.CacheUnloads())
	return root, err
}

func (s *StateDB) leafCallback(leaf []byte, parent common.Hash) error {
	var account Account
	if err := rlp.DecodeBytes(leaf, &account); err != nil {
		return nil
	}
	if account.Root != emptyState {
		s.db.TrieDB().Reference(account.Root, parent)
	}
	code := common.BytesToHash(account.CodeHash)
	if code != emptyCode {
		s.db.TrieDB().Reference(code, parent)
	}
	return nil
}
