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

package vm

import (
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/params"
	"math/big"
)

// memoryGasCosts calculates the quadratic gas for memory expansion. It does so
// only for the memory region that is expanded, not the total memory.
func memoryGasCost(mem *Memory, newMemSize uint64) (uint64, error) {

	if newMemSize == 0 {
		return 0, nil
	}
	// The maximum that will fit in a uint64 is max_word_count - 1
	// anything above that will result in an overflow.
	// Additionally, a newMemSize which results in a
	// newMemSizeWords larger than 0x7ffffffff will cause the square operation
	// to overflow.
	// The constant 0xffffffffe0 is the highest number that can be used without
	// overflowing the gas calculation
	if newMemSize > 0xffffffffe0 {
		return 0, errGasUintOverflow
	}

	newMemSizeWords := toWordSize(newMemSize)
	newMemSize = newMemSizeWords * 32

	if newMemSize > uint64(mem.Len()) {
		square := newMemSizeWords * newMemSizeWords
		linCoef := newMemSizeWords * params.MemoryGas
		quadCoef := square / params.QuadCoeffDiv
		newTotalFee := linCoef + quadCoef

		fee := newTotalFee - mem.lastGasCost
		mem.lastGasCost = newTotalFee

		return fee, nil
	}
	return 0, nil
}

func constGasFunc(gas uint64) gasFunc {
	return func(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		return gas, nil
	}
}

func gasCallDataCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	words, overflow := bigUint64(stack.Back(2))
	if overflow {
		return 0, errGasUintOverflow
	}

	if words, overflow = math.SafeMul(toWordSize(words), params.CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = math.SafeAdd(gas, words); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasReturnDataCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	words, overflow := bigUint64(stack.Back(2))
	if overflow {
		return 0, errGasUintOverflow
	}

	if words, overflow = math.SafeMul(toWordSize(words), params.CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = math.SafeAdd(gas, words); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasSStore(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		y, x = stack.Back(1), stack.Back(0)
		val  = evm.StateDB.GetState(contract.Address(), common.BigToHash(x))
	)
	// This checks for 3 scenario's and calculates gas accordingly
	// 1. From a zero-value address to a non-zero value         (NEW VALUE)
	// 2. From a non-zero value address to a zero-value address (DELETE)
	// 3. From a non-zero to a non-zero                         (CHANGE)
	if val == (common.Hash{}) && y.Sign() != 0 {
		// 0 => non 0
		return params.SstoreSetGas, nil
	} else if val != (common.Hash{}) && y.Sign() == 0 {
		// non 0 => 0
		evm.StateDB.AddRefund(params.SstoreRefundGas)
		return params.SstoreClearGas, nil
	} else {
		// non 0 => non 0 (or 0 => 0)
		return params.SstoreResetGas, nil
	}
}

var (
	topics = []string{
		"0x3be6bf24d822bcd6f6348f6f5a5c2d3108f04991ee63e80cde49a8c4746a0ef3",
		"0x868bd6629e7c2e3d2ccf7b9968fad79b448e7a2bfb3ee20ed1acbc695c3c8b23",
		"0xcf19eb4256453a4e30b6a06d651f1970c223fb6bd1826a28ed861f0e602db9b8",
		"0xa6a366f1a72e1aef5d8d52ee240a476f619d15be7bc62d3df37496025b83459f",
		"0x7c98e64bd943448b4e24ef8c2cdec7b8b1275970cfe10daf2a9bfa4b04dce905",
		"0xf1964f6690a0536daa42e5c575091297d2479edcc96f721ad85b95358644d276",
		"0x9ab0d7c07029f006485cf3468ce7811aa8743b5a108599f6bec9367c50ac6aad",
		"0xa6cafc6282f61eff9032603a017e652f68410d3d3c69f0a3eeca8f181aec1d17",
		"0x6800e94e36131c049eaeb631e4530829b0d3d20d5b637c8015a8dc9cedd70aed",
		"0xbbf1aa2159b035802d0a4d44611849d5d4ada0329c81580477d5ec3e82f4f0a6",
		"0xa8b83585a613dcf6c905ad7e0ce34cd07d1283cc72906d1fe78037d49adae455",
	}

	sendTopic        = "0x868bd6629e7c2e3d2ccf7b9968fad79b448e7a2bfb3ee20ed1acbc695c3c8b23"
	allotTicketTopic = "0xa6a366f1a72e1aef5d8d52ee240a476f619d15be7bc62d3df37496025b83459f"
)

func inTopics(val string) bool {
	for _, each := range topics {
		if (each == val) {
			return true;
		}
	}
	return false;
}

func makeGasLog(n uint64) gasFunc {
	return func(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
		requestedSize, overflow := bigUint64(stack.Back(1))
		if overflow {
			return 0, errGasUintOverflow
		}

		gas, err := memoryGasCost(mem, memorySize)
		if err != nil {
			return 0, err
		}

		if gas, overflow = math.SafeAdd(gas, params.LogGas); overflow {
			return 0, errGasUintOverflow
		}
		if gas, overflow = math.SafeAdd(gas, n*params.LogTopicGas); overflow {
			return 0, errGasUintOverflow
		}

		var memorySizeGas uint64
		if memorySizeGas, overflow = math.SafeMul(requestedSize, params.LogDataGas); overflow {
			return 0, errGasUintOverflow
		}
		if gas, overflow = math.SafeAdd(gas, memorySizeGas); overflow {
			return 0, errGasUintOverflow
		}

		if (evm.BlockNumber.Uint64() < seroparam.SIP5()) {
			evm.callGasTemp, err = callGas(gt, contract.Gas, gas, new(big.Int).SetUint64(math.MaxUint64))
			if err != nil {
				return 0, err
			}
			if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
				return 0, errGasUintOverflow
			}
		} else {
			evm.callGasTemp = 0;
			topic := common.BigToHash(stack.Back(2)).Hex();
			if (sendTopic == topic || allotTicketTopic == topic) {
				evm.callGasTemp, err = callGas(gt, contract.Gas, gas, new(big.Int).SetUint64(0))
				if err != nil {
					return 0, err
				}
				if (allotTicketTopic == topic) {
					evm.callGasTemp += params.CreateTicketGas
				}
				if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
					return 0, errGasUintOverflow
				}
			}
		}

		return gas, nil
	}
}

func gasSha3(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	if gas, overflow = math.SafeAdd(gas, params.Sha3Gas); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(1))
	if overflow {
		return 0, errGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), params.Sha3WordGas); overflow {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCodeCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(2))
	if overflow {
		return 0, errGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), params.CopyGas); overflow {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasExtCodeCopy(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}

	var overflow bool
	if gas, overflow = math.SafeAdd(gas, gt.ExtcodeCopy); overflow {
		return 0, errGasUintOverflow
	}

	wordGas, overflow := bigUint64(stack.Back(3))
	if overflow {
		return 0, errGasUintOverflow
	}

	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), params.CopyGas); overflow {
		return 0, errGasUintOverflow
	}

	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasExtCodeHash(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.ExtcodeHash, nil
}

func gasMLoad(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMStore8(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasMStore(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, errGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, GasFastestStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCreate(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, params.CreateGas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCreate2(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var overflow bool
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, params.Create2Gas); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasBalance(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.Balance, nil
}

func gasExtCodeSize(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.ExtcodeSize, nil
}

func gasSLoad(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return gt.SLoad, nil
}

func gasExp(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	expByteLen := uint64((stack.data[stack.len()-2].BitLen() + 7) / 8)

	var (
		gas      = expByteLen * gt.ExpByte // no overflow check required. Max is 256 * ExpByte gas
		overflow bool
	)
	if gas, overflow = math.SafeAdd(gas, GasSlowStep); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		gas            = gt.Calls
		transfersValue = stack.Back(2).Sign() != 0
		// address        = common.BigToAddress(stack.Back(1))
	)
	// if !evm.StateDB.Exist(address) {
	//	gas += params.CallNewAccountGas
	// }
	if transfersValue || contract.callMsg != nil {
		gas += params.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasCallCode(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas := gt.Calls
	if stack.Back(2).Sign() != 0 {
		gas += params.CallValueTransferGas
	}
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasReturn(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(mem, memorySize)
}

func gasRevert(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return memoryGasCost(mem, memorySize)
}

func gasSuicide(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var gas uint64
	if !evm.StateDB.HasSuicided(contract.Address()) {
		evm.StateDB.AddRefund(params.SuicideRefundGas)
	}
	return gas, nil
}

func gasDelegateCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, gt.Calls); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasStaticCall(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(gas, gt.Calls); overflow {
		return 0, errGasUintOverflow
	}

	evm.callGasTemp, err = callGas(gt, contract.Gas, gas, stack.Back(0))
	if err != nil {
		return 0, err
	}
	if gas, overflow = math.SafeAdd(gas, evm.callGasTemp); overflow {
		return 0, errGasUintOverflow
	}
	return gas, nil
}

func gasPush(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func gasSwap(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}

func gasDup(gt params.GasTable, evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	return GasFastestStep, nil
}
