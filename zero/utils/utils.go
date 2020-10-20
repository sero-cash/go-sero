// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"strings"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common"
)

func Uint64ToBytes(r uint64) []byte {
	value := new(big.Int).SetUint64(r)
	return value.Bytes()
}
func Int64ToBytes(r int64) []byte {
	value := new(big.Int).SetInt64(r)
	return value.Bytes()
}
func Uint256SliceCut(is []c_type.Uint256, l int) (ret []c_type.Uint256) {
	is_l := len(is)
	if is_l < l {
		l = is_l
	}
	ret = is[:l]
	return
}

type Uint256s []c_type.Uint256

func (self Uint256s) Len() int {
	return len(self)
}
func (self Uint256s) Less(i, j int) bool {
	return bytes.Compare(self[i][:], self[j][:]) < 0
}
func (self Uint256s) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func CurrencyToUint256(str string) (ret c_type.Uint256) {
	bs := CurrencyToBytes(str)
	copy(ret[:], bs)
	return
}

func Uint256ToCurrency(u *c_type.Uint256) (ret string) {
	return BytesToCurrency(u[:])
}

func CurrencyToBytes(currency string) []byte {
	return common.LeftPadBytes([]byte(strings.ToUpper(currency)), 32)
}

func BytesToCurrency(bs []byte) string {
	return common.BytesToString(bs)
}

func ShowStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	fmt.Printf("==> %s\n", string(buf[:n]))
}

func DecodeNumber32(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return binary.BigEndian.Uint32(data)
}

func EncodeNumber32(number uint32) []byte {
	enc := make([]byte, 4)
	binary.BigEndian.PutUint32(enc, number)
	return enc
}

func EncodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func DecodeNumber(data []byte) uint64 {
	if len(data) == 0 {
		return 0
	}
	return binary.BigEndian.Uint64(data)
}

var (
//sero = 1000000000000000000
//gta  = big.NewInt(1000000000)
//ta   = big.NewInt(1)
)

func ParseAmount(s string) (*big.Int, error) {
	s = strings.ToUpper(s)
	base := 1.0
	if strings.HasSuffix(s, "SERO") {
		s = s[0 : len(s)-4]
		base = 1000000000000000000.0
	} else if strings.HasSuffix(s, "GTA") {
		s = s[0 : len(s)-3]
		base = 1000000000.0
	} else if strings.HasSuffix(s, "TA") {
		s = s[0 : len(s)-2]
	}
	if valFloat, ok := new(big.Float).SetString(s); ok {
		valFloat = new(big.Float).Mul(valFloat, big.NewFloat(base))
		ret, _ := valFloat.Int(nil)
		return ret, nil
	}

	return nil, errors.New("illegal args")
}
