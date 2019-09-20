package ethapi

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"regexp"
	"strconv"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-czero-import/seroparam"
)

type decError struct{ msg string }

func (err decError) Error() string { return err.msg }

var (
	ErrEmptyString   = &decError{"empty input string"}
	ErrSyntax        = &decError{"invalid hex string"}
	ErrMissingPrefix = &decError{"hex string without 0x prefix"}
	ErrOddLength     = &decError{"hex string of odd length"}
	ErrUint64Range   = &decError{"hex number > 64 bits"}
)

func IsBase58Str(s string) bool {

	pattern := "^[" + "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz" + "]+$"
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match

}
func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

func bytesHave0xPrefix(input []byte) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func checkText(input []byte, wantPrefix bool) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil // empty strings are allowed
	}
	if bytesHave0xPrefix(input) {
		input = input[2:]
	} else if wantPrefix {
		return nil, ErrMissingPrefix
	}
	if len(input)%2 != 0 {
		return nil, ErrOddLength
	}
	return input, nil
}

func mapError(err error) error {
	if err, ok := err.(*strconv.NumError); ok {
		switch err.Err {
		case strconv.ErrRange:
			return ErrUint64Range
		case strconv.ErrSyntax:
			return ErrSyntax
		}
	}
	if _, ok := err.(hex.InvalidByteError); ok {
		return ErrSyntax
	}
	if err == hex.ErrLength {
		return ErrOddLength
	}
	return err
}

type Big big.Int

func (b Big) MarshalJSON() ([]byte, error) {
	i := big.Int(b)
	by, err := i.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if seroparam.IsExchangeValueStr() {
		bytes := make([]byte, len(by)+2)
		copy(bytes[1:len(bytes)-1], by[:])
		bytes[0] = '"'
		bytes[len(bytes)-1] = '"'
		return bytes, nil
	} else {
		return by, err
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Big) UnmarshalJSON(input []byte) error {
	if isString(input) {
		input = input[1 : len(input)-1]
	}
	i := big.Int{}
	if e := i.UnmarshalText(input); e != nil {
		return e
	} else {
		*b = Big(i)
		return nil
	}
}

func (b *Big) ToInt() *big.Int {
	return (*big.Int)(b)
}

type PKAddress [64]byte

func (b PKAddress) ToUint512() c_type.Uint512 {
	result := c_type.Uint512{}
	copy(result[:], b[:])

	return result
}

func (b PKAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b[:])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if !addr.MatchProtocol("SP") {
			return errors.New("address protocol is not pk")
		}
		if len(addr.Bytes) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], addr.Bytes)
			if superzk.IsPKValid(&pk) {
				copy(b[:], pk[:])
			} else {
				return errors.New("invalid PK")
			}
		} else {
			return errors.New("ivalid PK")
		}
		return nil
	}
}

type TKAddress [64]byte

func (b TKAddress) ToUint512() c_type.Uint512 {
	result := c_type.Uint512{}
	copy(result[:], b[:])

	return result
}

func (b TKAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b[:])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *TKAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if !addr.MatchProtocol("ST") {
			return errors.New("address protocol is not tk")
		}
		if len(addr.Bytes) == 64 {
			copy(b[:], addr.Bytes)
		} else {
			return errors.New("ivalid TK")
		}
		return nil
	}
}

type PKrAddress [96]byte

func (b PKrAddress) ToPKr() *c_type.PKr {
	result := &c_type.PKr{}
	copy(result[:], b[:])

	return result
}

func (b PKrAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b[:])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKrAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if !addr.MatchProtocol("SC") {
			return errors.New("address protocol is not pkr")
		}
		if len(addr.Bytes) == 96 {
			pkr := c_type.PKr{}
			copy(pkr[:], addr.Bytes[:])
			if superzk.IsPKrValid(&pkr) {
				copy(b[:], pkr[:])
				return nil
			} else {
				return errors.New("invalid PKr")
			}
		} else {
			return errors.New("ivalid PKr")
		}
	}
}

type MixAdrress []byte

func (b MixAdrress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *MixAdrress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return nil
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		out := addr.Bytes
		if len(out) == 96 {
			pkr := c_type.PKr{}
			copy(pkr[:], out[:])
			if superzk.IsPKrValid(&pkr) {
				*b = out[:]
				return nil
			} else {
				return errors.New("invalid PKr")
			}
		} else if len(out) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], out[:])
			if superzk.IsPKValid(&pk) {
				*b = out[:]
				return nil
			} else {
				return errors.New("invalid PK")
			}
		} else {
			return errors.New("invalid mix address")
		}
	}
}

type MixBase58Adrress []byte

func (b MixBase58Adrress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *MixBase58Adrress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if addr.IsHex {
			return errors.New("is not base58 address")
		}
		out := addr.Bytes
		if len(out) == 96 {
			pkr := c_type.PKr{}
			copy(pkr[:], out[:])
			if superzk.IsPKrValid(&pkr) {
				*b = out[:]
				return nil
			} else {
				return errors.New("invalid PKr")
			}
		} else if len(out) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], out[:])
			if superzk.IsPKValid(&pk) {
				*b = out[:]
				return nil
			} else {
				return errors.New("invalid PK")
			}
		} else {
			return errors.New("invalid mix address")
		}
	}
}

type AllMixedAddress []byte

func (b AllMixedAddress) IsContract() bool {
	empty := common.Hash{}
	if len(b) == 96 {
		if bytes.Compare(b[64:], empty[:]) == 0 {
			return true
		}
	}
	return false
}

func (b AllMixedAddress) ToPKr() (ret c_type.PKr) {
	copy(ret[:], b[:])
	return
}

func (b AllMixedAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *AllMixedAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		out := addr.Bytes
		if len(out) == 96 {
			addr := common.Address{}
			copy(addr[:], out)
			if isContract, err := txtool.Ref_inst.Bc.IsContract(addr); err != nil {
				return err
			} else {
				if isContract {
					*b = out[:]
					return nil
				} else {
					pkr := c_type.PKr{}
					copy(pkr[:], out[:])
					if superzk.IsPKrValid(&pkr) {
						*b = out[:]
						return nil
					} else {
						return errors.New("invalid PKr")
					}
				}
			}
		} else if len(out) == 64 {
			contract_addr := common.Address{}
			copy(contract_addr[:], out)
			if isContract, err := txtool.Ref_inst.Bc.IsContract(contract_addr); err != nil {
				return err
			} else {
				if isContract {
					*b = contract_addr[:]
					return nil
				} else {
					pk := c_type.Uint512{}
					copy(pk[:], out[:])
					if superzk.IsPKValid(&pk) {
						pkr := superzk.Pk2PKr(&pk, nil)
						*b = pkr[:]
						return nil
					} else {
						return errors.New("invalid PK")
					}
				}
			}
		} else {
			return errors.New("AllMixedAddress must be length 64 or 96")
		}
	}

	return nil

}

type ContractAddress c_type.PKr

func (b ContractAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b[:])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *ContractAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if !addr.MatchProtocol("SS") {
			return errors.New("address protocol is not contract")
		}
		out := addr.Bytes
		if len(out) == 96 {
			addr := common.Address{}
			copy(addr[:], out)
			if isContract, err := txtool.Ref_inst.Bc.IsContract(addr); err != nil {
				return err
			} else {
				if isContract {
					copy(b[:], out)
					return nil
				} else {
					return errors.New("this 96 bytes not contract address")
				}
			}
		} else if len(out) == 64 {
			contract_addr := common.Address{}
			copy(contract_addr[:], out)
			if isContract, err := txtool.Ref_inst.Bc.IsContract(contract_addr); err != nil {
				return err
			} else {
				if isContract {
					copy(b[:], contract_addr[:])
					return nil
				} else {
					return errors.New("this 64 bytes not contract address")
				}
			}
		} else {
			return errors.New("ContractAddress must be length 64 or 96")
		}
	}

}
