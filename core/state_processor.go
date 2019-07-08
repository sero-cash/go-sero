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

package core

import (
	"errors"
	"math/big"

	"github.com/sero-cash/go-sero/zero/stake"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/consensus"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/core/vm"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/params"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase).
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		header   = block.Header()
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)

	// Iterate over and process the individual transactions
	gasReward := uint64(0)

	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		receipt, gas, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			return nil, nil, 0, err
		}
		gasReward += new(big.Int).Mul(new(big.Int).SetUint64(gas), tx.GasPrice()).Uint64()
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts, gasReward)

	return receipts, allLogs, *usedGas, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	msg, err := tx.AsMessage()
	if err != nil {
		return nil, 0, err
	}

	err = statedb.GetZState().AddStx(tx.GetZZSTX())
	if err != nil {
		return nil, 0, err
	}

	key := header.Coinbase.ToCaddr()
	statedb.AddNonceAddress(key[:], header.Coinbase)

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := ApplyMessage(vmenv, msg, gp)

	if err != nil {
		gp.AddGas(gas)
		return nil, 0, err
	}

	if err = applyStake(msg.From(), tx.GetZZSTX().Desc_Cmd, statedb, tx.Hash(), header.Number.Uint64()); err != nil {
		return nil, 0, err
	}

	root := statedb.IntermediateRoot(true).Bytes()
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil && len(msg.Data()) > 0 && !failed {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, statedb.GetContrctNonce(), msg.Data()[0:16])
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, err
}

var (
	poolValueThreshold, _ = new(big.Int).SetString("1000000000000000000", 10)
)

func applyStake(from common.Address, stakeDesc stx.DescCmd, statedb *state.StateDB, txHash common.Hash, number uint64) error {
	stakeState := stake.NewStakeState(statedb)
	pkr := *from.ToPKr()
	if stakeDesc.BuyShare != nil {
		var stakePool *stake.StakePool
		if stakeDesc.BuyShare.Pool != nil {
			poolId := common.BytesToHash(stakeDesc.BuyShare.Pool[:])
			stakePool = stakeState.GetStakePool(poolId)
			if stakePool == nil {
				return errors.New("not exist pool,id" + poolId.String())
			}
			if stakePool.Closed {
				return errors.New("pool is closed,id" + poolId.String())
			}
		}

		value := stakeDesc.BuyShare.Value.ToInt()
		num, avgPrice := stakeState.CaleAvgPrice(value)

		if num > 0 {
			amount := new(big.Int).Mul(avgPrice, big.NewInt(int64(num)))
			refund := new(big.Int).Sub(new(big.Int).Set(stakeDesc.BuyShare.Value.ToInt()), amount)
			if refund.Sign() > 0 {
				asset := assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
					Value:    utils.U256(*refund),
				},
				}
				statedb.GetZState().AddTxOut(from, asset)
			}

			share := &stake.Share{PKr: pkr, VotePKr: stakeDesc.BuyShare.Vote, Value: avgPrice, TransactionHash: txHash, BlockNumber: number, InitNum: num, Num: num, Profit: big.NewInt(0)}
			if stakePool != nil {
				hash := common.BytesToHash(stakePool.Id())
				share.PoolId = &hash
				share.Fee = stakePool.Fee
			}
			stakeState.UpdateShare(share)
		} else {
			asset := assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
				Value:    stakeDesc.BuyShare.Value,
			},
			}
			statedb.GetZState().AddTxOut(from, asset)
		}
	} else if stakeDesc.RegistPool != nil {
		if poolValueThreshold.Cmp(stakeDesc.RegistPool.Value.ToInt()) <= 0 {
			cmd := stakeDesc.RegistPool
			pool := &stake.StakePool{PKr: pkr, Amount: cmd.Value.ToInt(), VotePKr: cmd.Vote, TransactionHash: txHash, Fee: uint16(cmd.FeeRate), Profit: big.NewInt(0)}
			stakeState.UpdateStakePool(pool)
		} else {
			return errors.New("value < poolValueThreshold(1000000000000000000)")
		}
	} else if stakeDesc.ClosePool != nil {
		poolId := crypto.Keccak256Hash(pkr[:])
		stakePool := stakeState.GetStakePool(poolId)
		if stakePool == nil {
			return errors.New("not exist pool,id" + poolId.String())
		}
		stakePool.Closed = true
		stakeState.UpdateStakePool(stakePool)
	}
	return nil
}
