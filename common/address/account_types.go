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
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"
)

// Lengths of hashes and Accountes in bytes.
const (
	// AccountAddressLength is the expected length of the adddress
	AccountAddressLength = 64
	SeedLength           = 32
)

type Seed [SeedLength]byte

func (priv *Seed) SeedToUint256() *c_type.Uint256 {
	seed := c_type.Uint256{}
	copy(seed[:], priv[:])
	return &seed

}

func IsBase58Str(s string) bool {

	pattern := "^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$"
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match

}

type MixBase58Adrress []byte

func (b MixBase58Adrress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b)), nil
}

func (b MixBase58Adrress) IsPkr() bool {
	return len(b) == 96
}

func (b MixBase58Adrress) ToPkr() c_type.PKr {
	var pkr c_type.PKr
	if b.IsPkr() {
		copy(pkr[:], b[:])
	} else {
		var pk c_type.Uint512
		copy(pk[:], b[:])
		pkr = superzk.Pk2PKr(&pk, nil)
	}
	return pkr
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *MixBase58Adrress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return errors.New("empty string")
	}
	if IsBase58Str(string(input)) {
		out := base58.Decode(string(input))
		if len(out) == 96 {
			err := ValidPkr(out)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
		} else if len(out) == 64 {
			err := ValidPk(out)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
		} else {
			return errors.New("invalid mix address")
		}
	} else {
		return errors.New("is not base58 address")
	}
}

type TKAddress [64]byte

func Base58ToTk(str string) (ret TKAddress) {
	b := base58.Decode(str)
	copy(ret[:], b)
	return
}

func (b TKAddress) ToTk() c_type.Tk {
	result := c_type.Tk{}
	copy(result[:], b[:])

	return result
}

func (b *TKAddress) ToPk() (ret PKAddress) {
	pk, _ := superzk.Tk2Pk(b.ToTk().NewRef())
	copy(ret[:], pk[:])
	return
}

func (c TKAddress) String() string {
	return base58.Encode(c[:])
}

func (b TKAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b[:])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *TKAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return nil
	}
	if IsBase58Str(string(input)) {
		out := base58.Decode(string(input))
		if len(out) == 64 {
			copy(b[:], out)
		} else {
			return errors.New("ivalid TK")
		}
		return nil

	} else {
		return errors.New("is not base58 string")
	}
}

type PKAddress [64]byte

func StringToPk(str string) (ret PKAddress) {
	out := base58.Decode(str)
	copy(ret[:], out)
	return
}

func (b PKAddress) String() string {
	return base58.Encode(b[:])
}

func (b PKAddress) ToUint512() c_type.Uint512 {
	result := c_type.Uint512{}
	copy(result[:], b[:])

	return result
}

func NewPKAddres(b []byte) (ret PKAddress) {
	copy(ret[:], b)
	return
}

func (b PKAddress) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKAddress) UnmarshalText(input []byte) (e error) {
	if len(input) == 0 {
		return nil
	}
	var out []byte
	if IsBase58Str(string(input)) {
		out = base58.Decode(string(input))

	} else if IsHex(string(input)) {
		out, e = DecodeHex(string(input))
		if e != nil {
			return
		}

	} else {
		return errors.New("invalid pk string")
	}
	if len(out) == 64 {
		e = ValidPk(out)
		if e != nil {
			return e
		}
		copy(b[:], out)
		return
	} else {
		return errors.New("pk address must be 64 bytes")
	}
}

func ValidPk(addr []byte) error {
	if len(addr) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], addr)
		if !superzk.IsPKValid(&pk) {
			return errors.New("invalid PK")
		}
	} else {
		return errors.New("pk address must be 64 bytes")
	}
	return nil
}

func ValidPkr(addr []byte) error {
	if len(addr) == 96 {
		var pkr c_type.PKr
		copy(pkr[:], addr)
		if !superzk.IsPKrValid(&pkr) {
			return errors.New("invalid pkr")
		}
	} else {
		return errors.New("pkr address must be 96 bytes")
	}
	return nil
}

func Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty hex strin")
	}
	if !has0xPrefix(input) {
		return nil, errors.New("hex string without 0x prefix")
	}
	b, err := hex.DecodeString(input[2:])
	if err != nil {
		return nil, err
	}
	return b, err
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

func IsHex(s string) bool {
	if has0xPrefix(s) {
		s = s[2:]
	}

	for _, c := range []byte(s) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}
func DecodeHex(hex string) (bytes []byte, err error) {
	if strings.Index(hex, "0x") != 0 {
		hex = "0x" + hex
	}
	if bytes, err = Decode(hex); err != nil {
		return
	} else {
		if len(bytes) == 0 {
			err = errors.New("the bytes length is 0")
			return
		} else {
			return
		}
	}
}

func DecodeAddr(input []byte) (bytes []byte, e error) {
	if IsBase58Str(string(input)) {
		bytes = base58.Decode(string(input))
		return
	} else if IsHex(string(input)) {
		return DecodeHex(string(input))

	} else {
		e = errors.New("invalid address string")
		return
	}
}
