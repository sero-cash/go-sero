package utils

import (
	"fmt"
	"io"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
)

type U256 big.Int

var U256_0 U256 = U256(*big.NewInt(0))

func NewU256(i uint64) (ret U256) {
	ret = U256(*big.NewInt(int64(i)))
	return
}

func (x *U256) GobEncode() ([]byte, error) {
	b := big.Int(*x)
	return b.GobEncode()
}

func (z *U256) GobDecode(buf []byte) error {
	var a big.Int
	if err := a.GobDecode(buf); err != nil {
		return err
	}
	*z = U256(a)
	return nil
}

func (b U256) EncodeRLP(w io.Writer) error {
	i := big.Int(b)
	if bytes, e := i.GobEncode(); e != nil {
		return e
	} else {
		if e = rlp.Encode(w, bytes); e != nil {
			return e
		} else {
			return nil
		}
	}
}

func (b *U256) DecodeRLP(s *rlp.Stream) error {
	bytes := []byte{}
	if e := s.Decode(&bytes); e != nil {
		return e
	} else {
		i := big.Int{}
		if e := i.GobDecode(bytes); e != nil {
			return e
		} else {
			*b = U256(i)
			return nil
		}
	}
}

func (b U256) MarshalJSON() ([]byte, error) {
	i := big.Int(b)
	return i.MarshalJSON()
}

func (b U256) MarshalText() ([]byte, error) {
	i := big.Int(b)
	return i.MarshalText()
}

func (b *U256) UnmarshalJSON(input []byte) error {
	fmt.Println(string(input))
	i := big.Int{}
	if e := i.UnmarshalJSON(input); e != nil {
		return e
	} else {
		*b = U256(i)
		return nil
	}
}

func (b *U256) UnmarshalText(input []byte) error {
	i := big.Int{}
	if e := i.UnmarshalText(input); e != nil {
		return e
	} else {
		*b = U256(i)
		return nil
	}
}

func NewU256_ByKey(k *keys.Uint256) (ret U256) {
	bytes := *k.NewRef()
	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-1-i] = bytes[len(bytes)-1-i], bytes[i]
	}
	r := big.NewInt(0)
	r.SetBytes(bytes[:])
	ret = U256(*r)
	return
}

func (self U256) ToRef() (ret *U256) {
	s := big.Int(self)
	r := big.NewInt(0)
	r.Set(&s)
	t := U256(*r)
	ret = &t
	return
}

func (self *U256) ToI256() (ret I256) {
	b := big.Int(*self.ToRef())
	ret = I256(b)
	return
}

func (self *U256) ToUint256() (ret keys.Uint256) {
	i := big.Int(*self.ToRef())
	bytes := i.Bytes()
	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-1-i] = bytes[len(bytes)-1-i], bytes[i]
	}
	copy(ret[:], bytes[:])
	return
}

func (self *U256) ToIntRef() (ret *big.Int) {
	r := big.Int(*self.ToRef())
	ret = &r
	return
}

func (self *U256) AddU(a *U256) {
	l := big.Int(*self.ToRef())
	r := big.Int(*a)
	l.Add(&l, &r)
	*self = U256(l)
	return
}

func (self *U256) Cmp(a *U256) int {
	l := big.Int(*self)
	r := big.Int(*a)
	return l.Cmp(&r)
}

func (self *U256) SubU(a *U256) {
	l := big.Int(*self.ToRef())
	r := big.Int(*a)
	l.Sub(&l, &r)
	*self = U256(l)
	return
}
