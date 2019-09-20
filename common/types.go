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
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-sero/common/addrutil"

	"github.com/sero-cash/go-sero/crypto/sha3"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/common/hexutil"
)

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the adddress
	AddressLength = 96
)

var (
	hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(Address{})
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// BigToHash sets byte representation of b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

// HexToHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToHash(s string) Hash { return BytesToHash(FromHex(s)) }

// Bytes gets the byte representation of the underlying hash.
func (h Hash) Bytes() []byte { return h[:] }

// Big converts a hash to a big integer.
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }

// Hex converts a hash to a hex string.
func (h Hash) Hex() string { return hexutil.Encode(h[:]) }

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (h Hash) TerminalString() string {
	return fmt.Sprintf("%x…%x", h[:3], h[29:])
}

// String implements the stringer interface and is used also by the logger when
// doing full logging into a file.
func (h Hash) String() string {
	return h.Hex()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (h Hash) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), h[:])
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Hash", input, h[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (h *Hash) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(hashT, input, h[:])
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

// Scan implements Scanner for database/sql.
func (h *Hash) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into Hash", src)
	}
	if len(srcB) != HashLength {
		return fmt.Errorf("can't scan []byte of len %d into Hash, want %d", len(srcB), HashLength)
	}
	copy(h[:], srcB)
	return nil
}

// Value implements valuer for database/sql.
func (h Hash) Value() (driver.Value, error) {
	return h[:], nil
}

func (h Hash) HashToUint256() *c_type.Uint256 {
	u256 := c_type.Uint256{}
	copy(u256[:], h[:])
	return &u256
}

func HashToHex(hashs []Hash) []string {
	hexs := []string{}
	for _, hash := range hashs {
		hexs = append(hexs, hash.Hex())
	}
	return hexs
}

// UnprefixedHash allows marshaling a Hash without 0x prefix.
type UnprefixedHash Hash

// UnmarshalText decodes the hash from hex. The 0x prefix is optional.
func (h *UnprefixedHash) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedUnprefixedText("UnprefixedHash", input, h[:])
}

// MarshalText encodes the hash as hex.
func (h UnprefixedHash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

/////////// Data

func keccak512(data ...[]byte) []byte {
	d := sha3.NewKeccak512()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Data represents the 64 byte Data of an Ethereum account.
type Address [AddressLength]byte

type ContractAddress [20]byte

func (a Address) ToCaddr() ContractAddress {
	var addr ContractAddress
	pkr := new(c_type.PKr)
	copy(pkr[:], a[:])
	hash := superzk.HashPKr(pkr)
	addr.SetBytes(hash[:])
	return addr
}

// Hash converts an Data to a hash by left-padding it with zeros.
func (a ContractAddress) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

func (a *ContractAddress) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-20:]
	}
	copy(a[20-len(b):], b)
}

func (a ContractAddress) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (a *ContractAddress) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("ContractAddress", input, a[:])
}

func BytesToString(b []byte) string {
	return strings.Trim(string(b), string([]byte{0}))
}

// BytesToAddress returns Data with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func BytesToContractAddress(b []byte) ContractAddress {
	var a ContractAddress
	a.SetBytes(b)
	return a
}

// BigToAddress returns Data with byte values of b.
// If b is larger than len(h), b will be cropped from the left.
func BigToAddress(b *big.Int) Address { return BytesToAddress(b.Bytes()) }

func BigToContractAddress(b *big.Int) ContractAddress { return BytesToContractAddress(b.Bytes()) }

// HexToAddress returns Data with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
//func HexToAddress(s string) Data { return BytesToAddress(FromHex(s)) }

func Base58ToAddress(s string) Address {
	out := base58.Decode(s)
	return BytesToAddress(out[:])
}

// Bytes gets the string representation of the underlying Data.
func (a Address) Bytes() []byte { return a[:] }

func (a Address) ToPKr() *c_type.PKr {
	pubKey := c_type.PKr{}
	copy(pubKey[:], a[:])
	return &pubKey
}

func (a Address) ToUint512() *c_type.Uint512 {
	pubKey := c_type.Uint512{}
	copy(pubKey[:], a[:])
	return &pubKey
}

// Big converts an Data to a big integer.
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// Base58 returns base58 string representation of the Data.
func (a Address) Base58() string {
	return base58.Encode(a[:])
}

// String implements fmt.Stringer.
func (a Address) String() string {
	if a.IsAccountAddress() {
		return base58.Encode(a[:64])
	} else {
		return a.Base58()
	}
}

func (a Address) IsAccountAddress() bool {
	zerobytes := [32]byte{}
	var suffix [32]byte
	copy(suffix[:], a[64:])
	if suffix == zerobytes {
		return true
	}
	return false
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a Address) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "%"+string(c), a[:])
}

// SetBytes sets the Data to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	copy(a[:], b)
	//if len(b) > len(a) {
	//	b = b[len(b)-AddressLength:]
	//}
	//copy(a[AddressLength-len(b):], b)
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalBase58Text()
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	out, err := addrutil.IsValidBase58Address(input)
	if err != nil {
		return err
	}
	copy(a[:], out)
	return nil
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	if !addrutil.IsString(input) {
		return errors.New("not string")
	} else {
		return a.UnmarshalText(input[1 : len(input)-1])

	}
}

// Scan implements Scanner for database/sql.
func (a *Address) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into Data", src)
	}
	if len(srcB) != AddressLength {
		return fmt.Errorf("can't scan []byte of len %d into Data, want %d", len(srcB), AddressLength)
	}
	copy(a[:], srcB)
	return nil
}

//func (a *Data) IsContract() bool {
//	return strings.HasSuffix(string(a[:]),"contract")
//}

// Value implements valuer for database/sql.
func (a Address) Value() (driver.Value, error) {
	return a[:], nil
}

type AddressList []Address

func (self AddressList) Len() int {
	return len(self)
}
func (self AddressList) Less(i, j int) bool {
	return bytes.Compare(self[i][:], self[j][:]) < 0
}
func (self AddressList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

// UnprefixedAddress allows marshaling an Data without 0x prefix.
type UnprefixedAddress Address

// UnmarshalText decodes the Data from hex. The 0x prefix is optional.
func (a *UnprefixedAddress) UnmarshalText(input []byte) error {
	if addrutil.IsBase58Str(string(input)) {
		out := base58.Decode(string(input))
		copy(a[:], out[:])
		return nil
	} else {
		return errors.New("invalid Address")
	}
}

// MarshalText encodes the Data as hex.
func (a UnprefixedAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(a[:])), nil
}

func ByteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
