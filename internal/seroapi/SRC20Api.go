package seroapi

import (
	"strings"

	"github.com/sero-cash/go-sero/accounts/abi"

	"github.com/sero-cash/go-sero/common/address"
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
	contractAddress address.AccountAddress
	tokenName       string
	definition      string
	method          string
}

func NewSRC20Decimal(contractAddress address.AccountAddress, tokenName string) []SRC20Decimal {
	result := []SRC20Decimal{
		{contractAddress, tokenName, decimalsDefinition, "decimals"},
		{contractAddress, tokenName, getDecimalByNameDefinition, "getDecimal"},
		{contractAddress, tokenName, getDecimalDefinition, "getDecimal"},
	}
	return result
}

type SRCAbi interface {
	Pack() ([]byte, error)
	Unpack(outData []byte) (*uint8, error)
}

func packDecimalData(contractAddess address.AccountAddress, out []byte) []byte {
	prefix := [18]byte{}
	copy(prefix[:], contractAddess[:16])
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
	return packDecimalData(d.contractAddress, out[:]), nil

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
