// Copyright 2017 The go-ethereum Authors
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

package ethash

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sero-cash/go-sero/crypto"
	"math/big"
	"runtime"
	"time"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/consensus"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/params"
)

// Ethash proof-of-work protocol constants.
var (
	allowedFutureBlockTime          = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
	bigOne                 *big.Int = big.NewInt(1)
	big200W                *big.Int = big.NewInt(2000000)
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errZeroBlockTime     = errors.New("timestamp equals parent's")
	errInvalidDifficulty = errors.New("non-positive difficulty")
	errInvalidMixDigest  = errors.New("invalid mix digest")
	errInvalidPoW        = errors.New("invalid proof-of-work")
)

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (ethash *Ethash) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
func (ethash *Ethash) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake {
		return nil
	}
	// Short circuit if the header is known, or it's parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Sanity checks passed, do a proper verification
	return ethash.verifyHeader(chain, header, parent, seal)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (ethash *Ethash) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	// If we're running a full engine faking, accept any input as valid
	if ethash.config.PowMode == ModeFullFake || len(headers) == 0 {
		abort, results := make(chan struct{}), make(chan error, len(headers))
		for i := 0; i < len(headers); i++ {
			results <- nil
		}
		return abort, results
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]error, len(headers))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = ethash.verifyHeaderWorker(chain, headers, seals, index)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
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

func (ethash *Ethash) verifyHeaderWorker(chain consensus.ChainReader, headers []*types.Header, seals []bool, index int) error {
	var parent *types.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if chain.GetHeader(headers[index].Hash(), headers[index].Number.Uint64()) != nil {
		return nil // known block
	}
	return ethash.verifyHeader(chain, headers[index], parent, seals[index])
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
// See YP section 4.3.4. "Block Header Validity"
func (ethash *Ethash) verifyHeader(chain consensus.ChainReader, header, parent *types.Header, seal bool) error {
	if !keys.CheckLICr(header.Coinbase.ToPKr(), &header.Licr, header.Number.Uint64()) {
		return fmt.Errorf("invalid Licr : pkr %v, licr %v", header.Coinbase, header.Licr)
	}

	if !header.Valid() {
		return fmt.Errorf("invalid Licr : pkr %v, licr %v, disable", header.Coinbase, header.Licr)
	}
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if header.Time.Cmp(big.NewInt(time.Now().Add(allowedFutureBlockTime).Unix())) > 0 {
		return consensus.ErrFutureBlock
	}

	if header.Time.Cmp(parent.Time) <= 0 {
		return errZeroBlockTime
	}
	// Verify the block's difficulty based in it's timestamp and parent's difficulty
	expected := ethash.CalcDifficulty(chain, header.Time.Uint64(), parent)

	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	divisor := uint64(1024)
	if diff < 0 {
		diff *= -1
		divisor = uint64(128)
	}
	limit := parent.GasLimit / divisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	// Verify the engine specific seal securing the block
	if seal {
		if err := ethash.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (ethash *Ethash) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return CalcDifficulty(chain.Config(), time, parent)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(config *params.ChainConfig, time uint64, parent *types.Header) *big.Int {
	return calcDifficultyAutumnTwilight(time, parent)
}

// Some weird constants to avoid constant memory allocs for them.
var (
	expDiffPeriod = big.NewInt(100000)
	big1          = big.NewInt(1)
	big2          = big.NewInt(2)
	big6          = big.NewInt(6)
	big9          = big.NewInt(9)
	bigMinus99    = big.NewInt(-99)
)

// calcDifficultyAutumnTwilight is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time given the
// parent block's time and difficulty. The calculation uses the AutumnTwilight rules.
func calcDifficultyAutumnTwilight(time uint64, parent *types.Header) *big.Int {
	// https://github.com/ethereum/EIPs/issues/100.
	// algorithm:
	// diff = (parent_diff +
	//         (parent_diff / 2048 * max((1 - ((timestamp - parent.timestamp) // 9), -99))
	//        )
	bigTime := new(big.Int).SetUint64(time)
	bigParentTime := new(big.Int).Set(parent.Time)

	// holds intermediate values to make the algo easier to read & audit
	x := new(big.Int)
	y := new(big.Int)

	// 1 - (block_timestamp - parent_timestamp) // 9
	x.Sub(bigTime, bigParentTime)
	x.Div(x, big9)
	x.Sub(big1, x)
	// max(1 - (block_timestamp - parent_timestamp) // 9, -99)
	if x.Cmp(bigMinus99) < 0 {
		x.Set(bigMinus99)
	}
	// parent_diff + (parent_diff / 2048 * max(1 - ((timestamp - parent.timestamp) // 9), -99))
	y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
	x.Mul(y, x)
	x.Add(parent.Difficulty, x)

	// minimum difficulty can ever be (before exponential factor)
	if x.Cmp(params.MinimumDifficulty) < 0 {
		x.Set(params.MinimumDifficulty)
	}
	return x
}

// calcDifficultyFrontier is the difficulty adjustment algorithm. It returns the
// difficulty that a new block should have when created at time given the parent
// block's time and difficulty. The calculation uses the Frontier rules.
func calcDifficultyFrontier(time uint64, parent *types.Header) *big.Int {
	diff := new(big.Int)
	adjust := new(big.Int).Div(parent.Difficulty, params.DifficultyBoundDivisor)
	bigTime := new(big.Int)
	bigParentTime := new(big.Int)

	bigTime.SetUint64(time)
	bigParentTime.Set(parent.Time)

	if bigTime.Sub(bigTime, bigParentTime).Cmp(params.DurationLimit) < 0 {
		diff.Add(parent.Difficulty, adjust)
	} else {
		diff.Sub(parent.Difficulty, adjust)
	}
	if diff.Cmp(params.MinimumDifficulty) < 0 {
		diff.Set(params.MinimumDifficulty)
	}

	periodCount := new(big.Int).Add(parent.Number, big1)
	periodCount.Div(periodCount, expDiffPeriod)
	if periodCount.Cmp(big1) > 0 {
		// diff = diff + 2^(periodCount - 2)
		expDiff := periodCount.Sub(periodCount, big2)
		expDiff.Exp(big2, expDiff, nil)
		diff.Add(diff, expDiff)
		diff = math.BigMax(diff, params.MinimumDifficulty)
	}
	return diff
}

// VerifySeal implements consensus.Engine, checking whether the given block satisfies
// the PoW difficulty requirements.
func (ethash *Ethash) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	// If we're running a fake PoW, accept any seal as valid
	if ethash.config.PowMode == ModeFake || ethash.config.PowMode == ModeFullFake {
		time.Sleep(ethash.fakeDelay)
		if ethash.fakeFail == header.Number.Uint64() {
			return errInvalidPoW
		}
		return nil
	}
	// If we're running a shared PoW, delegate verification to it
	if ethash.shared != nil {
		return ethash.shared.VerifySeal(chain, header)
	}
	// Ensure that we have a valid difficulty for the block
	if header.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}
	// Recompute the digest and PoW value and verify against the header
	number := header.Number.Uint64()

	cache := ethash.cache(number)
	size := datasetSize(number)
	if ethash.config.PowMode == ModeTest {
		size = 32 * 1024
	}
	digest, result := hashimotoLight(size, cache.cache, header.HashNoNonce().Bytes(), header.Nonce.Uint64())
	// Caches are unmapped in a finalizer. Ensure that the cache stays live
	// until after the call to hashimotoLight so it's not unmapped while being used.
	runtime.KeepAlive(cache)

	if !bytes.Equal(header.MixDigest[:], digest) {
		return errInvalidMixDigest
	}
	target := new(big.Int).Div(maxUint256, header.ActualDifficulty())
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the ethash protocol. The changes are done inline.
func (ethash *Ethash) Prepare(chain consensus.ChainReader, header *types.Header) error {
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Difficulty = ethash.CalcDifficulty(chain, header.Time.Uint64(), parent)
	return nil
}

// Finalize implements consensus.Engine, accumulating the block rewards,
// setting the final state and assembling the block.
func (ethash *Ethash) Finalize(chain consensus.ChainReader, header *types.Header, stateDB *state.StateDB, txs []*types.Transaction, receipts []*types.Receipt, gasReward uint64) (*types.Block, error) {
	if header.Number.Uint64() == V2Number {
		stateDB.SetBalance(state.EmptyAddress, "SERO", new(big.Int))
	}

	// Accumulate any block rewards and commit the final state root
	accumulateRewards(chain.Config(), stateDB, header, gasReward)
	header.Root = stateDB.IntermediateRoot(true)

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, receipts), nil
}
const V2Number = uint64(130000)
var (
	base    = big.NewInt(1e+17)
	big100  = big.NewInt(100)
	oneSero = new(big.Int).Mul(big.NewInt(10), base)

	oriReward    = new(big.Int).Mul(big.NewInt(66773505743), big.NewInt(1000000000))
	interval     = big.NewInt(8294400)
	halveNimber  = big.NewInt(3057600)
	difficultyL1 = big.NewInt(340000000)
	difficultyL2 = big.NewInt(1700000000)
	difficultyL3 = big.NewInt(4000000000)
	difficultyL4 = big.NewInt(17000000000)

	teamRewardPool      = common.BytesToAddress(crypto.Keccak512([]byte{1}))
	communityRewardPool = common.BytesToAddress(crypto.Keccak512([]byte{2}))

	teamAddress      = common.Base58ToAddress("RnRpAdXWaS2BnUzrUuzR8WPRfFackV65CzyqWU8mK4Np2aCgDUvrhYciDJoQZpMzWpaaKqsicf1u8fRd4ZKXeSUF2pMLHXXaiCX8XzHw3VRyX2Q7ko4BrRj9xTrNaErnTkg")
	communityAddress = common.Base58ToAddress("ZkVB2f8H1usYBSeViS7wPqSSFseXnCYXEbT2XxCSuRhfFg9KbBKbTvpTBj7dmSZxEKTp6rsqS3EX9js6StgRijZQBkaok2U5Fy8oLuGFrt1C5jwdAYB4Nqn8KNRniiQyCeb")
)


// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward .
func accumulateRewards(config *params.ChainConfig, statedb *state.StateDB, header *types.Header, gasReward uint64) {

	var reward *big.Int
	if header.Number.Uint64() >= V2Number {
		reward = accumulateRewardsV2(statedb, header)
	} else {
		reward = accumulateRewardsV1(config, statedb, header)
	}

	//log.Info(fmt.Sprintf("BlockNumber = %v, gasLimie = %v, gasUsed = %v, reward = %v", header.Number.Uint64(), header.GasLimit, header.GasUsed, reward))
	reward.Add(reward, new(big.Int).SetUint64(gasReward))
	asset := assets.Asset{Tkn: &assets.Token{
		Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
		Value:    utils.U256(*reward),
	},
	}
	statedb.GetZState().AddTxOut(header.Coinbase, asset)
}

func accumulateRewardsV1(config *params.ChainConfig, statedb *state.StateDB, header *types.Header) *big.Int {
	poolBalance := statedb.GetBalance(state.EmptyAddress, "SERO")
	if poolBalance.Sign() <= 0 {
		return big.NewInt(0)
	}

	reward := new(big.Int).Mul(big.NewInt(350), base)

	difficulty := big.NewInt(1717986918)
	if config.ChainID == params.AlphanetChainConfig.ChainID {
		difficulty = big.NewInt(51485767)
	} else if config.ChainID == params.DevnetChainConfig.ChainID {
		difficulty = big.NewInt(1048576)
	}

	if header.Difficulty.Cmp(difficulty) < 0 {
		ratio := new(big.Int).Div(new(big.Int).Mul(header.Difficulty, big100), difficulty).Uint64()
		if ratio >= 80 {
			reward = reward.Mul(reward, big.NewInt(4)).Div(reward, big.NewInt(5))
		} else if ratio >= 60 {
			reward = reward.Mul(reward, big.NewInt(3)).Div(reward, big.NewInt(5))
		} else if ratio >= 40 {
			reward = reward.Mul(reward, big.NewInt(2)).Div(reward, big.NewInt(5))
		} else if ratio >= 20 {
			reward = reward.Mul(reward, big.NewInt(1)).Div(reward, big.NewInt(5))
		} else {
			reward = big.NewInt(0).Set(oneSero)
		}
	}

	ratio := new(big.Int).Div(new(big.Int).Mul(new(big.Int).SetUint64(header.GasUsed), big100), new(big.Int).SetUint64(header.GasLimit)).Uint64()
	if ratio >= 80 {
		reward = new(big.Int).Div(new(big.Int).Mul(reward, big6), big.NewInt(5))
	} else {
		reward = reward.Mul(reward, big.NewInt(4)).Div(reward, big.NewInt(5))
	}

	if reward.Cmp(oneSero) < 0 {
		reward = big.NewInt(0).Set(oneSero)
	}
	statedb.SubBalance(state.EmptyAddress, "SERO", reward)
	return reward
}

func accumulateRewardsV2(statedb *state.StateDB, header *types.Header) *big.Int {
	reward := new(big.Int).Set(oriReward)
	if header.Number.Uint64() >= halveNimber.Uint64() {
		i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(header.Number, halveNimber), interval), big1)
		reward.Div(reward, new(big.Int).Exp(big2, i, nil))
	}

	var ratio *big.Int
	if header.Difficulty.Cmp(difficultyL1) < 0 { //<3.4
		reward = new(big.Int).Mul(big.NewInt(10), base)
		ratio = oriReward
	} else if header.Difficulty.Cmp(difficultyL2) < 0 { //<17
		ratio = new(big.Int).Add(new(big.Int).Mul(big.NewInt(56), base), new(big.Int).Mul(big.NewInt(16470000000), new(big.Int).Sub(header.Difficulty, difficultyL1)))
	} else if header.Difficulty.Cmp(difficultyL3) < 0 { //<40
		ratio = new(big.Int).Add(new(big.Int).Mul(big.NewInt(280), base), new(big.Int).Mul(big.NewInt(2170000000), new(big.Int).Sub(header.Difficulty, difficultyL2)))
	} else if header.Difficulty.Cmp(difficultyL4) < 0 { //<170
		ratio = new(big.Int).Add(new(big.Int).Mul(big.NewInt(330), base), new(big.Int).Mul(big.NewInt(2590000000), new(big.Int).Sub(header.Difficulty, difficultyL3)))
	} else {
		ratio = oriReward
	}

	reward = new(big.Int).Div(reward.Mul(reward, ratio), oriReward)
	if statedb == nil {
		return reward
	}
	statedb.AddBalance(communityRewardPool, "SERO", new(big.Int).Div(reward, big.NewInt(15)))
	statedb.AddBalance(teamRewardPool, "SERO", new(big.Int).Div(new(big.Int).Mul(reward, big2), big.NewInt(15)))

	if header.Number.Uint64()%5000 == 0 {
		balance := statedb.GetBalance(teamRewardPool, "SERO")
		statedb.SubBalance(teamRewardPool, "SERO", balance)
		assetTeam := assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*balance),
		},
		}
		statedb.GetZState().AddTxOut(teamAddress, assetTeam)

		balance = statedb.GetBalance(communityRewardPool, "SERO")
		statedb.SubBalance(communityRewardPool, "SERO", balance)
		assetCommunity := assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*balance),
		},
		}
		statedb.GetZState().AddTxOut(communityAddress, assetCommunity)
	}
	return reward
}