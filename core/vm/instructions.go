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
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/big"
	"strings"

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
	hashZero                 = common.LeftPadBytes([]byte{0}, 32)

	topic_issueToken    = common.HexToHash("0x25d7b0676eb16f8e7d1160b577510f33f54887867efcbc6ec3c712354706b6b0")
	topic_sendAnonymous = common.HexToHash("0x2da7d636de04e3fdf864ae424b77b7c8ca8561cafd21f3317d8e131b96a22b02")
	topic_selfBalance   = common.HexToHash("0x5c8c6e4c5998a1ee83ab3aa45fa23ca7c8ed4a75c654632eb22adcc5e1e6fb37")
	topic_issueTicket   = common.HexToHash("0xa6a77d98ff69ab4b74330af67416ef5d357dfb5c0048677a7f1675bf4f67076b")
	topic_getCurrency   = common.HexToHash("0x71b861b13fc896f3c531ebd724c652505e48aa8fbbba3a50f6d6804aa791a22f")
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
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(common.BigToAddress(slot))))

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

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	pkg := assets.Package{Tkn: &assets.Token{
		Currency: *common.BytesToHash(common.LeftPadBytes([]byte("sero"), 32)).HashToUint256(),
		Value:    utils.U256(*value),
	},
	}
	ret, returnGas, err := interpreter.evm.Call(contract, toAddr, args, gas, pkg)
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

	if value.Sign() != 0 {
		gas += params.CallStipend
	}

	pkg := assets.Package{Tkn: &assets.Token{
		Currency: *common.BytesToHash(common.LeftPadBytes([]byte("sero"), 32)).HashToUint256(),
		Value:    utils.U256(*value),
	},
	}
	ret, returnGas, err := interpreter.evm.CallCode(contract, toAddr, args, gas, pkg)
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
	toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, common.BigToContractAddress(stack.pop()))
	interpreter.evm.StateDB.Suicide(contract.Address(), toAddr)
	return nil, nil
}

func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i] = common.BigToHash(stack.pop())
		}

		d := memory.Get(mStart.Int64(), mSize.Int64())

		end := mSize.Uint64()
		if topics[0] == topic_issueTicket {
			nameLen := new(big.Int).SetBytes(d[end-32:]).Uint64()
			categoryName := string(d[:new(big.Int).SetBytes(d[end-32:]).Uint64()])
			match, err := regexp.Match("^[A-Za-z]{2,16}$", d[0:nameLen])
			if err != nil || !match {
				log.Trace("issueToken error ", "contract", contract.Address(), "categoryName", categoryName, "error", "illegal categoryName")
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}

			if !interpreter.evm.StateDB.RegisterTicket(contract.Address(), categoryName) {
				log.Trace("issueToken error ", "contract", contract.Address(), "categoryName", categoryName, "error", "categoryName registered by other")
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}

			pkg := assets.Package{Tkt: &assets.Ticket{
				Category: *common.BytesToHash(common.LeftPadBytes([]byte(categoryName), 32)).HashToUint256(),
				Value:    *crypto.Keccak256Hash(append([]byte(categoryName), []byte{}...)).HashToUint256(),
			},
			}
			interpreter.evm.StateDB.GetZState().AddTxOut(contract.CallerAddress, pkg)
			memory.Set(mStart.Uint64()+end-32, 32, hashTrue)
		} else if topics[0] == topic_issueToken {
			total := new(big.Int).SetBytes(d[end-32:])
			nameLen := new(big.Int).SetBytes(d[end-64 : end-32]).Uint64()

			coinName := string(d[0:nameLen])
			match, err := regexp.Match("^[A-Za-z]{2,16}$", d[0:nameLen])
			if err != nil || !match {
				log.Trace("issueToken error ", "contract", contract.Address(), "coinName", coinName, "total", total, "error", "illegal coinName")
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}

			if !interpreter.evm.StateDB.RegisterToken(contract.Address(), coinName) {
				log.Trace("issueToken error ", "contract", contract.Address(), "coinName", coinName, "total", total, "error", "coinName registered by other")
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}
			interpreter.evm.StateDB.AddBalance(contract.Address(), coinName, total)
			memory.Set(mStart.Uint64()+end-32, 32, hashTrue)
		} else if topics[0] == topic_selfBalance {
			coinName := string(d[:new(big.Int).SetBytes(d[end-32:]).Uint64()])
			balance := interpreter.evm.StateDB.GetBalance(contract.Address(), coinName)
			memory.Set(mStart.Uint64()+end-32, 32, common.LeftPadBytes(balance.Bytes(), 32))
			return balance.Bytes(), nil
		} else if topics[0] == topic_sendAnonymous {

			value := new(big.Int).SetBytes(d[end-32:])
			addr := common.BytesToContractAddress(d[end-64+12 : end-32])
			toAddr := contract.GetNonceAddress(interpreter.evm.StateDB, addr)
			if toAddr == (common.Address{}) {
				log.Trace("sendAnonymous error ", "contract", contract.Address(), "toAddr", toAddr, "error", "not load toAddrss")
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}

			len := new(big.Int).SetBytes(d[end-96 : end-64]).Uint64()
			var currency string
			if len == 0 {
				currency = "sero"
			} else {
				currency = string(d[:len])
			}

			balance := interpreter.evm.StateDB.GetBalance(contract.Address(), currency)
			if balance.Cmp(value) < 0 {
				log.Trace("sendAnonymous error ", "contract", contract.Address(), "error", "balance not enough")
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}
			gas := interpreter.evm.callGasTemp + params.CallStipend

			pkg := assets.Package{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*value),
			},
			}
			ret, returnGas, err := interpreter.evm.Call(contract, toAddr, nil, gas, pkg)

			if err != nil {
				log.Trace("sendAnonymous error ", "contract", contract.Address(), "error", err)
				memory.Set(mStart.Uint64()+end-32, 32, hashFalse)
				return hashFalse, nil
			}
			contract.Gas += returnGas
			memory.Set(mStart.Uint64()+end-32, 32, hashTrue)
			return ret, nil
		} else if topics[0] == topic_getCurrency{
			if contract.pkg.Tkn != nil {
				currency := strings.Trim(string(contract.pkg.Tkn.Currency[:]), string([]byte{0}))
				memory.Set(mStart.Uint64(), 32, common.BigToHash(big.NewInt(int64(len(currency)))).Bytes())
				memory.Set(mStart.Uint64()+32, 32, []byte(currency))
			} else {
				memory.Set(mStart.Uint64(), 32, big.NewInt(0).Bytes())
				memory.Set(mStart.Uint64(), 32, []byte{})
			}
		} else {
			interpreter.evm.StateDB.AddLog(&types.Log{
				Address: contract.Address(),
				Topics:  topics,
				Data:    d,
				// This is a non-consensus field, but assigned here because
				// core/state doesn't know the current block number.
				BlockNumber: interpreter.evm.BlockNumber.Uint64(),
			})
		}

		interpreter.intPool.put(mStart, mSize)
		return nil, nil
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
