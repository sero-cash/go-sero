// Copyright 2018 The go-ethereum Authors
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
	"runtime"

	"github.com/sero-cash/go-sero/zero/txtool/verify"

	"github.com/sero-cash/go-sero/core/types"
)

type CheckDesc struct {
	tx            *types.Transaction
	block         *types.Block
	hasReceptions bool
}

func NewTxChecker(bc *BlockChain, chain types.Blocks) (chan<- struct{}, <-chan error) {
	txs := []CheckDesc{}
	for _, block := range chain {
		rpts := bc.GetReceiptsByHash(block.Hash())
		for _, tx := range block.Transactions() {
			cd := CheckDesc{}
			cd.tx = tx
			cd.block = block
			if len(rpts) > 0 {
				cd.hasReceptions = true
			}
			txs = append(txs, cd)
		}
	}

	if len(txs) == 0 {
		return make(chan struct{}), nil
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(txs) < workers {
		workers = len(txs)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]error, len(txs))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				tx := txs[index]
				if tx.hasReceptions {
					errors[index] = nil
				} else {
					errors[index] = verify.VerifyWithoutState(tx.tx.Ehash().NewRef(), tx.tx.GetZZSTX(), tx.block.NumberU64())
				}
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(txs))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(txs))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(txs) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(txs)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}
