package addrutil

import (
	"errors"
	"regexp"

	"github.com/sero-cash/go-czero-import/c_czero"

	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-czero-import/c_type"
)

func IsBase58Str(s string) bool {

	pattern := "^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$"
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match

}

func IsString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

var ZEROBYTES = [32]byte{}

func getAddressSuffix(addr []byte) [32]byte {
	suffix := [32]byte{}
	copy(suffix[:], addr[64:])
	return suffix
}

func IsValidBase58Address(input []byte) ([]byte, error) {
	if IsBase58Str(string(input)) {
		out := base58.Decode(string(input))
		if len(out) == 64 {
			return out, nil
		} else if len(out) == 96 {
			suffix := getAddressSuffix(out)
			if ZEROBYTES == suffix {
				return out, nil
			} else {
				pkr := c_type.PKr{}
				copy(pkr[:], out[:])
				if !c_czero.PKrValid(&pkr) {
					return nil, errors.New("invalid PKr base58")
				} else {
					return out, nil
				}
			}
		} else {
			return nil, errors.New("invalid base58 length")
		}
	} else {
		return nil, errors.New("invalid base58 address")
	}
}

func IsValidBase58AcccountAddress(input []byte) ([]byte, error) {
	if IsBase58Str(string(input)) {
		out := base58.Decode(string(input))
		if len(out) == 64 {
			return out, nil
		} else {
			return nil, errors.New("invalid base58 length")
		}
	} else {
		return nil, errors.New("invalid base58 address")
	}
}

func IsValidAccountAddress(input []byte) ([]byte, error) {
	if IsBase58Str(string(input)) {
		out := base58.Decode(string(input))
		if len(out) == 64 {
			pk := c_type.Uint512{}
			copy(pk[:], out[:])
			if !c_czero.IsPKValid(&pk) {
				return nil, errors.New("invalid PK base58")
			} else {
				return out, nil
			}
		} else {
			return nil, errors.New("invalid base58 length")
		}
	} else {
		return nil, errors.New("invalid base58 address")
	}

}
