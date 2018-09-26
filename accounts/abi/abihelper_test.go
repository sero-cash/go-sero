package abi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
)

func test_ValueTo(typ Type, v interface{}) (val reflect.Value, addrs []common.Address, err error) {

	switch typ.T {
	case SliceTy:
		if _, ok := v.([]interface{}); ok {
			val := reflect.MakeSlice(typ.Type, 0, 0)
			for _, t := range v.([]interface{}) {
				res, addresss, err := test_ValueTo(*typ.Elem, t)
				if err != nil {
					return reflect.Value{}, nil, err
				}
				addrs = append(addrs, addresss...)
				val = reflect.Append(val, res)
			}
			return val, addrs, nil
		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}

	case ArrayTy:
		if _, ok := v.([]interface{}); ok {
			val := reflect.MakeSlice(reflect.SliceOf(typ.Elem.Type), 0, 0)
			for _, t := range v.([]interface{}) {
				res, addresss, err := test_ValueTo(*typ.Elem, t)
				if err != nil {
					return reflect.Value{}, nil, err
				}
				addrs = append(addrs, addresss...)
				val = reflect.Append(val, res)
			}
			arr := reflect.ValueOf(reflect.New(typ.Type).Interface()).Elem()
			reflect.Copy(arr, val)
			return arr, addrs, nil
		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}

	case UintTy:
		var numStr string
		if _, ok := v.(string); ok {
			numStr = v.(string)
			if _, ok := big.NewInt(1).SetString(numStr, 10); !ok {

				return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
			}
		} else {
			if _, ok := v.(json.Number); ok {
				numStr = string(v.(json.Number))
			} else {
				return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
			}
		}

		if typ.Size == 8 || typ.Size == 16 || typ.Size == 32 || typ.Size == 64 {
			elem := reflect.New(typ.Type).Elem()
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			elem.SetUint(newInt.Uint64())
			return elem, nil, nil

		} else {
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			return reflect.ValueOf(newInt), nil, nil
		}
	case IntTy:
		var numStr string
		if _, ok := v.(string); ok {
			numStr = v.(string)
			if _, ok := big.NewInt(1).SetString(numStr, 10); !ok {
				return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
			}
		} else {
			if _, ok := v.(json.Number); ok {
				numStr = string(v.(json.Number))
			} else {
				return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
			}
		}
		if typ.Size == 8 || typ.Size == 16 || typ.Size == 32 || typ.Size == 64 {
			elem := reflect.New(typ.Type).Elem()
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			elem.SetInt(newInt.Int64())
			return elem, nil, nil
		} else {
			newInt := big.NewInt(1)
			newInt.SetString(numStr, 10)
			return reflect.ValueOf(newInt), nil, nil
		}
	case BoolTy:
		if _, ok := v.(bool); ok {
			return reflect.ValueOf(v.(bool)), nil, nil
		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}
	case AddressTy:
		if _, ok := v.(string); ok {
			address := common.Base58ToAddress(v.(string))
			pkr := keys.Addr2PKr(address.ToUint512(), keys.RandUint256().NewRef())
			onceAddr := common.Address{}
			onceAddr.SetBytes(pkr[:])
			return reflect.ValueOf(onceAddr.ToCaddr()), []common.Address{onceAddr}, nil

		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}

	case StringTy:
		if _, ok := v.(string); ok {
			return reflect.ValueOf(v.(string)), nil, nil
		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}
	case FixedBytesTy:
		if _, ok := v.([]interface{}); ok {
			elem := reflect.New(typ.Type).Elem()
			for i, t := range v.([]interface{}) {
				if _, ok := t.(json.Number); ok {
					v, _ := t.(json.Number).Int64()
					elem.Index(i).Set(reflect.ValueOf(byte(v)))
				} else {
					return reflect.Value{}, nil, errors.New("param type erroy,must be byte number ")
				}
			}
			return elem, nil, nil
		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}

	case BytesTy:
		if _, ok := v.([]interface{}); ok {
			elem := reflect.New(typ.Type).Elem()
			buffer := []byte{}
			for _, t := range v.([]interface{}) {
				if _, ok := t.(json.Number); ok {
					i, _ := t.(json.Number).Int64()
					buffer = append(buffer, byte(i))
				} else {
					return reflect.Value{}, nil, errors.New("param type erroy,must be byte number ")
				}
			}
			elem.SetBytes(buffer)
			return elem, nil, nil
		} else {
			return reflect.Value{}, nil, errors.New("The parameter type is wrong, please follow the ABI definition ")
		}

	default:
		return reflect.Value{}, nil, errors.New("not support abi parameter type")
	}
}

func TestABI_PrefixPack(t *testing.T) {

	testPairs := []string{
		`{"address[]":["sdfasfd","sdfsfsfd"]}`,
		`{"uint256":22342ee3}`,
		`{"uint256":"21312dddd"}`,
	}

	address := []common.Address{}
	for _, line := range testPairs {
		vs := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		if err := dec.Decode(&vs); err != nil {
			fmt.Printf("%v", err)
		}
		for k, v := range vs {
			//fmt.Printf("%t\n %t\n", k, v)
			typeT, _ := NewType(k)
			_, addrs, err := test_ValueTo(typeT, v)
			if err != nil {
				fmt.Printf("%v", err)
			}
			address = append(address, addrs...)
		}

	}

}
