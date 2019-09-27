package utils

import (
	"crypto/md5"
	"regexp"
	"strings"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/btcsuite/btcutil/base58"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-sero/common/hexutil"
)

type Address struct {
	Bytes    []byte
	Base58   string
	Protocol string
	Version  string
	Sum      string
	IsHex    bool
}

const hextable = "0123456789abcdef"

func (self *Address) calcSum() {
	c := append([]byte(self.Protocol+self.Version), self.Bytes...)
	m := md5.Sum(c)
	s := base58.Encode(m[:])
	self.Sum = s[:2]
}

func (self *Address) genVersion() {
	if c_superzk.IsFlagSet(self.Bytes) {
		self.Version = "1"
	} else {
		self.Version = "0"
	}
}

func (self *Address) setBytes(bs []byte) {
	self.Bytes = bs
	self.Base58 = base58.Encode(self.Bytes)
	self.genVersion()
}

func (self *Address) setHex(hex string) (e error) {
	self.IsHex = true
	if strings.Index(hex, "0x") != 0 {
		hex = "0x" + hex
	}
	if bytes, err := hexutil.Decode(hex); err != nil {
		e = err
		return
	} else {
		if len(bytes) == 0 {
			e = errors.New("the bytes length is 0")
			return
		} else {
			self.setBytes(bytes)
			return
		}
	}
}

func (self *Address) setBase58(bs string) (e error) {
	if IsBase58Str(bs) {
		bytes := base58.Decode(bs)
		if len(bytes) == 0 {
			e = errors.New("the bytes length is 0")
			return
		} else {
			self.Base58 = bs
			self.Bytes = bytes
			self.genVersion()
			return
		}
	} else {
		e = errors.New("the addr is not base58")
		return
	}

}
func (self *Address) ToCode() (ret string) {
	return self.Protocol + self.Version + "." + self.Base58 + "." + self.Sum
}

func (self *Address) ToHex() (ret string) {
	return hexutil.Encode(self.Bytes)
}

func (self *Address) ToBase58() (ret string) {
	return self.Base58
}

func (self *Address) MatchProtocol(ver string) bool {
	if len(self.Protocol) > 0 {
		return self.Protocol == ver
	} else {
		return true
	}
}

func (self *Address) SetProtocol(p string) *Address {
	self.Protocol = p
	self.calcSum()
	return self
}

func NewAddressByBytes(addr []byte) (ret Address) {
	ret.setBytes(addr)
	ret.calcSum()
	return
}

func NewAddressByHex(addr string) (ret Address, e error) {
	if e = ret.setHex(addr); e != nil {
		return
	}
	ret.calcSum()
	return
}

func IsBase58Str(s string) bool {
	pattern := "^[" + "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz" + "]+$"
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return match
}

func NewAddressByBase58(addr string) (ret Address, e error) {
	if e = ret.setBase58(addr); e != nil {
		return
	}
	ret.calcSum()
	return
}

var reg, _ = regexp.Compile(`^(.*)\.(.*)\.(.*)$`)

func NewAddressByString(addr string) (ret Address, e error) {
	if strs := reg.FindStringSubmatch(addr); len(strs) != 4 {
		if IsBase58Str(addr) {
			return NewAddressByBase58(addr)
		} else {
			return NewAddressByHex(addr)
		}
	} else {
		if e = ret.setBase58(strs[2]); e != nil {
			return
		}
		ret.Protocol = strs[1][:len(strs[1])-1]
		ret.calcSum()
		if ret.Version != strs[1][len(strs[1])-1:] {
			e = errors.New("the version check failed")
			return
		}
		if ret.Sum != strs[3] {
			e = errors.New("the sum check failed")
			return
		}
		return
	}
}
