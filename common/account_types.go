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

package common

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/base58"
	"github.com/sero-cash/go-sero/common/hexutil"
)

// Lengths of hashes and Accountes in bytes.
const (
	// AccountAddressLength is the expected length of the adddress
	AccountAddressLength = 64
	SeedLength           = 32
)

// Data represents the 64 byte Data of an Ethereum account.
type AccountAddress [AccountAddressLength]byte

// If b is larger than len(h), b will be cropped from the left.
func BytesToAccount(b []byte) AccountAddress {
	var a AccountAddress
	a.SetBytes(b)
	return a
}

// BigToAccount returns Data with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAccount(b *big.Int) AccountAddress { return BytesToAccount(b.Bytes()) }

// HexToAccount returns Data with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
//func HexToAccount(s string) Data { return BytesToAccount(FromHex(s)) }

func Base58ToAccount(s string) AccountAddress {
	out := [AccountAddressLength]byte{}
	FromBase58(s, out[:])
	return BytesToAccount(out[:])
}

// IsBase58Account verifies whether a string can represent a valid hex-encoded
// Ethereum Data or not.
func IsBase58Account(s string) bool {
	if base58.IsBase58Str(s) {
		account := Base58ToAccount(s)
		if !keys.IsPKValid(account.ToUint512()) {
			return false
		}
		temp := Base58ToAccount(s).Base58()
		if temp == s {
			return true
		}
		return false
	}
	return false

	return base58.IsBase58Str(s)
}

// Bytes gets the string representation of the underlying Data.
func (a AccountAddress) Bytes() []byte { return a[:] }

func (a AccountAddress) ToPKr() *keys.PKr {
	pubKey := keys.PKr{}
	copy(pubKey[:], a[:])
	return &pubKey
}

func (a AccountAddress) ToUint512() *keys.Uint512 {
	pubKey := keys.Uint512{}
	copy(pubKey[:], a[:])
	return &pubKey
}

// Big converts an Data to a big integer.
func (a AccountAddress) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// Base58 returns base58 string representation of the Data.
func (a AccountAddress) Base58() string {
	return base58.EncodeToString(a[:])
}

// String implements fmt.Stringer.
func (a AccountAddress) String() string {
	return a.Base58()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a AccountAddress) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), a[:])
}

// SetBytes sets the Data to the value of b.
// If b is larger than len(a) it will panic.
func (a *AccountAddress) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AccountAddressLength:]
	}
	copy(a[AccountAddressLength-len(b):], b)
}

// MarshalText returns the hex representation of a.
func (a AccountAddress) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalBase58Text()
}

// UnmarshalText parses a hash in hex syntax.
func (a *AccountAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedBase58Text("Data", input, a[:])
}

// Scan implements Scanner for database/sql.
func (a *AccountAddress) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into Data", src)
	}
	if len(srcB) != AccountAddressLength {
		return fmt.Errorf("can't scan []byte of len %d into Data, want %d", len(srcB), AccountAddressLength)
	}
	copy(a[:], srcB)
	return nil
}

//func (a *Data) IsContract() bool {
//	return strings.HasSuffix(string(a[:]),"contract")
//}

// Value implements valuer for database/sql.
func (a AccountAddress) Value() (driver.Value, error) {
	return a[:], nil
}

type Accountes []AccountAddress

func (self Accountes) Len() int {
	return len(self)
}
func (self Accountes) Less(i, j int) bool {
	return bytes.Compare(self[i][:], self[j][:]) < 0
}
func (self Accountes) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type Seed [SeedLength]byte

func (priv *Seed) SeedToUint256() *keys.Uint256 {
	seed := keys.Uint256{}
	copy(seed[:], priv[:])
	return &seed

}
