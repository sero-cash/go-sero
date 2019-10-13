package ethapi

import (
	"bytes"
	"math/big"

	"github.com/sero-cash/go-sero/zero/account"

	"github.com/sero-cash/go-czero-import/c_superzk"

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
	if c_superzk.IsSzkPKr(b.ToPKr()) {
		a := account.NewAddressByBytes(b[:])
		a.SetProtocol("SC")
		return []byte(a.ToCode()), nil
	} else {
		bs := base58.Encode(b[:])
		return []byte(bs), nil
	}
}

func (b PKrAddress) String() string {
	a := account.NewAddressByBytes(b[:])
	a.SetProtocol("SC")
	return a.ToCode()
}

func (b PKrAddress) Base58() string {
	return base58.Encode(b[:])
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKrAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := account.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		err := account.ValidPkr(addr)
		if err != nil {
			return err
		}
		copy(b[:], addr.Bytes)
		return nil
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

	if addr, e := account.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		out := addr.Bytes
		if len(out) == 96 {
			err := account.ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
		} else if len(out) == 64 {
			err := account.ValidPk(addr)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
		} else {
			return errors.New("invalid mix address")
		}
	}
	return nil
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
	if addr, e := account.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		out := addr.Bytes
		if len(out) == 96 {
			if isContract, err := IsContract(out); err != nil {
				return err
			} else {
				if isContract {
					copy(b[:], out)
					return nil
				} else {
					err := account.ValidPkr(addr)
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
					err := account.ValidPk(addr)
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

	if addr, e := account.NewAddressByString(string(input)); e != nil {
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

type MixedcaseAddress struct {
	Addr     []byte
	Origin   string
	Contract bool
}

func (b MixedcaseAddress) String() string {
	return b.Origin
}
func (b MixedcaseAddress) IsPkr() bool {
	if b.Contract {
		return false
	} else {
		return len(b.Addr) == 96
	}
}
func (b MixedcaseAddress) IsContract() bool {
	return b.Contract
}
func (b MixedcaseAddress) ToPkr(r *c_type.Uint256) (ret c_type.PKr) {
	if b.IsContract() {
		copy(ret[:], b.Addr)
		return
	}
	if b.IsPkr() {
		copy(ret[:], b.Addr)
		return
	}
	pk := c_type.Uint512{}
	copy(pk[:], b.Addr)
	pkr := superzk.Pk2PKr(&pk, r)
	copy(ret[:], pkr[:])
	return
}

func (b MixedcaseAddress) MarshalText() ([]byte, error) {
	return []byte(b.Origin), nil
}

func (b *MixedcaseAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}

	if addr, e := account.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		b.Origin = string(input)

		if isContract, err := IsContract(addr.Bytes); err == nil {
			if isContract {
				b.Contract = true
				if !addr.MatchProtocol("SS") {
					return errors.New("address protocol is not contract")
				}
				b.Addr = addr.Bytes
				return nil
			} else {
				if len(addr.Bytes) == 96 {
					err := account.ValidPkr(addr)
					if err != nil {
						return err
					}
					b.Addr = addr.Bytes
					return nil

				} else if len(addr.Bytes) == 64 {
					err := account.ValidPk(addr)
					if err != nil {
						return err
					}
					b.Addr = addr.Bytes
					return nil
				} else {
					return errors.New("invalid address")
				}

			}

		} else {
			return err
		}

	}
}
