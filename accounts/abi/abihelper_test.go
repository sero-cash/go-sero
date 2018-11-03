package abi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/sero-cash/go-sero/common/hexutil"

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

func verify(t *testing.T, jsondata, calldata string, exp string) {

	abispec, err := JSON(strings.NewReader(jsondata))
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range abispec.Methods {
		x, _ := hexutil.Bytes(m.Id()).MarshalText()
		fmt.Printf("\n %s", x)
	}

	cd := common.Hex2Bytes(calldata)
	sigdata, argdata := cd[:4], cd[4:]
	method, err := abispec.MethodById(sigdata)

	if err != nil {
		t.Fatal(err)
	}

	//data, err := method.Outputs.UnpackValues(argdata)
	outd, err := method.Outputs.TranserApiValues(argdata, nil, common.Address{}, nil)

	data, _ := hexutil.Bytes(outd[:]).MarshalText()
	if string(data) != exp {
		t.Fatalf("Mismatched length, \nexpected %s, \ngot %s", exp, string(data))
	}

}
func TestNewUnpacker(t *testing.T) {
	type unpackTest struct {
		jsondata string
		calldata string
		exp      string
	}
	testcases := []unpackTest{

		{ // https://solidity.readthedocs.io/en/develop/abi-spec.html#use-of-dynamic-types
			`[{"type":"function","name":"f", "outputs":[{"type":"uint256"},{"type":"uint32[]"},{"type":"bytes10"},{"type":"bytes"}]}]`,
			// 0x123, [0x456, 0x789], "1234567890", "Hello, world!"
			"26121ff0" + "00000000000000000000000000000000000000000000000000000000000001230000000000000000000000000000000000000000000000000000000000000080313233343536373839300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000004560000000000000000000000000000000000000000000000000000000000000789000000000000000000000000000000000000000000000000000000000000000d48656c6c6f2c20776f726c642100000000000000000000000000000000000000",
			"00000000000000000000000000000000000000000000000000000000000001230000000000000000000000000000000000000000000000000000000000000080313233343536373839300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000004560000000000000000000000000000000000000000000000000000000000000789000000000000000000000000000000000000000000000000000000000000000d48656c6c6f2c20776f726c642100000000000000000000000000000000000000",
		}, { // https://github.com/ethereum/wiki/wiki/Ethereum-Contract-ABI#examples
			`[{"type":"function","name":"sam","outputs":[{"type":"bytes"},{"type":"bool"},{"type":"uint256[]"}]}]`,
			//  "dave", true and [1,2,3]
			"7edba6c80000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003",
			"a5643bf20000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000464617665000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003",
		}, {
			`[{"type":"function","name":"send","outputs":[{"type":"uint256"}]}]`,
			"b46300ec0000000000000000000000000000000000000000000000000000000000000012",
			"0000000000000000000000000000000000000000000000000000000000000012",
		}, {
			`[{"type":"function","name":"compareAndApprove","outputs":[{"type":"address"},{"type":"uint256"},{"type":"uint256"}]}]`,
			"ce79fdce00000000000000000000000000000133700000deadbeef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
			"00000000000000000000000000000133700000deadbeef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
		},
	}
	for _, c := range testcases {
		verify(t, c.jsondata, c.calldata, c.exp)
	}

}
