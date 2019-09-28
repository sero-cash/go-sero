package ethapi

import (
	"bytes"
	"math/big"

	"github.com/sero-cash/go-czero-import/c_superzk"

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
		err := ValidPk(addr)
		if err != nil {
			return err
		}
		copy(b[:], addr.Bytes)
		return nil
	}
}

func ValidPk(addr utils.Address) error {
	if len(addr.Bytes) == 64 {
		if !addr.MatchProtocol("SP") {
			return errors.New("address protocol is not pk")
		}
		pk := c_type.Uint512{}
		copy(pk[:], addr.Bytes)
		if c_superzk.IsSzkPK(&pk) {
			if addr.Protocol == "" {
				return errors.New("pk is new version mush have prefix scp1")
			}
		}
		if !superzk.IsPKValid(&pk) {
			return errors.New("invalid PK")
		}
	} else {
		return errors.New("pk address must be 64 bytes")
	}
	return nil
}

type TKAddress [64]byte

func (b TKAddress) ToTk() c_type.Tk {
	result := c_type.Tk{}
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
		err := ValidPkr(addr)
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

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		out := addr.Bytes
		if len(out) == 96 {
			err := ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
		} else if len(out) == 64 {
			err := ValidPk(addr)
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

func ValidPkr(addr utils.Address) error {
	if len(addr.Bytes) == 96 {
		if !addr.MatchProtocol("SC") {
			return errors.New("address protocol is not pkr")
		}
		var pkr c_type.PKr
		copy(pkr[:], addr.Bytes)
		if c_superzk.IsSzkPKr(&pkr) {
			if addr.Protocol == "" {
				return errors.New("pkr address is new version  must have prefix SC")
			}
		}
		if !superzk.IsPKrValid(&pkr) {
			return errors.New("invalid pkr")
		}
	} else {
		return errors.New("pkr address must be 96 bytes")
	}
	return nil
}

type MixBase58Adrress []byte

func (b MixBase58Adrress) MarshalText() ([]byte, error) {
	return []byte(base58.Encode(b)), nil
}

func (b MixBase58Adrress) IsPkr() bool {
	return len(b) == 96
}

func (b MixBase58Adrress) toPkr() c_type.PKr {
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
			err := ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
		} else if len(out) == 64 {
			err := ValidPk(addr)
			if err != nil {
				return err
			}
			*b = out[:]
			return nil
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
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		out := addr.Bytes
		if len(out) == 96 {
			if isContract, err := IsContract(out); err != nil {
				return err
			} else {
				if isContract {
					*b = out[:]
					return nil
				} else {
					err := ValidPkr(addr)
					if err != nil {
						return err
					}
					*b = out[:]
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
					*b = contract_addr[:]
					return nil
				} else {
					err := ValidPk(addr)
					if err != nil {
						return err
					}
					pk := c_type.Uint512{}
					copy(pk[:], out[:])
					pkr := superzk.Pk2PKr(&pk, nil)
					*b = pkr[:]
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

func (b *MixedcaseAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		b.Origin = string(input)

		if isContract, err := IsContract(addr.Bytes); err != nil {
			if isContract {
				b.Contract = true
				if !addr.MatchProtocol("SS") {
					return errors.New("address protocol is not contract")
				}
				b.Addr = addr.Bytes
				return nil
			} else {
				if len(addr.Bytes) == 96 {
					err := ValidPkr(addr)
					if err != nil {
						return nil
					}
					b.Addr = addr.Bytes
					return nil

				} else if len(addr.Bytes) == 64 {
					err := ValidPk(addr)
					if err != nil {
						return nil
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
