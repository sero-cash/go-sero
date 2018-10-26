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

package stx

import (
	"encoding/hex"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/crypto/sha3"
)

type Out_O struct {
	Addr keys.Uint512
	Pkg  assets.Package
	Memo keys.Uint512
}

func (self *Out_O) ToHash() (ret keys.Uint256) {
	hash := crypto.Keccak256(
		self.Addr[:],
		self.Pkg.ToHash().NewRef()[:],
		self.Memo[:],
	)
	copy(ret[:], hash)
	return ret
}

type In_O struct {
	Root keys.Uint256
	Sign keys.Uint256
}

func (ino In_O) MarshalText() ([]byte, error) {
	input := [64]byte{}
	copy(input[:32], ino.Root[:])
	copy(input[32:], ino.Sign[:])
	result := make([]byte, len(input)*2+2)
	copy(result, `0x`)
	hex.Encode(result[2:], input[:])
	return result, nil
}

func (self *In_O) ToHash_for_z() (ret keys.Uint256) {
	copy(ret[:], self.Root[:])
	return ret
}

type Desc_O struct {
	Ins  []In_O
	Outs []Out_O
}

func (self *Desc_O) ToHash_for_z() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash_for_z().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}
