package base58

import (
	"fmt"
	"regexp"

	"github.com/sero-cash/go-czero-import/cpt"
)

var (
	b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
)

type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("encoding/base58: invalid byte: %#U", rune(e))
}

func EncodeToString(input []byte) string {
	return *cpt.Base58Encode(input)
}

func Encode(input []byte) []byte {

	return []byte(EncodeToString(input))
}

func DecodeString(s string, out []byte) error {

	err := cpt.Base58Decode(&s, out[:])

	return err

}

func IsBase58Str(s string) bool {

	pattern := "^[" + string(b58Alphabet) + "]+$"
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match

}
