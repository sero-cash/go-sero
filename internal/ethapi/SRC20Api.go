package ethapi

import (
	"strings"

	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/accounts/abi"
)

const getDecimalByNameDefinition = `[{
	"constant": false,
	"inputs": [{
		"name": "name",
		"type": "string"
	}],
	"name": "getDecimal",
	"outputs": [{
		"name": "",
		"type": "uint8"
	}],
	"payable": false,
	"type": "function"
	}]`
const getDecimalDefinition = `[{
	"constant": true,
	"inputs": [],
	"name": "getDecimal",
	"outputs": [{
		"name": "",
		"type": "uint8"
	}],
	"payable": false,
	"type": "function"
    }]`

const decimalsDefinition = `[{
	"constant": true,
	"inputs": [],
	"name": "decimals",
	"outputs": [{
		"name": "",
		"type": "uint8"
	}],
	"payable": false,
	"type": "function"
    }]`

type SRC20Decimal struct {
	tokenName  string
	definition string
	method     string
}

func NewSRC20Decimal(tokenName string) []SRC20Decimal {
	result := []SRC20Decimal{
		{tokenName, decimalsDefinition, "decimals"},
		{tokenName, getDecimalByNameDefinition, "getDecimal"},
		{tokenName, getDecimalDefinition, "getDecimal"},
	}
	return result
}

type SRCAbi interface {
	Pack() ([]byte, error)
	Unpack(outData []byte) (*uint8, error)
}

func packDecimalData(out []byte) []byte {
	prefix := [18]byte{}
	rand := c_type.RandUint128()
	copy(prefix[:], rand[:])
	l := 18 + len(out)
	result := make([]byte, l)
	copy(result[:18], prefix[:])
	copy(result[18:], out[:])
	return result[:]
}

func (d SRC20Decimal) Pack() ([]byte, error) {
	abi, err := abi.JSON(strings.NewReader(d.definition))
	if err != nil {
		return []byte{}, err
	}
	method := abi.Methods[d.method]
	var out []byte
	if len(method.Inputs) > 0 {
		out, err = abi.Pack(d.method, d.tokenName)
	} else {
		out, err = abi.Pack(d.method)
	}
	if err != nil {
		return []byte{}, err
	}
	return packDecimalData(out[:]), nil

}

func (d SRC20Decimal) Unpack(outData []byte) (*uint8, error) {
	abi, err := abi.JSON(strings.NewReader(d.definition))
	if err != nil {
		return nil, err
	}
	var decimal uint8
	err = abi.Unpack(&decimal, d.method, outData)
	if err != nil {
		return nil, err
	}
	return &decimal, nil

}
