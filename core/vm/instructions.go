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

package vm

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/rlp"

	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/log"

	"regexp"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/params"
)

var (
	bigZero                  = new(big.Int)
	tt255                    = math.BigPow(2, 255)
	errWriteProtection       = errors.New("evm: write protection")
	errReturnDataOutOfBounds = errors.New("evm: return data out of bounds")
	errExecutionReverted     = errors.New("evm: execution reverted")
	errMaxCodeSizeExceeded   = errors.New("evm: max code size exceeded")
	ErrToAddressError        = errors.New("evm: toAddr error")
	hashTrue                 = common.LeftPadBytes([]byte{1}, 32)
	hashFalse                = common.LeftPadBytes([]byte{0}, 32)

	topic_issueToken    = common.HexToHash("0x3be6bf24d822bcd6f6348f6f5a5c2d3108f04991ee63e80cde49a8c4746a0ef3")
	topic_send          = common.HexToHash("0x868bd6629e7c2e3d2ccf7b9968fad79b448e7a2bfb3ee20ed1acbc695c3c8b23")
	topic_balanceOf     = common.HexToHash("0xcf19eb4256453a4e30b6a06d651f1970c223fb6bd1826a28ed861f0e602db9b8")
	topic_allotTicket   = common.HexToHash("0xa6a366f1a72e1aef5d8d52ee240a476f619d15be7bc62d3df37496025b83459f")
	topic_currency      = common.HexToHash("0x7c98e64bd943448b4e24ef8c2cdec7b8b1275970cfe10daf2a9bfa4b04dce905")
	topic_category      = common.HexToHash("0xf1964f6690a0536daa42e5c575091297d2479edcc96f721ad85b95358644d276")
	topic_ticket        = common.HexToHash("0x9ab0d7c07029f006485cf3468ce7811aa8743b5a108599f6bec9367c50ac6aad")
	topic_setCallValues = common.HexToHash("0xa6cafc6282f61eff9032603a017e652f68410d3d3c69f0a3eeca8f181aec1d17")
	topic_setTokenRate  = common.HexToHash("0x6800e94e36131c049eaeb631e4530829b0d3d20d5b637c8015a8dc9cedd70aed")
	topic_closePkg      = common.HexToHash("0xbbf1aa2159b035802d0a4d44611849d5d4ada0329c81580477d5ec3e82f4f0a6")
	topic_transferPkg   = common.HexToHash("0xa8b83585a613dcf6c905ad7e0ce34cd07d1283cc72906d1fe78037d49adae455")
)

func opAdd(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	math.U256(y.Add(x, y))

	interpreter.intPool.put(x)
	return nil, nil
}

func opSub(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	math.U256(y.Sub(x, y))

	interpreter.intPool.put(x)
	return nil, nil
}

func opMul(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(math.U256(x.Mul(x, y)))

	interpreter.intPool.put(y)

	return nil, nil
}

func opDiv(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if y.Sign() != 0 {
		math.U256(y.Div(x, y))
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := interpreter.intPool.getZero()

	if y.Sign() == 0 || x.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() != y.Sign() {
			res.Div(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Div(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	interpreter.intPool.put(x, y)
	return nil, nil
}

func opMod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	if y.Sign() == 0 {
		stack.push(x.SetUint64(0))
	} else {
		stack.push(math.U256(x.Mod(x, y)))
	}
	interpreter.intPool.put(y)
	return nil, nil
}

func opSmod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := math.S256(stack.pop()), math.S256(stack.pop())
	res := interpreter.intPool.getZero()

	if y.Sign() == 0 {
		stack.push(res)
	} else {
		if x.Sign() < 0 {
			res.Mod(x.Abs(x), y.Abs(y))
			res.Neg(res)
		} else {
			res.Mod(x.Abs(x), y.Abs(y))
		}
		stack.push(math.U256(res))
	}
	interpreter.intPool.put(x, y)
	return nil, nil
}

func opExp(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	base, exponent := stack.pop(), stack.pop()
	stack.push(math.Exp(base, exponent))

	interpreter.intPool.put(base, exponent)

	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	back := stack.pop()
	if back.Cmp(big.NewInt(31)) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := stack.pop()
		mask := back.Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if num.Bit(int(bit)) > 0 {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}

		stack.push(math.U256(num))
	}

	interpreter.intPool.put(back)
	return nil, nil
}

func opNot(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	math.U256(x.Not(x))
	return nil, nil
}

func opLt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	return nil, nil
}

func opGt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	return nil, nil
}

func opSlt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(1)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(0)

	default:
		if x.Cmp(y) < 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	interpreter.intPool.put(x)
	return nil, nil
}

func opSgt(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()

	xSign := x.Cmp(tt255)
	ySign := y.Cmp(tt255)

	switch {
	case xSign >= 0 && ySign < 0:
		y.SetUint64(0)

	case xSign < 0 && ySign >= 0:
		y.SetUint64(1)

	default:
		if x.Cmp(y) > 0 {
			y.SetUint64(1)
		} else {
			y.SetUint64(0)
		}
	}
	interpreter.intPool.put(x)
	return nil, nil
}

func opEq(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Cmp(y) == 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	interpreter.intPool.put(x)
	return nil, nil
}

func opIszero(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if x.Sign() > 0 {
		x.SetUint64(0)
	} else {
		x.SetUint64(1)
	}
	return nil, nil
}

func opAnd(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.pop()
	stack.push(x.And(x, y))

	interpreter.intPool.put(y)
	return nil, nil
}

func opOr(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Or(x, y)

	interpreter.intPool.put(x)
	return nil, nil
}

func opXor(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Xor(x, y)

	interpreter.intPool.put(x)
	return nil, nil
}

func opByte(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	th, val := stack.pop(), stack.peek()
	if th.Cmp(common.Big32) < 0 {
		b := math.Byte(val, 32, int(th.Int64()))
		val.SetUint64(uint64(b))
	} else {
		val.SetUint64(0)
	}
	interpreter.intPool.put(th)
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Add(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}
	interpreter.intPool.put(y, z)
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.pop()
	if z.Cmp(bigZero) > 0 {
		x.Mul(x, y)
		x.Mod(x, z)
		stack.push(math.U256(x))
	} else {
		stack.push(x.SetUint64(0))
	}
	interpreter.intPool.put(y, z)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	defer interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Lsh(value, n))

	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := math.U256(stack.pop()), math.U256(stack.peek())
	defer interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		value.SetUint64(0)
		return nil, nil
	}
	n := uint(shift.Uint64())
	math.U256(value.Rsh(value, n))

	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Note, S256 returns (potentially) a new bigint, so we're popping, not peeking this one
	shift, value := math.U256(stack.pop()), math.S256(stack.pop())
	defer interpreter.intPool.put(shift) // First operand back into the pool

	if shift.Cmp(common.Big256) >= 0 {
		if value.Sign() > 0 {
			value.SetUint64(0)
		} else {
			value.SetInt64(-1)
		}
		stack.push(math.U256(value))
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.Rsh(value, n)
	stack.push(math.U256(value))

	return nil, nil
}

func opSha3(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	data := memory.Get(offset.Int64(), size.Int64())
	hash := crypto.Keccak256(data)
	evm := interpreter.evm

	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(common.BytesToHash(hash), data)
	}
	stack.push(interpreter.intPool.get().SetBytes(hash))

	interpreter.intPool.put(offset, size)
	return nil, nil
}

func opAddress(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(contract.Address().ToCaddr().Big())
	return nil, nil
}

func opBalance(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	slot.Set(new(big.Int))
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.evm.Origin.ToCaddr().Big())
	return nil, nil
}

func opCaller(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(contract.Caller().ToCaddr().Big())
	return nil, nil
}

func opCallValue(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().Set(contract.Value()))

	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().SetBytes(getDataBig(contract.Input, stack.pop(), big32)))
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().SetInt64(int64(len(contract.Input))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	memory.Set(memOffset.Uint64(), length.Uint64(), getDataBig(contract.Input, dataOffset, length))

	interpreter.intPool.put(memOffset, dataOffset, length)
	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().SetUint64(uint64(len(interpreter.returnData))))
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()

		end = interpreter.intPool.get().Add(dataOffset, length)
	)
	defer interpreter.intPool.put(memOffset, dataOffset, length, end)

	if end.BitLen() > 64 || uint64(len(interpreter.returnData)) < end.Uint64() {
		return nil, errReturnDataOutOfBounds
	}
	memory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[dataOffset.Uint64():end.Uint64()])

	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	address := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(slot))
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(address)))
	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	l := interpreter.intPool.get().SetInt64(int64(len(contract.Code)))
	stack.push(l)

	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	codeCopy := getDataBig(contract.Code, codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	interpreter.intPool.put(memOffset, codeOffset, length)
	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		addr       = common.BigToAddress(stack.pop())
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	codeCopy := getDataBig(interpreter.evm.StateDB.GetCode(addr), codeOffset, length)
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	interpreter.intPool.put(memOffset, codeOffset, length)
	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal contract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without contract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 wei, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	slot := stack.peek()
	slot.SetBytes(interpreter.evm.StateDB.GetCodeHash(common.BigToAddress(slot)).Bytes())
	return nil, nil
}

func opGasprice(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().Set(interpreter.evm.GasPrice))
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	num := stack.pop()

	n := interpreter.intPool.get().Sub(interpreter.evm.BlockNumber, common.Big257)
	if num.Cmp(n) > 0 && num.Cmp(interpreter.evm.BlockNumber) < 0 {
		stack.push(interpreter.evm.GetHash(num.Uint64()).Big())
	} else {
		stack.push(interpreter.intPool.getZero())
	}
	interpreter.intPool.put(num, n)
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.evm.Coinbase.ToCaddr().Big())
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(interpreter.intPool.get().Set(interpreter.evm.Time)))
	return nil, nil
}

func opNumber(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(interpreter.intPool.get().Set(interpreter.evm.BlockNumber)))
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(interpreter.intPool.get().Set(interpreter.evm.Difficulty)))
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(math.U256(interpreter.intPool.get().SetUint64(interpreter.evm.GasLimit)))
	return nil, nil
}

func opPop(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	interpreter.intPool.put(stack.pop())
	return nil, nil
}

func opMload(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset := stack.pop()
	val := interpreter.intPool.get().SetBytes(memory.Get(offset.Int64(), 32))
	stack.push(val)

	interpreter.intPool.put(offset)
	return nil, nil
}

func opMstore(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set32(mStart.Uint64(), val)

	interpreter.intPool.put(mStart, val)
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	off, val := stack.pop().Int64(), stack.pop().Int64()
	memory.store[off] = byte(val & 0xff)

	return nil, nil
}

func opSload(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := stack.peek()
	val := interpreter.evm.StateDB.GetState(contract.Address(), common.BigToHash(loc))
	loc.SetBytes(val.Bytes())
	return nil, nil
}

func opSstore(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := common.BigToHash(stack.pop())
	val := stack.pop()
	interpreter.evm.StateDB.SetState(contract.Address(), loc, common.BigToHash(val))

	interpreter.intPool.put(val)
	return nil, nil
}

func opJump(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos := stack.pop()
	if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
		nop := contract.GetOp(pos.Uint64())
		return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
	}
	*pc = pos.Uint64()

	interpreter.intPool.put(pos)
	return nil, nil
}

func opJumpi(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos, cond := stack.pop(), stack.pop()
	if cond.Sign() != 0 {
		if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
			nop := contract.GetOp(pos.Uint64())
			return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	interpreter.intPool.put(pos, cond)
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opPc(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().SetInt64(int64(memory.Len())))
	return nil, nil
}

func opGas(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(interpreter.intPool.get().SetUint64(contract.Gas))
	return nil, nil
}

func opCall(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in interpreter.evm.callGasTemp.
	interpreter.intPool.put(stack.pop())
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(addr))
	//TODO
	if toAddr == (common.Address{}) {
		return nil, ErrToAddressError
	}
	value = math.U256(value)
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	asset := contract.GetCallMsg()
	if asset == nil && value.Sign() != 0 {
		asset = &assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*value),
		},
		}
	}

	if asset != nil {
		gas += params.CallStipend
	}

	ret, returnGas, err := interpreter.evm.Call(contract, toAddr, args, gas, asset)
	contract.SetCallMsg(nil)
	if err != nil {
		stack.push(interpreter.intPool.getZero())
	} else {
		stack.push(interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, value, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opCallCode(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	interpreter.intPool.put(stack.pop())
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(addr))
	//TODO
	if toAddr == (common.Address{}) {
		return nil, ErrToAddressError
	}
	value = math.U256(value)
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	asset := contract.GetCallMsg()
	if asset == nil && value.Sign() != 0 {
		asset = &assets.Asset{Tkn: &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
			Value:    utils.U256(*value),
		},
		}
	}

	if asset != nil {
		gas += params.CallStipend
	}

	ret, returnGas, err := interpreter.evm.Call(contract, toAddr, args, gas, asset)
	contract.SetCallMsg(nil)
	if err != nil {
		stack.push(interpreter.intPool.getZero())
	} else {
		stack.push(interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, value, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	interpreter.intPool.put(stack.pop())
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(addr))
	//TODO
	if toAddr == (common.Address{}) {
		return nil, ErrToAddressError
	}
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := interpreter.evm.DelegateCall(contract, toAddr, args, gas)
	if err != nil {
		stack.push(interpreter.intPool.getZero())
	} else {
		stack.push(interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	interpreter.intPool.put(stack.pop())
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(addr))
	//TODO
	if toAddr == (common.Address{}) {
		return nil, ErrToAddressError
	}
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := interpreter.evm.StaticCall(contract, toAddr, args, gas)
	if err != nil {
		stack.push(interpreter.intPool.getZero())
	} else {
		stack.push(interpreter.intPool.get().SetUint64(1))
	}
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opReturn(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	interpreter.intPool.put(offset, size)
	return ret, nil
}

func opRevert(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	interpreter.intPool.put(offset, size)
	return ret, nil
}

func opStop(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	//toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(stack.pop()))
	//interpreter.evm.StateDB.Suicide(contract.Address(), toAddr)
	return nil, nil
}

func handleAllotTicket(d []byte, evm *EVM, contract *Contract, mem []byte) (common.Hash, uint64, error) {
	offset := new(big.Int).SetBytes(d[64:96]).Uint64()
	len := new(big.Int).SetBytes(mem[offset:offset+32]).Uint64()
	if len == 0 {
		return common.Hash{}, 0, fmt.Errorf("allotTicket error , contract : %s, error : %s", contract.Address(), "nameLen is zero")
	}

	categoryName := string(mem[offset+32 : offset+32+len])
	match, err := regexp.Match("^[A-Z][A-Z0-9_]{0,31}$", []byte(categoryName))
	if err != nil || !match {
		return common.Hash{}, 0, fmt.Errorf("allotTicket error , contract : %s, error : %s", contract.Address(), "illegal categoryName")
	}

	categoryName = strings.ToUpper(categoryName)
	if strings.Contains(categoryName, "SERO") {
		return common.Hash{}, 0, fmt.Errorf("allotTicket error , contract : %s, error : %s", contract.Address(), "categoryName can not contains SERO")
	}
	value := common.BytesToHash(d[0:32])
	if value == (common.Hash{}) {
		if !evm.StateDB.RegisterTicket(contract.Address(), categoryName) {
			return common.Hash{}, 0, fmt.Errorf("allotTicket error , contract : %s, error : %s", contract.Address(), "categoryName registered by other")
		}

		nonce := evm.StateDB.GetTicketNonce(contract.Address())
		evm.StateDB.SetTicketNonce(contract.Address(), nonce+1)
		bytes, _ := rlp.EncodeToBytes([]interface{}{contract.Address(), categoryName, nonce})
		value = crypto.Keccak256Hash(bytes)
		evm.StateDB.AddTicket(contract.Address(), categoryName, value)
	}

	toAddr := evm.StateDB.GetNonceAddress(d[44:64])
	if toAddr != (common.Address{}) && toAddr != contract.Address() {
		asset := assets.Asset{
			Tkt: &assets.Ticket{
				Category: *common.BytesToHash(common.LeftPadBytes([]byte(categoryName), 32)).HashToUint256(),
				Value:    *value.HashToUint256(),
			},
		}

		gas := evm.callGasTemp + params.CallStipend
		_, returnGas, err := evm.Call(contract, toAddr, nil, gas, &asset)
		//contract.Gas += returnGas
		if err != nil {
			return common.Hash{}, returnGas, err
		}
	}

	return value, evm.callGasTemp, nil
}

var (
	base   = new(big.Int).SetUint64(1e+17)
	level1 = new(big.Int).Mul(big.NewInt(10000000), base)
	level2 = new(big.Int).Mul(big.NewInt(5000000), base)
	level3 = new(big.Int).Mul(big.NewInt(100000), base)
	level4 = new(big.Int).Mul(big.NewInt(10000), base)
	level5 = new(big.Int).Mul(big.NewInt(100), base)
	level6 = new(big.Int).Mul(big.NewInt(10), base)
	level7 = new(big.Int).Mul(big.NewInt(1), base)
)

func tokenFee(coinName string) *big.Int {
	switch len(coinName) {
	case 1:
		return level1
	case 2:
		return level2
	case 3:
		return level3
	case 4:
		return level4
	case 5:
		return level5
	case 6:
		return level6
	}
	return level7
}

var foundationAccount1 = common.Base58ToAddress("hcZCikh7h3By7FBoDhCj8a6swKukeZRKm5Xg1s9RStPTNV7GzE3rJDzJsDwCWy9p86A8amMdCHgF7jKaMfSt1zfaBEbqkapuveHyFko2V9ZKisMC5qMp4VacYnApXokr4ea")

var foundationAccount2 = common.Base58ToAddress("5niHmAcSoDzaekKTUpLR3qkQf6djC7AGhnJnuPr8w7ArqQzhxyhEf61Rp68WhpoYo57r5q8CVsLopTJ9uc5VS92fRSHsjBqY9rqMJfQ4DBMw5QyXvT4oyeF7P9sb7ruvwZD")

func handleIssueToken(d []byte, evm *EVM, contract *Contract, mem []byte) (bool, error) {
	offset := new(big.Int).SetBytes(d[0:32]).Uint64()
	len := new(big.Int).SetBytes(mem[offset:offset+32]).Uint64()
	if len == 0 {
		return false, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "nameLen is zero")
	}

	coinName := string(mem[offset+32 : offset+32+len])
	match, err := regexp.Match("^[A-Z][A-Z0-9_]{0,31}$", []byte(coinName))
	if err != nil || !match {
		return false, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "illegal coinName")
	}

	coinName = strings.ToUpper(coinName)
	if strings.Contains(coinName, "SERO") {
		return false, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "coinName can not contains SERO")
	}
	address := evm.StateDB.GetContrctAddressByToken(coinName)
	if address == (common.Address{}) {
		fee := new(big.Int).Set(level6)
		if evm.chainConfig.ChainID.Uint64() == 2019 {
			fee = tokenFee(coinName)
		}
		if evm.StateDB.GetBalance(contract.Address(), "SERO").Cmp(fee) >= 0 {
			if !evm.StateDB.RegisterToken(contract.Address(), coinName) {
				return false, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "coinName registered by other")
			} else {
				evm.StateDB.SubBalance(contract.Address(), "SERO", fee)
				asset := assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
					Value:    utils.U256(*fee),
				},
				}
				if (evm.BlockNumber.Uint64() >= 300000) {
					if err := evm.StateDB.GetZState().AddTxOutWithCheck(foundationAccount2, asset); err != nil {
						return false, err
					}
				} else {
					if err := evm.StateDB.GetZState().AddTxOutWithCheck(foundationAccount1, asset); err != nil {
						return false, err
					}
				}
			}
		} else {
			return false, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "insufficient balance for token fee")
		}
	} else if address != contract.Address() {
		return false, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "coinName registered by other")
	}

	total := new(big.Int).SetBytes(d[32:64])
	evm.StateDB.AddBalance(contract.Address(), coinName, total)
	return true, nil
}

func handleSend(d []byte, evm *EVM, contract *Contract, mem []byte) ([]byte, uint64, error) {
	addr := common.BytesToContractAddress(d[12:32])
	toAddr := evm.StateDB.GetNonceAddress(addr[:])
	if toAddr == (common.Address{}) {
		return nil, 0, fmt.Errorf("handleSend error , contract : %s, toAddr : %s, error : %s", contract.Address(), toAddr, "not load toAddrss")
	}
	currency_offset := new(big.Int).SetBytes(d[32:64]).Uint64()
	length := new(big.Int).SetBytes(mem[currency_offset:currency_offset+32]).Uint64()
	var currency string
	if length != 0 {
		currency = string(mem[currency_offset+32 : currency_offset+32+length])
	}

	var category string
	category_offset := new(big.Int).SetBytes(d[96:128]).Uint64()
	length = new(big.Int).SetBytes(mem[category_offset:category_offset+32]).Uint64()
	if length != 0 {
		category = string(mem[category_offset+32 : category_offset+32+length])
	}

	currency = strings.ToUpper(currency)
	category = strings.ToUpper(category)

	amount := new(big.Int).SetBytes(d[64:96])
	ticketHash := common.BytesToHash(d[128:160])

	var token *assets.Token
	if len(currency) != 0 && amount.Sign() != 0 {
		balance := evm.StateDB.GetBalance(contract.Address(), currency)
		if balance.Cmp(amount) < 0 {
			return nil, 0, fmt.Errorf("handleSend error , contract : %s, toAddr : %s, error : %s", contract.Address(), toAddr, "balance not enough")
		}
		token = &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
			Value:    utils.U256(*amount),
		}
	}

	var ticket *assets.Ticket
	if len(category) != 0 && ticketHash != (common.Hash{}) {
		if !evm.StateDB.OwnTicket(contract.Address(), category, ticketHash) {
			return nil, 0, fmt.Errorf("handleSend error , contract : %s, toAddr : %s, error : %s", contract.Address(), toAddr, "ticket not own")
		}
		ticket = &assets.Ticket{
			Category: *common.BytesToHash(common.LeftPadBytes([]byte(category), 32)).HashToUint256(),
			Value:    *ticketHash.HashToUint256(),
		}
	}

	asset := assets.Asset{Tkn: token, Tkt: ticket}
	gas := evm.callGasTemp + params.CallStipend
	return evm.Call(contract, toAddr, nil, gas, &asset)
}

var (
	memLens = map[common.Hash]int{
		topic_issueToken:    64,
		topic_send:          160,
		topic_balanceOf:     32,
		topic_allotTicket:   96,
		topic_currency:      32,
		topic_category:      32,
		topic_ticket:        32,
		topic_setCallValues: 128,
		topic_setTokenRate:  96,
		topic_closePkg:      256,
		topic_transferPkg:   64,
	}
)

func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i] = common.BigToHash(stack.pop())
		}
		data := memory.Get(0, int64(memory.Len()))

		d := memory.Get(mStart.Int64(), mSize.Int64())
		length := mSize.Uint64()
		if l, ok := memLens[topics[0]]; ok {
			if l != len(d) {
				return d, ErrCodeInvalid
			}
		}
		if topics[0] == topic_allotTicket {
			hash, returnGas, err := handleAllotTicket(d, interpreter.evm, contract, data)
			contract.Gas += returnGas
			if err != nil {
				log.Trace("IssueToken error ", "contract", contract.Address(), "error", err)
			}
			memory.Set(mStart.Uint64()+length-32, 32, hash[:])
		} else if topics[0] == topic_issueToken {
			if ok, err := handleIssueToken(d, interpreter.evm, contract, data); ok {
				log.Trace("IssueToken error ", "contract", contract.Address(), "error", err)
				memory.Set(mStart.Uint64()+length-32, 32, hashTrue)
			} else {
				log.Trace(err.Error())
				memory.Set(mStart.Uint64()+length-32, 32, hashFalse)
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_balanceOf {
			offset := new(big.Int).SetBytes(d[0:32]).Uint64()
			len := new(big.Int).SetBytes(data[offset:offset+32]).Uint64()
			balance := new(big.Int)
			if len != 0 {
				coinName := string(data[offset+32 : offset+32+len])
				balance = interpreter.evm.StateDB.GetBalance(contract.Address(), coinName)
			}
			memory.Set(mStart.Uint64(), 32, common.LeftPadBytes(balance.Bytes(), 32))
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_send {
			_, returnGas, err := handleSend(d, interpreter.evm, contract, data)
			contract.Gas += returnGas
			if err != nil {
				log.Trace("send error ", "contract", contract.Address(), "error", err)
				memory.Set(mStart.Uint64()+length-32, 32, hashFalse)
			} else {
				memory.Set(mStart.Uint64()+length-32, 32, hashTrue)
			}
		} else if topics[0] == topic_currency {
			if contract.asset != nil && contract.asset.Tkn != nil {
				currency := strings.Trim(string(contract.asset.Tkn.Currency[:]), string([]byte{0}))
				memory.Set(mStart.Uint64(), 32, []byte(currency))
			} else {
				memory.Set(mStart.Uint64(), 32, []byte{})
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_category {
			if contract.asset != nil && contract.asset.Tkt != nil {
				category := strings.Trim(string(contract.asset.Tkt.Category[:]), string([]byte{0}))
				memory.Set(mStart.Uint64(), 32, []byte(category))
			} else {
				memory.Set(mStart.Uint64(), 32, []byte{})
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_ticket {
			if contract.asset != nil && contract.asset.Tkt != nil {
				memory.Set(mStart.Uint64(), 32, contract.asset.Tkt.Value[:])
			} else {
				memory.Set(mStart.Uint64(), 32, []byte{})
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_setCallValues {
			setCallValues(d, data, contract)
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_setTokenRate {
			offset := new(big.Int).SetBytes(d[0:32]).Uint64()
			len := new(big.Int).SetBytes(data[offset:offset+32]).Uint64()
			if len == 0 {
				return nil, fmt.Errorf("setTokenRate error , contract : %s, error : %s", contract.Address(), "coinName len=0")
			}

			coinName := string(data[offset+32 : offset+32+len])
			match, err := regexp.Match("^[A-Z][A-Z0-9_]{0,31}$", []byte(coinName))
			if err != nil || !match {
				return nil, fmt.Errorf("issueToken error , contract : %s, error : %s", contract.Address(), "illegal coinName")
			}

			if interpreter.evm.StateDB.SetTokenRate(contract.Address(), coinName, new(big.Int).SetBytes(d[32:64]), new(big.Int).SetBytes(d[64:96])) {
				memory.Set(mStart.Uint64()+length-32, 32, hashTrue)
			} else {
				memory.Set(mStart.Uint64()+length-32, 32, hashFalse)
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_closePkg {
			id := keys.Uint256{}
			copy(id[:], d[0:32])

			key := keys.Uint256{}
			copy(key[:], d[32:64])
			pkg, err := interpreter.evm.StateDB.GetPkgState().Close(&id, contract.Address().ToPKr(), &key)
			if err != nil {
				memory.Set(mStart.Uint64(), 256, make([]byte, 256))
			} else {
				if pkg.O.Asset.Tkn != nil {
					currency := common.BytesToString(pkg.O.Asset.Tkn.Currency[:])
					amount := pkg.O.Asset.Tkn.Value.ToIntRef()
					if len(currency) != 0 && amount.Sign() > 0 {
						interpreter.evm.StateDB.AddBalance(contract.Address(), currency, amount)
					}
					memory.Set(mStart.Uint64(), 32, pkg.O.Asset.Tkn.Currency[:])
					hash := common.BigToHash(pkg.O.Asset.Tkn.Value.ToIntRef())
					memory.Set(mStart.Uint64()+32, 32, hash[:])
				}

				if pkg.O.Asset.Tkt != nil {
					category := common.BytesToString(pkg.O.Asset.Tkt.Category[:])
					ticket := common.BytesToHash(pkg.O.Asset.Tkt.Value[:])
					if len(category) != 0 && ticket != (common.Hash{}) {
						interpreter.evm.StateDB.AddTicket(contract.Address(), category, ticket)
					}
					memory.Set(mStart.Uint64()+64, 32, pkg.O.Asset.Tkt.Category[:])
					memory.Set(mStart.Uint64()+96, 32, pkg.O.Asset.Tkt.Value[:])
				}
				from := common.BytesToAddress(pkg.Z.From[:]).ToCaddr()
				memory.Set(mStart.Uint64()+128, 32, common.LeftPadBytes(from[:], 32))
				memory.Set(mStart.Uint64()+160, 32, common.LeftPadBytes(big.NewInt(0).SetUint64(pkg.Z.High).Bytes(), 32))
				if len(pkg.O.Memo) > 0 {
					memory.Set(mStart.Uint64()+192, 32, pkg.O.Memo[0:32])
					memory.Set(mStart.Uint64()+224, 32, pkg.O.Memo[32:64])
				}
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else if topics[0] == topic_transferPkg {
			id := keys.Uint256{}
			copy(id[:], d[0:32])

			toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BytesToContractAddress(d[32:64]))
			if toAddr == (common.Address{}) {
				return nil, ErrToAddressError
			}
			if err := interpreter.evm.StateDB.GetPkgState().Transfer(&id, contract.Address().ToPKr(), toAddr.ToPKr()); err != nil {
				memory.Set(mStart.Uint64()+length-32, 32, hashFalse)
			} else {
				memory.Set(mStart.Uint64()+length-32, 32, hashTrue)
			}
			contract.Gas += interpreter.evm.callGasTemp
		} else {
			interpreter.evm.StateDB.AddLog(&types.Log{
				Address: contract.Address(),
				Topics:  topics,
				Data:    d,
				// This is a non-consensus field, but assigned here because
				// core/state doesn't know the current block number.
				BlockNumber: interpreter.evm.BlockNumber.Uint64(),
			})
			contract.Gas += interpreter.evm.callGasTemp
		}

		interpreter.intPool.put(mStart, mSize)
		return nil, nil
	}
}

func setCallValues(d []byte, data []byte, contract *Contract) {
	currency_offset := new(big.Int).SetBytes(d[0:32]).Uint64()
	length := new(big.Int).SetBytes(data[currency_offset:currency_offset+32]).Uint64()
	var currency string
	if length != 0 {
		currency = string(data[currency_offset+32 : currency_offset+32+length])
	}
	var category string
	category_offset := new(big.Int).SetBytes(d[64:96]).Uint64()
	length = new(big.Int).SetBytes(data[category_offset:category_offset+32]).Uint64()
	if length != 0 {
		category = string(data[category_offset+32 : category_offset+32+length])
	}
	currency = strings.ToUpper(currency)
	category = strings.ToUpper(category)
	amount := new(big.Int).SetBytes(d[32:64])
	ticketHash := common.BytesToHash(d[96:128])
	var token *assets.Token
	if len(currency) != 0 && amount.Sign() != 0 {
		token = &assets.Token{
			Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
			Value:    utils.U256(*amount),
		}
	}
	var ticket *assets.Ticket
	if len(category) != 0 && ticketHash != (common.Hash{}) {
		ticket = &assets.Ticket{
			Category: *common.BytesToHash(common.LeftPadBytes([]byte(category), 32)).HashToUint256(),
			Value:    *ticketHash.HashToUint256(),
		}
	}
	if token != nil || ticket != nil {
		contract.SetCallMsg(&assets.Asset{Tkn: token, Tkt: ticket})
	}
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		codeLen := len(contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := interpreter.intPool.get()
		stack.push(integer.SetBytes(common.RightPadBytes(contract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.dup(interpreter.intPool, int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.swap(int(size))
		return nil, nil
	}
}
