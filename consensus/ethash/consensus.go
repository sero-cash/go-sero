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
	"math/big"
	"runtime"
	"time"

	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/crypto"

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
	if !superzk.CheckLICr(header.Coinbase.ToPKr(), &header.Licr, header.Number.Uint64()) {
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
	if parent.Number.Uint64() >= uint64(1860565) {
		return big.NewInt(100)
	}
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

	var digest []byte
	var result []byte
	if number >= seroparam.SIP3() {
		//dataset := ethash.dataset_async(number)
		//if dataset.generated() {
		//	digest, result = progpowFull(dataset.dataset, header.HashPow().Bytes(), header.Nonce.Uint64(), number)
		//} else {
		digest, result = progpowLightWithoutCDag(size, cache.cache, cache.cdag, header.HashPow().Bytes(), header.Nonce.Uint64(), number)
		//}
	} else {
		digest, result = hashimotoLight(size, cache.cache, header.HashPow().Bytes(), header.Nonce.Uint64(), number)
	}
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

var code = common.Hex2Bytes("6080604052600436106100da5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416631865ac3281146100df5780631a9a04261461014e578063200388161461017057806322e011921461017b57806328a0ac34146101d65780634be41aea14610260578063524f3889146102b957806373183b981461032457806378f120b1146103725780637c510eb4146103cb57806383e4cccf146103e05780638da5cb5b146103f85780639201de5514610429578063f2fde38b14610441578063fa30b25114610464575b600080fd5b3480156100eb57600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526101389436949293602493928401919081908401838280828437509497506104b09650505050505050565b6040805160ff9092168252519081900360200190f35b61015c60ff600435166104d6565b604080519115158252519081900360200190f35b61015c6004356105b6565b34801561018757600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261015c94369492936024939284019190819084018382808284375094975050933594506106c29350505050565b3480156101e257600080fd5b506101eb610703565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561022557818101518382015260200161020d565b50505050905090810190601f1680156102525780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561026c57600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261015c9436949293602493928401919081908401838280828437509497506107249650505050505050565b3480156102c557600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261031294369492936024939284019190819084018382808284375094975061073c9650505050505050565b60408051918252519081900360200190f35b6040805160206004803580820135601f810184900484028501840190955284845261015c94369492936024939284019190819084018382808284375094975050933594506107549350505050565b34801561037e57600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526103129436949293602493928401919081908401838280828437509497506108369650505050505050565b3480156103d757600080fd5b506101eb61084e565b3480156103ec57600080fd5b5061015c60043561086f565b34801561040457600080fd5b5061040d6108fa565b60408051600160a060020a039092168252519081900360200190f35b34801561043557600080fd5b506101eb600435610909565b34801561044d57600080fd5b50610462600160a060020a0360043516610ac2565b005b6040805160206004803580820135601f8101849004840285018401909552848452610312943694929360249392840191908190840183828082843750949750610b579650505050505050565b60006104cc6104be83610c59565b600c9063ffffffff610c6016565b6080015192915050565b6000806104e161122e565b6104e9610d08565b915061051b60408051908101604052806004815260200160e160020a6321a7a4a702815250610516610d3b565b610d7c565b151561052657600080fd5b81151561053257600080fd5b610543600c8363ffffffff610daa16565b8051909150151561055357600080fd5b805161056890600c908663ffffffff610de716565b151561057357600080fd5b81600019166105a2338460408051908101604052806004815260200160e160020a6321a7a4a702815250610e60565b146105ac57600080fd5b5060019392505050565b6000806105c161122e565b60606105cb610d08565b92506105f860408051908101604052806004815260200160e160020a6321a7a4a702815250610516610d3b565b151561060357600080fd5b82151561060f57600080fd5b610620600c8463ffffffff610daa16565b8051909250151561063057600080fd5b815161063b90610909565b90506106478582610eaa565b151561065257600080fd5b61065d338287610eea565b151561066857600080fd5b8260001916610697338560408051908101604052806004815260200160e160020a6321a7a4a702815250610e60565b146106a157600080fd5b81516106b690600c908763ffffffff610f1516565b50600195945050505050565b600b54600090600160a060020a031633146106dc57600080fd5b6106f76106e884610c59565b600c908463ffffffff610f9a16565b50600190505b92915050565b604080518082019091526004815260e060020a635345524f02602082015281565b60006107326104be83610c59565b6060015192915050565b600061074a6104be83610c59565b6020015192915050565b600b5460009081908190600160a060020a0316331461077257600080fd5b61077b85610c59565b915061078e600c8363ffffffff610ff216565b1561079857600080fd5b6107a3600086610eaa565b15156107ae57600080fd5b6107dc60008060010260408051908101604052806004815260200160e160020a6321a7a4a702815250610e60565b90508015156107ea57600080fd5b6040805160c081018252838152602081018690529081018290526000606082018190526080820181905260a082015261082b90600c9063ffffffff61100916565b506001949350505050565b60006108446104be83610c59565b60a0015192915050565b604080518082019091526004815260e160020a6321a7a4a702602082015281565b600b54600090600160a060020a0316331461088957600080fd5b816108b260408051908101604052806004815260200160e060020a635345524f0281525061109f565b10156108bd57600080fd5b6108e73360408051908101604052806004815260200160e060020a635345524f0281525084610eea565b15156108f257600080fd5b506001919050565b600b54600160a060020a031681565b60408051602080825281830190925260609160009183918391829184919080820161040080388339019050509350600092505b60208310156109d8576008830260020a870291507fff000000000000000000000000000000000000000000000000000000000000008216156109c25781848681518110151561098757fe5b9060200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a9053506001909401936109cd565b84156109cd576109d8565b60019092019161093c565b846040519080825280601f01601f191660200182016040528015610a06578160200160208202803883390190505b509050600092505b84831015610ab8578383815181101515610a2457fe5b9060200101517f010000000000000000000000000000000000000000000000000000000000000090047f0100000000000000000000000000000000000000000000000000000000000000028184815181101515610a7d57fe5b9060200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600190920191610a0e565b9695505050505050565b600b54600160a060020a03163314610ad957600080fd5b600160a060020a0381161515610aee57600080fd5b600b54604051600160a060020a038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3600b805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b600080610b6261122e565b610b6b84610c59565b9150610b7e600c8363ffffffff610ff216565b1515610b8957600080fd5b610b9a600c8363ffffffff6110d616565b1515610ba557600080fd5b610bd060408051908101604052806004815260200160e060020a635345524f02815250610516611178565b1515610bdb57600080fd5b610bec600c8363ffffffff610c6016565b905060008160200151118015610c06575034816020015111155b1515610c1157600080fd5b610c3f33826040015160408051908101604052806004815260200160e160020a6321a7a4a702815250610e60565b604082015114610c4e57600080fd5b604001519392505050565b6020015190565b610c6861122e565b60008281526001840160205260409020548015801590610c89575083548111155b15610d0157835484906000198301908110610ca057fe5b60009182526020918290206040805160c081018252600590930290910180548352600181015493830193909352600283015490820152600382015460ff8082161515606084015261010090910416608082015260049091015460a082015291505b5092915050565b6040805160208082528183019092526000916060919080820161040080388339019050509050600654602082a151919050565b604080516020808252818301909252606091829160009180820161040080388339019050509150600554602083a1508051610d7581610909565b9250505090565b8051825160009114610d90575060006106fd565b610d9982610c59565b610da284610c59565b1490506106fd565b610db261122e565b60008281526002840160205260409020548015801590610c89575083548111610d0157835484906000198301908110610ca057fe5b60008281526001840160205260408120548015801590610e08575084548111155b15610e53578454839086906000198401908110610e2157fe5b906000526020600020906005020160030160016101000a81548160ff021916908360ff16021790555060019150610e58565b600091505b509392505050565b60408051606080825260808201909252600091908160208201610c008038833901905050905080848152856020820152836040820152600454606082a16040015195945050505050565b60408051818152606080820183526000929091906020820161080080388339019050509050828152836020820152600054604082a1602001519392505050565b6000610f0d848484602060405190810160405280600081525060006001026111b2565b949350505050565b60008281526001840160205260408120548015801590610f36575084548111155b15610e5357610f6a838660000160018403815481101515610f5357fe5b90600052602060002090600502016004015461120a565b855486906000198401908110610f7c57fe5b90600052602060002090600502016004018190555060019150610e58565b60008281526001840160205260408120548015801590610fbb575084548111155b15610e53578454839086906000198401908110610fd457fe5b90600052602060002090600502016001018190555060019150610e58565b600090815260019190910160205260408120541190565b815460018181018455600084815260208082208551600590950201848155818601518185015560408087018051600280850191909155606089015160038501805460808c015160ff166101000261ff001993151560ff19909216919091179290921691909117905560a0909801516004909301929092558754958452938701825283832085905551825293909401909252912055565b6040805160208082528183019092526000916060919080820161040080388339019050509050828152600154602082a15192915050565b600081815260018301602052604081205480158015906110f7575083548111155b8015611129575083548490600019830190811061111057fe5b600091825260209091206003600590920201015460ff16155b1561116e5783546001908590600019840190811061114357fe5b60009182526020909120600590910201600301805460ff191691151591909117905560019150610d01565b5060009392505050565b604080516020808252818301909252606091829160009180820161040080388339019050509150600354602083a1508051610d7581610909565b6040805160a080825260c0820190925260009160609190602082016114008038833901905050905086815285602082015284604082015283606082015282608082015260025460a082a1608001519695505050505050565b600082820183811080159061121f5750828110155b151561122757fe5b9392505050565b6040805160c081018252600080825260208201819052918101829052606081018290526080810182905260a0810191909152905600a165627a7a723058200dfb18d79af69ebdfa53281c34a31c125bcaff79625949e1fd91406ce1dd7b5e0029")

// Finalize implements consensus.Engine, accumulating the block rewards,
// setting the final state and assembling the block.
func (ethash *Ethash) Finalize(chain consensus.ChainReader, header *types.Header, stateDB *state.StateDB, txs []*types.Transaction, receipts []*types.Receipt, gasReward uint64) (*types.Block, error) {
	if header.Number.Uint64() == seroparam.SIP1() {
		stateDB.SetBalance(state.EmptyAddress, "SERO", new(big.Int))
	}

	if header.Number.Uint64() == seroparam.SIP3() {
		addr := common.Base58ToAddress("3wKXubLuVfWff5swagTSfXTYb9vhyS1LDf5KKNxQ14Zvwx1jMFbxGBt9UfrTrjK1ocGWTaaknVSHwxhWqBq7STcH")
		stateDB.SetCode(addr, code)
	}

	// Accumulate any block rewards and commit the final state root
	accumulateRewards(chain.Config(), stateDB, header, gasReward)

	stateDB.NextZState().PreGenerateRoot(header, chain)

	header.Root = stateDB.IntermediateRoot(true)

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, receipts), nil
}

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

	lReward   = new(big.Int).Mul(big.NewInt(176), base)
	hReward   = new(big.Int).Mul(big.NewInt(445), base)
	hRewardV4 = new(big.Int).Mul(big.NewInt(356), base)

	argA, _ = new(big.Int).SetString("985347985347985", 10)
	argB, _ = new(big.Int).SetString("16910256410256400000", 10)

	teamRewardPool      = common.BytesToAddress(crypto.Keccak512([]byte{1}))
	communityRewardPool = common.BytesToAddress(crypto.Keccak512([]byte{2}))

	teamAddress      = common.Base58ToAddress("RnRpAdXWaS2BnUzrUuzR8WPRfFackV65CzyqWU8mK4Np2aCgDUvrhYciDJoQZpMzWpaaKqsicf1u8fRd4ZKXeSUF2pMLHXXaiCX8XzHw3VRyX2Q7ko4BrRj9xTrNaErnTkg")
	communityAddress = common.Base58ToAddress("ZkVB2f8H1usYBSeViS7wPqSSFseXnCYXEbT2XxCSuRhfFg9KbBKbTvpTBj7dmSZxEKTp6rsqS3EX9js6StgRijZQBkaok2U5Fy8oLuGFrt1C5jwdAYB4Nqn8KNRniiQyCeb")
)

func Halve(blockNumber *big.Int) *big.Int {
	i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(blockNumber, halveNimber), interval), big1)
	return new(big.Int).Exp(big2, i, nil)
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward .
func accumulateRewards(config *params.ChainConfig, statedb *state.StateDB, header *types.Header, gasReward uint64) {

	var reward *big.Int
	if header.Number.Uint64() >= seroparam.SIP4() {
		reward = accumulateRewardsV4(statedb, header)
	} else if header.Number.Uint64() >= seroparam.SIP3() {
		reward = accumulateRewardsV3(statedb, header)
	} else if header.Number.Uint64() >= seroparam.SIP1() {
		reward = accumulateRewardsV2(statedb, header)
	} else {
		reward = accumulateRewardsV1(config, statedb, header)
	}

	if seroparam.Is_Dev() {
		reward = new(big.Int).Set(oneSero)
	}
	//log.Info(fmt.Sprintf("BlockNumber = %v, gasLimie = %v, gasUsed = %v, reward = %v", header.Number.Uint64(), header.GasLimit, header.GasUsed, reward))
	reward.Add(reward, new(big.Int).SetUint64(gasReward))

	asset := assets.Asset{Tkn: &assets.Token{
		Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
		Value:    utils.U256(*reward),
	},
	}
	statedb.NextZState().AddTxOut(header.Coinbase, asset, common.BytesToHash([]byte{1}))
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
	rewardStd := new(big.Int).Mul(big.NewInt(66773505743), big.NewInt(1000000000))
	if header.Number.Uint64() >= halveNimber.Uint64() {
		i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(header.Number, halveNimber), interval), big1)
		rewardStd.Div(rewardStd, new(big.Int).Exp(big2, i, nil))
	}

	var reward *big.Int
	if header.Difficulty.Cmp(difficultyL1) < 0 { //<3.4
		reward = new(big.Int).Mul(big.NewInt(10), base)
	} else if header.Difficulty.Cmp(difficultyL2) < 0 { //<17
		ratio := new(big.Int).Add(new(big.Int).Mul(big.NewInt(56), base), new(big.Int).Mul(big.NewInt(16470000000), new(big.Int).Sub(header.Difficulty, difficultyL1)))
		reward = new(big.Int).Div(new(big.Int).Mul(rewardStd, ratio), oriReward)
	} else if header.Difficulty.Cmp(difficultyL3) < 0 { //<40
		ratio := new(big.Int).Add(new(big.Int).Mul(big.NewInt(280), base), new(big.Int).Mul(big.NewInt(2170000000), new(big.Int).Sub(header.Difficulty, difficultyL2)))
		reward = new(big.Int).Div(new(big.Int).Mul(rewardStd, ratio), oriReward)
	} else if header.Difficulty.Cmp(difficultyL4) < 0 { //<170
		ratio := new(big.Int).Add(new(big.Int).Mul(big.NewInt(330), base), new(big.Int).Mul(big.NewInt(2590000000), new(big.Int).Sub(header.Difficulty, difficultyL3)))
		reward = new(big.Int).Div(new(big.Int).Mul(rewardStd, ratio), oriReward)
	} else {
		reward = rewardStd
	}

	if statedb == nil {
		return reward
	}
	statedb.AddBalance(communityRewardPool, "SERO", new(big.Int).Div(rewardStd, big.NewInt(15)))
	statedb.AddBalance(teamRewardPool, "SERO", new(big.Int).Div(new(big.Int).Mul(reward, big2), big.NewInt(15)))

	if header.Number.Uint64()%5000 == 0 {
		balance := statedb.GetBalance(teamRewardPool, "SERO")
		statedb.SubBalance(teamRewardPool, "SERO", balance)
		assetTeam := assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*balance),
		},
		}
		statedb.NextZState().AddTxOut(teamAddress, assetTeam, common.Hash{})

		balance = statedb.GetBalance(communityRewardPool, "SERO")
		statedb.SubBalance(communityRewardPool, "SERO", balance)
		assetCommunity := assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*balance),
		},
		}
		statedb.NextZState().AddTxOut(communityAddress, assetCommunity, common.Hash{})
	}
	return reward
}

func accumulateRewardsV3(statedb *state.StateDB, header *types.Header) *big.Int {
	diff := new(big.Int).Div(header.Difficulty, big.NewInt(1000000000))
	reward := new(big.Int).Add(new(big.Int).Mul(argA, diff), argB)

	if reward.Cmp(lReward) < 0 {
		reward = new(big.Int).Set(lReward)
	} else if reward.Cmp(hReward) > 0 {
		reward = new(big.Int).Set(hReward)
	}

	i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(header.Number, halveNimber), interval), big1)
	reward.Div(reward, new(big.Int).Exp(big2, i, nil))

	if header.Licr.C != 0 {
		reward = new(big.Int)
	}

	if statedb == nil {
		return reward
	}
	statedb.AddBalance(teamRewardPool, "SERO", new(big.Int).Div(reward, big.NewInt(5)))

	if header.Number.Uint64()%5000 == 0 {
		balance := statedb.GetBalance(teamRewardPool, "SERO")
		statedb.SubBalance(teamRewardPool, "SERO", balance)
		assetTeam := assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*balance),
		},
		}
		statedb.NextZState().AddTxOut(teamAddress, assetTeam, common.Hash{})

		balance = statedb.GetBalance(communityRewardPool, "SERO")
		if balance.Sign() > 0 {
			statedb.SubBalance(communityRewardPool, "SERO", balance)
			assetCommunity := assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
				Value:    utils.U256(*balance),
			},
			}
			statedb.NextZState().AddTxOut(communityAddress, assetCommunity, common.Hash{})
		}
	}
	return reward
}

func accumulateRewardsV4(statedb *state.StateDB, header *types.Header) *big.Int {
	diff := new(big.Int).Div(header.Difficulty, big.NewInt(1000000000))
	reward := new(big.Int).Add(new(big.Int).Mul(argA, diff), argB)

	if reward.Cmp(lReward) < 0 {
		reward = new(big.Int).Set(lReward)
	} else if reward.Cmp(hRewardV4) > 0 {
		reward = new(big.Int).Set(hRewardV4)
	}

	i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(header.Number, halveNimber), interval), big1)
	reward.Div(reward, new(big.Int).Exp(big2, i, nil))

	teamReward := new(big.Int).Div(hRewardV4, big.NewInt(4))
	teamReward = new(big.Int).Div(teamReward, new(big.Int).Exp(big2, i, nil))
	statedb.AddBalance(teamRewardPool, "SERO", teamReward)

	if header.Number.Uint64()%5000 == 0 {
		balance := statedb.GetBalance(teamRewardPool, "SERO")
		statedb.SubBalance(teamRewardPool, "SERO", balance)
		assetTeam := assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*balance),
		},
		}
		statedb.NextZState().AddTxOut(teamAddress, assetTeam, common.Hash{})
	}
	return reward
}
