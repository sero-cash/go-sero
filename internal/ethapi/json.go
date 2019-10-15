package ethapi

import (
	"bytes"
	"math/big"

	"github.com/sero-cash/go-sero/common/address"

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

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

func bytesHave0xPrefix(input []byte) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
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

type PKrAddress [96]byte

func (b PKrAddress) ToPKr() *c_type.PKr {
	result := &c_type.PKr{}
	copy(result[:], b[:])

	return result
}

func (b PKrAddress) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b PKrAddress) String() string {
	return base58.Encode(b[:])

}

func (b PKrAddress) Base58() string {
	return base58.Encode(b[:])
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKrAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	out, err := address.DecodeAddr(input)
	if err != nil {
		return err
	}
	err = address.ValidPkr(out)
	if err != nil {
		return err
	}
	copy(b[:], out)
	return nil
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
	out, err := address.DecodeAddr(input)
	if err != nil {
		return err
	}

	if len(out) == 96 {
		err := address.ValidPkr(out)
		if err != nil {
			return err
		}
		*b = out[:]
		return nil
	} else if len(out) == 64 {
		err := address.ValidPk(out)
		if err != nil {
			return err
		}
		*b = out[:]
		return nil
	} else {
		return errors.New("invalid mix address")
	}
}

type AllMixedAddress [96]byte

func (b AllMixedAddress) setBytes(bs []byte) {
	copy(b[:], bs)
}

func (b AllMixedAddress) ToPKrAddress() (ret PKrAddress) {
	copy(ret[:], b[:])
	return
}

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
	if b.IsContract() {
		return []byte(base58.Encode(b[:64])), nil
	} else {
		return []byte(base58.Encode(b[:])), nil
	}
}

func IsContract(b []byte) (bool, error) {
	addr := common.Address{}
	copy(addr[:], b)
	return txtool.Ref_inst.Bc.IsContract(addr)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *AllMixedAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}
	out, err := address.DecodeAddr(input)
	if err != nil {
		return err
	}
	if len(out) == 96 {
		if isContract, err := IsContract(out); err != nil {
			return err
		} else {
			if isContract {
				copy(b[:], out)
				return nil
			} else {
				err := address.ValidPkr(out)
				if err != nil {
					return err
				}
				copy(b[:], out)
				return nil
			}
		}
	} else if len(out) == 64 {
		contract_addr := common.Address{}
		copy(contract_addr[:], out)
		if isContract, err := txtool.Ref_inst.Bc.IsContract(contract_addr); err != nil {
			return err
		} else {
			if isContract {
				copy(b[:], out)
				return nil
			} else {
				err := address.ValidPk(out)
				if err != nil {
					return err
				}
				pk := c_type.Uint512{}
				copy(pk[:], out[:])
				pkr := superzk.Pk2PKr(&pk, nil)
				copy(b[:], pkr[:])
			}
		}
	} else {
		return errors.New("AllMixedAddress must be length 64 or 96")
	}

	return nil

}

type ContractAddress c_type.PKr

func (b *ContractAddress) SetBytes(bs []byte) {
	copy(b[:], bs)
}

func (b ContractAddress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b[:64])), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *ContractAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}
	out, err := address.DecodeAddr(input)
	if err != nil {
		return err
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

type AllBase58Adrress []byte

func (b AllBase58Adrress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b)), nil
}

func (b AllBase58Adrress) String() string {
	return base58.Encode(b)
}
func (b AllBase58Adrress) Bytes() []byte {
	return []byte(b)
}

func (b AllBase58Adrress) ToPkr(isContract bool) (ret c_type.PKr) {
	if len(b) == 96 {
		copy(ret[:], b[:])
	} else {
		if isContract {
			copy(ret[:], b[:])
		} else {
			var pk c_type.Uint512
			copy(pk[:], b[:])
			ret = superzk.Pk2PKr(&pk, nil)
		}
	}
	return
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *AllBase58Adrress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return nil
	}
	out, err := address.DecodeAddr(input)
	if err != nil {
		return err
	}
	if len(out) == 96 {
		err := address.ValidPkr(out)
		if err != nil {
			return err
		}
		*b = out[:]
		return nil
	} else if len(out) == 64 {
		*b = out[:]
		return nil
	} else {
		return errors.New("invalid base58 address")
	}
}
