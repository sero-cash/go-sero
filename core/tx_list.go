// Copyright 2016 The go-ethereum Authors
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
	"container/heap"
	"math/big"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/log"
)

// priceHeap is a heap.Interface implementation over transactions for retrieving
// priced-sorted transactions to discard when the pool fills up.
type priceHeap []*types.Transaction

func (h priceHeap) Len() int { return len(h) }
func (h priceHeap) Swap(i, j int) {
	if i < 0 || j < 0 {
		return
	}
	h[i], h[j] = h[j], h[i]
}

func (h priceHeap) Less(i, j int) bool {
	// Sort primarily by priced, returning the cheaper one
	switch h[i].GasPrice().Cmp(h[j].GasPrice()) {
	case -1:
		return true
	case 1:
		return false
	}
	// If the prices match, stabilize via nonces (high nonce is worse)
	return false
}

func (h *priceHeap) Push(x interface{}) {
	*h = append(*h, x.(*types.Transaction))
}

func (h *priceHeap) Pop() interface{} {
	if h.Len() < 1 {
		return nil
	}
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// txPricedList is a priced-sorted heap to allow operating on transactions pool
// contents in a priced-incrementing way.
type txPricedList struct {
	all    *txLookup  // Pointer to the map of all transactions
	items  *priceHeap // Heap of prices of all the stored transactions
	stales int        // Number of stale priced points to (re-heap trigger)
}

// newTxPricedList creates a new priced-sorted transaction heap.
func newTxPricedList(all *txLookup) *txPricedList {
	return &txPricedList{
		all:   all,
		items: new(priceHeap),
	}
}

func (l *txPricedList) Get(hash common.Hash) *types.Transaction {
	return l.all.Get(hash)
}

func (l *txPricedList) Add(tx *types.Transaction, threshold *big.Int) bool {
	if tx.GasPrice().Cmp(threshold) < 0 {
		return false
	}
	if t := l.all.Get(tx.Hash()); t == nil {
		heap.Push(l.items, tx)
	}
	l.all.Add(tx)
	return true
}

func (l *txPricedList) Flatten() types.Transactions {
	txs := types.Transactions{}
	for _, i := range *l.items {
		txs = append(txs, i)
	}
	return txs
}

func (l *txPricedList) Ready() types.Transactions {
	// Otherwise start accumulating incremental transactions
	var ready types.Transactions

	if l.items == nil {
		return ready
	}
	for t := heap.Pop(l.items); t != nil; t = heap.Pop(l.items) {
		tx := t.((*types.Transaction))
		ready = append(ready, tx)
		l.all.Remove(tx.Hash())
	}

	return ready
}

func (l *txPricedList) Len() int {
	return l.items.Len()
}

func (l *txPricedList) Remove(tx *types.Transaction) bool {
	// Remove the transaction from the set
	if l.all.Get(tx.Hash()) == nil {
		return false
	}

	l.all.Remove(tx.Hash())
	for i, item := range *l.items {
		if item.Hash() == tx.Hash() {
			heap.Remove(l.items, i)
			return true
		}
	}
	return false
}

// Underpriced checks whether a transaction is cheaper than (or as cheap as) the
// lowest priced transaction currently being tracked.
func (l *txPricedList) Underpriced(tx *types.Transaction, local *accountSet) bool {
	// Local transactions cannot be underpriced
	if local.containsTx(tx) {
		return false
	}
	// Discard stale priced points if found at the heap start
	for len(*l.items) > 0 {
		head := []*types.Transaction(*l.items)[0]
		if l.all.Get(head.Hash()) == nil {
			l.stales--
			heap.Pop(l.items)
			continue
		}
		break
	}
	// Check if the transaction is underpriced or not
	if len(*l.items) == 0 {
		log.Error("Pricing query for empty pool") // This cannot happen, print to catch programming errors
		return false
	}
	cheapest := []*types.Transaction(*l.items)[0]
	return cheapest.GasPrice().Cmp(tx.GasPrice()) >= 0
}

// Discard finds a number of most underpriced transactions, removes them from the
// priced list and returns them for further removal from the entire pool.
func (l *txPricedList) Discard(threshold *big.Int, count int) types.Transactions {
	drop := make(types.Transactions, 0, count) // Remote underpriced transactions to drop
	save := make(types.Transactions, 0, 64)    // Local underpriced transactions to keep

	save_count := len(*l.items) - count
	for len(*l.items) > 0 {
		// Discard stale transactions if found during cleanup
		tx := heap.Pop(l.items).(*types.Transaction)
		if l.all.Get(tx.Hash()) == nil {
			l.stales--
			continue
		}
		l.all.Remove(tx.Hash())
		// Non stale transaction found, discard unless local
		if save_count > 0 && tx.GasPrice().Cmp(threshold) >= 0 {
			save = append(save, tx)
			save_count--
		} else {
			drop = append(drop, tx)
		}
	}
	for _, tx := range save {
		l.Add(tx, new(big.Int))
	}
	return drop
}
