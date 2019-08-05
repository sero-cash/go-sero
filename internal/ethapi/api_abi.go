package ethapi

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/accounts/abi"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/common/math"
	"math/big"
	"strings"
)

type PublicAbiAPI struct {

}



func decodeOutputParms(abiArgs abi.Arguments,)(interface{},bool){
	for _, arg:=range abiArgs{
		switch arg.Type.String() {
		case "address":
			return common.ContractAddress{},true
		case "address[]":
			return []common.ContractAddress{},true
		case "uint8":
			return hexutil.Uint8(0),false
		case "uint16":
			return hexutil.Uint16(0),false
		case "uint32":
			return  hexutil.Uint32(0),false
		case "uint64":
			return hexutil.Uint64(0),false
		case "uint256":
			var big hexutil.Big
			return  big,false
		case "string":
			return string(""),false
		case "string[]":
			return []string{},false
		case "bool":
			var result bool
		return result,false
		default:
			return nil, false
		}
	}
	return nil ,false
}

func decodeResult(abiArgs abi.Arguments,result interface{})(interface{}){
	for _, arg:=range abiArgs{
		switch arg.Type.String() {
		case "address":
			return result
		case "address[]":
			return result
		case "uint8":
			return hexutil.Uint64(result.(uint8))
		case "uint16":
			return hexutil.Uint64(result.(uint16))
		case "uint32":
			return  hexutil.Uint64(result.(uint32))
		case "uint64":
			return hexutil.Uint64(result.(uint64))
		case "uint256":
			return  hexutil.Big(*(result.(*big.Int)))
		case "string":
			return result
		case "string[]":
			return result
		case "bool":
			return result
		default:
			return nil
		}
	}
	return nil
}

func encodeStringParams(abiArgs abi.Arguments,args[]string)([]interface{},[]common.Address,error){

	argsLen:=len(args)
	packArgs:=make([]interface{},argsLen)
	address:=[]common.Address{}
	for index, arg:=range abiArgs{
		switch arg.Type.String() {
		case "address":
			var addr AllMixedAddress
			err:= addr.UnmarshalText([]byte(args[index]))
			if err!=nil{
				return nil, nil,err
			}
			caddr:=(keys.HashPKr(addr.ToPKr().NewRef()))
			packArgs[index]=common.BytesToContractAddress(caddr[:])
			address=append(address,common.BytesToAddress(addr[:]))
		case "address[]":
			var addrs []AllMixedAddress
			err:= json.Unmarshal([]byte(args[index]),&addrs)
			if err!=nil{
				return nil, nil,err
			}
			packArgs[index]=convertToContractAddr(addrs)
			pkrs:=convertToAddr(addrs)
			address=append(address,pkrs...)

		case "uint8":
			var num hexutil.Uint8
			err:=num.UnmarshalText([]byte(args[index]))
			if err!=nil{
				return nil,nil, err
			}
			packArgs[index]=uint8(num)
		case "uint16":
			var num hexutil.Uint16
			err:=num.UnmarshalText([]byte(args[index]))
			if err!=nil{
				return nil,nil, err
			}
			packArgs[index]=uint16(num)
		case "uint32":
			var num hexutil.Uint32
			err:=num.UnmarshalText([]byte(args[index]))
			if err!=nil{
				return nil,nil, err
			}
			packArgs[index]=uint32(num)
		case "uint64":
			var num hexutil.Uint64
			err:=num.UnmarshalText([]byte(args[index]))
			if err!=nil{
				return nil,nil, err
			}
			packArgs[index]=uint64(num)
		case "uint256":
			var num hexutil.Big
			err:=num.UnmarshalText([]byte(args[index]))
			if err!=nil{
				return nil,nil, err
			}
			packArgs[index]=num.ToInt()
		case "string":
			packArgs[index]=args[index]
		case "string[]":
			strs:=[]string{}
			value := strings.Replace(strings.Replace(args[index], "[", "", 1), "]", "", 1)
			values := strings.Split(value, ",")
			for _, vv := range values {
				strs = append(strs, vv)
			}
			packArgs[index]=strs

		case "bool":
			switch args[index]{
			case "true":
				packArgs[index]=true
			case  "false":
				packArgs[index]=false
			default:
				return nil,nil,errors.New("error args !")
			}

		default:
			return nil,nil, fmt.Errorf("unsupported arg type")
		}

	}

	return packArgs[:],address,nil
}


func PackMethod(abi *abi.ABI,contractAddr ContractAddress,methodName string,args []string)(hexutil.Bytes,error){
	if abi==nil {
		return hexutil.Bytes{},errors.New("ABI can not be nil")
	}

	method, exist := abi.Methods[methodName]
	if !exist {
		return nil, fmt.Errorf("method '%s' not found", methodName)
	}
	abiArgs:=method.Inputs
	if len(abiArgs)!= len(args){
		return nil, fmt.Errorf("argument count mismatch: %d for %d", len(args), len(abiArgs))
	}
	packArgs,address,err:= encodeStringParams(abiArgs,args)
	input,err:=abi.Pack(methodName,packArgs...)
	if err!=nil {
		return hexutil.Bytes{},err
	}

	prefix := [18]byte{}
	addressLen:= len(address)
	copy(prefix[:], contractAddr[:16])
	if len(address)>1000{
		return nil ,errors.New("too many address args")
	}
	lenBytes:=math.PaddedBigBytes(big.NewInt(int64(addressLen)), 2)
	copy(prefix[16:],lenBytes[:])
	result :=[]byte{}
	result=append(result,prefix[:]...)
	for _,addr:=range address{
		result=append(result,addr[:]...)
	}
	result=append(result,input[:]...)
	return hexutil.Bytes(result[:]) ,nil
}

func (s *PublicAbiAPI) PackMethod(abi *abi.ABI,contractAddr ContractAddress,methodName string,args []string)(hexutil.Bytes,error){
	return PackMethod(abi,contractAddr,methodName,args)
}


func convertToContractAddr(addrs[] AllMixedAddress)(result []common.ContractAddress){
    for _,addr:=range addrs{
		caddr:=(keys.HashPKr(addr.ToPKr().NewRef()))
    	result =append(result,common.BytesToContractAddress(caddr[:]))
	}
    return
}


func convertToAddr(addrs[] AllMixedAddress)(result []common.Address){
	for _,addr:=range addrs{
		result =append(result,common.BytesToAddress(addr[:]))
	}
	return
}


func PackConstruct(abi *abi.ABI,data hexutil.Bytes,args []string)(hexutil.Bytes,error){
	if abi==nil {
		return hexutil.Bytes{},errors.New("ABI can not be nil")
	}

	abiArgs:=abi.Constructor.Inputs
	if len(abiArgs)!= len(args){
		return nil, fmt.Errorf("argument count mismatch: %d for %d", len(args), len(abiArgs))
	}
	packArgs,address,err:= encodeStringParams(abiArgs,args)
	if err!=nil {
		return hexutil.Bytes{},err
	}
	input,err:=abi.Pack("",packArgs...)
	if err!=nil {
		return hexutil.Bytes{},err
	}

	prefix := [18]byte{}
	addressLen:= len(address)
	rand:=keys.RandUint128()
	copy(prefix[:], rand[:])
	if len(address)>1000{
		return nil ,errors.New("too many address args")
	}
	lenBytes:=math.PaddedBigBytes(big.NewInt(int64(addressLen)), 2)
	copy(prefix[16:],lenBytes[:])
	result :=[]byte{}
	result=append(result,prefix[:]...)
	for _,addr:=range address{
		result=append(result,addr[:]...)
	}
	fmt.Printf(hexutil.Encode(data))
	result=append(result,data[:]...)
	result=append(result,input[:]...)
	return hexutil.Bytes(result[:]) ,nil
}


func (s *PublicAbiAPI) PackConstruct(abi *abi.ABI,data hexutil.Bytes,args []string)(hexutil.Bytes,error){

   return PackConstruct(abi,data,args)

}

func UnPack(abi *abi.ABI,name string,output hexutil.Bytes)(interface{},error){
	if abi==nil {
		return hexutil.Bytes{},errors.New("ABI can not be nil")
	}
	if method, ok := abi.Methods[name]; ok {
		marshalledValues, err := method.Outputs.UnpackValues(output)
		if err != nil {
			return nil, err
		}
		return marshalledValues,nil
	}else{
		return nil,errors.New("can not find method="+name)
	}
	return nil ,nil
}


func (s *PublicAbiAPI) UnPack(abi *abi.ABI,name string,output hexutil.Bytes)(interface{},error){
	return UnPack(abi,name,output)
}
