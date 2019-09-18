package ethapi

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"regexp"
	"strconv"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"

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
	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {

		out := base58.Decode(string(input))
		if len(out) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], out[:])
			if c_czero.IsPKValid(&pk) {
				copy(b[:], out[:])
			} else {
				return errors.New("invalid PK")
			}
		} else {
			return errors.New("ivalid PK")
		}
		return nil
	} else {
		raw, err := checkText(input, true)
		if err != nil {
			return err
		}
		dec := make([]byte, len(raw)/2)
		if _, err = hex.Decode(dec, raw); err != nil {
			err = mapError(err)
		} else {
			if len(dec) != 64 {
				return errors.New("PKAddress must be length 64 ")
			}
			pk := c_type.Uint512{}
			copy(pk[:], dec[:])
			if c_czero.IsPKValid(&pk) {
				copy(b[:], pk[:])
			} else {
				return errors.New("invalid PK")
			}
		}
		return err
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
	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {

		out := base58.Decode(string(input))
		if len(out) == 64 {
			copy(b[:], out[:])
		} else {
			return errors.New("ivalid TK")
		}
		return nil
	} else {
		raw, err := checkText(input, true)
		if err != nil {
			return err
		}
		dec := make([]byte, len(raw)/2)
		if _, err = hex.Decode(dec, raw); err != nil {
			err = mapError(err)
		} else {
			if len(dec) != 64 {
				return errors.New("TKAddress must be length 64 ")
			}
			copy(b[:], dec)
		}
		return err
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
	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {
		out := base58.Decode(string(input))
		if len(out) == 96 {
			pkr := c_type.PKr{}
			copy(pkr[:], out[:])
			if c_czero.PKrValid(&pkr) {
				copy(b[:], out[:])
			} else {
				return errors.New("invalid PKr")
			}
		} else {
			return errors.New("ivalid PKr")
		}
		return nil
	} else {
		raw, err := checkText(input, true)
		if err != nil {
			return err
		}
		dec := make([]byte, len(raw)/2)
		if _, err = hex.Decode(dec, raw); err != nil {
			err = mapError(err)
		} else {
			if len(dec) != 96 {
				return errors.New("PKrAddress must be length 96")
			}
			pkr := c_type.PKr{}
			copy(pkr[:], dec[:])
			if c_czero.PKrValid(&pkr) {
				copy(b[:], pkr[:])
			} else {
				return errors.New("invalid PKr")
			}
		}
		return err
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

	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {
		out := base58.Decode(string(input))
		if len(out) == 96 {
			pkr := c_type.PKr{}
			copy(pkr[:], out[:])
			if c_czero.PKrValid(&pkr) {
				*b = out[:]
			} else {
				return errors.New("invalid PKr")
			}
		} else if len(out) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], out[:])
			if c_czero.IsPKValid(&pk) {
				*b = out[:]
			} else {
				return errors.New("invalid PK")
			}
		} else {
			return errors.New("invalid mix address")
		}
		return nil

	} else {
		raw, err := checkText(input, true)
		if err != nil {
			return err
		}
		dec := make([]byte, len(raw)/2)
		if _, err = hex.Decode(dec, raw); err != nil {
			err = mapError(err)
		} else {
			if len(dec) != 64 && len(dec) != 96 {
				return errors.New("MixAddress must be length 64 or 96")
			}
			*b = dec
		}
		return err
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
	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {
		out := base58.Decode(string(input))
		if len(out) == 96 {
			pkr := c_type.PKr{}
			copy(pkr[:], out[:])
			if c_czero.PKrValid(&pkr) {
				*b = out[:]
			} else {
				return errors.New("invalid PKr")
			}
		} else if len(out) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], out[:])
			if c_czero.IsPKValid(&pk) {
				*b = out[:]
			} else {
				return errors.New("invalid PK")
			}
		} else {
			return errors.New("invalid mix address")
		}
		return nil

	} else {
		return errors.New("not base58 string")
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

	var out []byte
	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {
		out = base58.Decode(string(input))
	} else {
		if raw, err := checkText(input, true); err != nil {
			return err
		} else {
			dec := make([]byte, len(raw)/2)
			if _, err = hex.Decode(dec, raw); err != nil {
				return mapError(err)
			} else {
				out = dec
			}
		}
	}

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
				if c_czero.PKrValid(&pkr) {
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
				if c_czero.IsPKValid(&pk) {
					pkr := c_czero.Addr2PKr(&pk, nil)
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

	var out []byte
	if IsBase58Str(string(input)) && !bytesHave0xPrefix(input) {
		out = base58.Decode(string(input))
	} else {
		if raw, err := checkText(input, true); err != nil {
			return err
		} else {
			dec := make([]byte, len(raw)/2)
			if _, err = hex.Decode(dec, raw); err != nil {
				return mapError(err)
			} else {
				out = dec
			}
		}
	}

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
