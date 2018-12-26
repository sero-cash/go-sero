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
	"encoding/gob"
	"fmt"
	"math/big"
	"strings"

	"github.com/sero-cash/go-czero-import/keys"
)

func Uint64ToBytes(r uint64) []byte {
	value := new(big.Int).SetUint64(r)
	return value.Bytes()
}
func Int64ToBytes(r int64) []byte {
	value := new(big.Int).SetInt64(r)
	return value.Bytes()
}
func Uint256SliceCut(is []keys.Uint256, l int) (ret []keys.Uint256) {
	is_l := len(is)
	if is_l < l {
		l = is_l
	}
	ret = is[:l]
	return
}

func DeepCopy(dst, src interface{}) {
	//deepcopy.Copy(dst, src)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		panic(fmt.Sprintf("deepCopy encode error for : %v", src))
	}
	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst); err != nil {
		panic(fmt.Sprintf("deepCopy decode error for : %v", src))
	}
}

type Uint256s []keys.Uint256

func (self Uint256s) Len() int {
	return len(self)
}
func (self Uint256s) Less(i, j int) bool {
	return bytes.Compare(self[i][:], self[j][:]) < 0
}
func (self Uint256s) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func StringToUint256(str string) keys.Uint256 {
	var ret keys.Uint256
	b := []byte(strings.ToUpper(str))
	if len(b) > len(ret) {
		b = b[len(b)-len(ret):]
	}
	copy(ret[len(ret)-len(b):], b)
	return ret

}
