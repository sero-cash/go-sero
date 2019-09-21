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

package address

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/sero-cash/go-sero/common/addrutil"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-czero-import/c_type"
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
	out := base58.Decode(s)
	return BytesToAccount(out[:])
}

// Bytes gets the string representation of the underlying Data.
func (a AccountAddress) Bytes() []byte { return a[:] }

func (a AccountAddress) ToUint512() *c_type.Uint512 {
	pubKey := c_type.Uint512{}
	copy(pubKey[:], a[:])
	return &pubKey
}

func (a AccountAddress) ToTK() *c_type.Tk {
	tk := c_type.Tk{}
	copy(tk[:], a[:])
	return &tk
}

// Big converts an Data to a big integer.
func (a AccountAddress) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// Base58 returns base58 string representation of the Data.
func (a AccountAddress) Base58() string {
	return base58.Encode(a[:])
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

func isZeroSuffix(base58bytes [96]byte) bool {
	zerobytes := [32]byte{}
	suffix := [32]byte{}
	copy(suffix[:], base58bytes[64:])
	return (zerobytes == suffix)
}

// UnmarshalText parses a hash in hex syntax.
func (a *AccountAddress) UnmarshalText(input []byte) error {
	out, err := addrutil.IsValidBase58AcccountAddress(input)
	if err != nil {
		return err
	}
	copy(a[:], out)
	return nil
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

func (priv *Seed) SeedToUint256() *c_type.Uint256 {
	seed := c_type.Uint256{}
	copy(seed[:], priv[:])
	return &seed

}
