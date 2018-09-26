package abi

import (
	"encoding/json"
	"fmt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"

	"math/big"
	"reflect"
	"strings"
)

func ValueTo(typ Type, v interface{}, r *keys.Uint128, state *state.StateDB) (val reflect.Value, addrs []common.Address) {

	switch typ.T {
	case SliceTy:
		val := reflect.MakeSlice(typ.Type, 0, 0)
		fmt.Println("before : ", val)

		for _, t := range v.([]interface{}) {
			res, addresss := ValueTo(*typ.Elem, t, r, state)
			addrs = append(addrs, addresss...)
			val = reflect.Append(val, res)
		}
		fmt.Println("after : ", val)
		return val, addrs
	case ArrayTy:
		val := reflect.MakeSlice(reflect.SliceOf(typ.Elem.Type), 0, 0)
		fmt.Println("before : ", val)
		for _, t := range v.([]interface{}) {
			res, addresss := ValueTo(*typ.Elem, t, r, state)
			addrs = append(addrs, addresss...)
			val = reflect.Append(val, res)
		}
		fmt.Println("after : ", val)
		arr := reflect.ValueOf(reflect.New(typ.Type).Interface()).Elem()
		reflect.Copy(arr, val)
		return arr, addrs
	case UintTy:
		_, ok := v.(string)
		var numStr string
		if ok {
			numStr = v.(string)
		} else {
			numStr = string(v.(json.Number))
		}
		if typ.Size == 8 || typ.Size == 16 || typ.Size == 32 || typ.Size == 64 {
			elem := reflect.New(typ.Type).Elem()
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			elem.SetUint(newInt.Uint64())
			return elem, nil

		} else {
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			return reflect.ValueOf(newInt), nil
		}
	case IntTy:
		_, ok := v.(string)
		var numStr string
		if ok {
			numStr = v.(string)
		} else {
			numStr = string(v.(json.Number))
		}
		if typ.Size == 8 || typ.Size == 16 || typ.Size == 32 || typ.Size == 64 {
			elem := reflect.New(typ.Type).Elem()
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			elem.SetInt(newInt.Int64())
			return elem, nil
		} else {
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			return reflect.ValueOf(newInt), nil
		}
	case BoolTy:
		return reflect.ValueOf(v.(bool)), nil
	case AddressTy:
		address := common.Base58ToAddress(v.(string))
		if state.IsContract(address) {
			return reflect.ValueOf(address.ToCaddr()), []common.Address{address}
		} else {
			pkr := keys.Addr2PKr(address.ToUint512(), r.ToUint256().NewRef())
			onceAddr := common.Address{}
			onceAddr.SetBytes(pkr[:])
			return reflect.ValueOf(onceAddr.ToCaddr()), []common.Address{onceAddr}
		}
	case StringTy:
		return reflect.ValueOf(v.(string)), nil
	case FixedBytesTy:
		elem := reflect.New(typ.Type).Elem()
		for i, t := range v.([]interface{}) {
			v, _ := t.(json.Number).Int64()
			elem.Index(i).Set(reflect.ValueOf(byte(v)))
		}
		return elem, nil
	case BytesTy:
		elem := reflect.New(typ.Type).Elem()
		buffer := []byte{}
		for _, t := range v.([]interface{}) {
			i, _ := t.(json.Number).Int64()
			buffer = append(buffer, byte(i))
		}
		elem.SetBytes(buffer)
		return elem, nil
	default:
		panic(v)
	}
}

func getArgs(pairs []string, r *keys.Uint128, state *state.StateDB) ([]interface{}, []common.Address, error) {
	inputs := []interface{}{}
	address := []common.Address{}
	for _, line := range pairs {
		vs := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		if err := dec.Decode(&vs); err != nil {
			return nil, nil, err
		}
		for k, v := range vs {
			//fmt.Printf("%t\n %t\n", k, v)
			typeT, _ := NewType(k)
			tv, addrs := ValueTo(typeT, v, r, state)
			address = append(address, addrs...)
			inputs = append(inputs, tv.Interface())
		}

	}
	return inputs, address, nil
}

func (abi ABI) PrefixCreatePack(input []byte, pairs []string, r *keys.Uint128, state *state.StateDB) ([]byte, error) {

	args, address, err := getArgs(pairs, r, state)

	if err != nil {
		return nil, err
	}

	prefix := common.LeftPadBytes(big.NewInt(int64(len(address))).Bytes(), 2)
	for _, addr := range address {
		prefix = append(prefix, addr.Bytes()...)
	}
	// constructor
	arguments, err := abi.Constructor.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	return append(prefix, append(input, arguments...)...), nil

}

func (abi ABI) PrefixPack(input []byte, pairs []string, r *keys.Uint128, state *state.StateDB) ([]byte, error) {

	args, address, err := getArgs(pairs, r, state)

	if err != nil {
		return nil, err
	}

	prefix := common.LeftPadBytes(big.NewInt(int64(len(address))).Bytes(), 2)
	for _, addr := range address {
		prefix = append(prefix, addr.Bytes()...)
	}
	sigdata := input[:4]
	method, err := abi.MethodById(sigdata)
	if err != nil {
		return input, nil
	}

	arguments, err := method.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}
	// Pack up the method ID too if not a constructor and return
	return append(prefix, append(method.Id(), arguments...)...), nil
}

func (abi ABI) Transfer(input []byte, output []byte, state *state.StateDB, contractAddress common.Address, wallets []accounts.Wallet) ([]byte, error) {

	if len(output) == 0 {
		return output, nil
	}

	sigdata := input[:4]
	method, err := abi.MethodById(sigdata)
	if err != nil {
		return output, nil
	}

	args, err := method.Outputs.TranserApiValues(output, state, contractAddress, wallets)

	if err != nil {
		return output, err
	}

	return args, nil

}

func forEachArgs(t Type, output []byte, start, size int, state *state.StateDB, contractAddress common.Address, wallets []accounts.Wallet) ([]byte, error) {
	if size < 0 {
		return nil, fmt.Errorf("cannot marshal input to array, size is negative (%d)", size)
	}
	if start+32*size > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go array: offset %d would go over slice boundary (len=%d)", len(output), start+32*size)
	}

	// this value will become our slice or our array, depending on the type
	var refSlice reflect.Value

	if t.T == SliceTy {
		// declare our slice
		refSlice = reflect.MakeSlice(t.Type, size, size)
	} else if t.T == ArrayTy {
		// declare our array
		refSlice = reflect.New(t.Type).Elem()
	} else {
		return nil, fmt.Errorf("abi: invalid type in array/slice unpacking stage")
	}

	// Arrays have packed elements, resulting in longer unpack steps.
	// Slices have just 32 bytes per element (pointing to the contents).
	elemSize := 32
	if t.T == ArrayTy {
		elemSize = getFullElemSize(t.Elem)
	}

	for i, j := start, 0; j < size; i, j = i+elemSize, j+1 {

		inter, err := toAPIOut(i, *t.Elem, output, state, contractAddress, wallets)
		if err != nil {
			return nil, err
		}

		// append the item to our reflect slice
		refSlice.Index(j).Set(reflect.ValueOf(inter))
	}

	// return the interface
	return refSlice.Bytes(), nil
}

func toAPIOut(index int, t Type, output []byte, state *state.StateDB, contractAddress common.Address, wallets []accounts.Wallet) ([]byte, error) {
	if index+32 > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %d require %d", len(output), index+32)
	}

	var (
		returnOutput []byte
		begin, end   int
		err          error
	)

	// if we require a length prefix, find the beginning word and size returned.
	if t.requiresLengthPrefix() {
		begin, end, err = lengthPrefixPointsTo(index, output)
		if err != nil {
			return nil, err
		}
	} else {
		returnOutput = output[index : index+32]
	}

	switch t.T {
	case SliceTy:
		return forEachArgs(t, output, begin, end, state, contractAddress, wallets)
	case ArrayTy:
		return forEachArgs(t, output, index, t.Size, state, contractAddress, wallets)
	case AddressTy:
		once := getMine(wallets, state.GetNonceAddress(returnOutput[12:]))
		return once.Bytes(), nil
	default:
		return output, nil
	}
}

func getMine(wallets []accounts.Wallet, once common.Address) common.Address {
	if len(wallets) == 0 {
		return once
	}
	for _, wallet := range wallets {
		if wallet.IsMine(once) {
			return wallet.Accounts()[0].Address
		}
	}
	return once

}

func (arguments Arguments) TranserApiValues(data []byte, state *state.StateDB, contractAddress common.Address, wallets []accounts.Wallet) ([]byte, error) {
	retval := make([]byte, 0, arguments.LengthNonIndexed())
	virtualArgs := 0
	for index, arg := range arguments.NonIndexed() {
		marshalledValue, err := toAPIOut((index+virtualArgs)*32, arg.Type, data, state, contractAddress, wallets)
		if arg.Type.T == ArrayTy {
			// If we have a static array, like [3]uint256, these are coded as
			// just like uint256,uint256,uint256.
			// This means that we need to add two 'virtual' arguments when
			// we count the index from now on.
			//
			// Array values nested multiple levels deep are also encoded inline:
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// Calculate the full array size to get the correct offset for the next argument.
			// Decrement it by 1, as the normal index increment is still applied.
			virtualArgs += getArraySize(&arg.Type) - 1
		}
		if err != nil {
			return nil, err
		}
		retval = append(retval, marshalledValue...)
	}
	return retval, nil
}
