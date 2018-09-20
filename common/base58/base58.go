package base58

import (
	"fmt"
	"github.com/sero-cash/go-czero-import/cpt"
	"regexp"
)

var (
	//base        = big.NewInt(58)
	b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
)

type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("encoding/base58: invalid byte: %#U", rune(e))
}

func ReverseBytes(runes []byte) {
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
}

func EncodeToString(input []byte) string {
	return *cpt.Base58Encode(input)
}

//var Big0 = big.NewInt(0)

func Encode(input []byte) []byte {

	return []byte(EncodeToString(input))
	//uint512 := big.NewInt(0).SetBytes(input)
	//u := big.NewInt(0)
	//var result []byte
	//for uint512!=Big0 {
	//	uint512, u = uint512.DivMod(uint512, base, u)
	//	result = append(result, b58Alphabet[u.Int64()])
	//}
	//
	//ReverseBytes(result)
	//
	//for _, b := range input {
	//	if b == 0x00 {
	//		result = append([]byte{b58Alphabet[0]}, result...)
	//
	//	} else {
	//		break
	//	}
	//}
	//return result
}

func DecodeString(s string, out []byte) error {

	err := cpt.Base58Decode(&s, out[:])
	if err != nil {
		return err
	}
	return nil

	//input := []byte(s)
	//zeroBytes := 0
	//
	//for _, b := range input {
	//	if b != b58Alphabet[0] {
	//		break
	//	}
	//
	//	zeroBytes++
	//}
	//
	//uint512 := big.NewInt(0)
	//for _, c := range s {
	//	charIndex := bytes.IndexByte(b58Alphabet, byte(c))
	//	if charIndex == -1 {
	//		return nil, InvalidByteError(c)
	//	}
	//	u := big.NewInt(int64(charIndex))
	//	uint512 = uint512.Mul(uint512, base)
	//	uint512 = uint512.Add(uint512, u)
	//}
	//
	//result := uint512.Bytes()
	//result = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), result...)
	//
	//return result, nil
}

func IsBase58Str(s string) bool {

	pattern := "[" + string(b58Alphabet) + "]+"
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match

}
