package utils

import (
	"io"
	"math/big"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/rlp"
)

type I256 big.Int

var I256_0 I256 = I256(*big.NewInt(0))

func NewI256(i int64) (ret I256) {
	ret = I256(*big.NewInt(i))
	return
}

func (self I256) DeepCopy() interface{} {
	bi := big.Int(self)
	dc := I256(*big.NewInt(0).Set(&bi))
	return dc
}

func (x *I256) GobEncode() ([]byte, error) {
	b := big.Int(*x)
	return b.GobEncode()
}

func (z *I256) GobDecode(buf []byte) error {
	var a big.Int
	if err := a.GobDecode(buf); err != nil {
		return err
	}
	*z = I256(a)
	return nil
}

func (b I256) EncodeRLP(w io.Writer) error {
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

func (b *I256) DecodeRLP(s *rlp.Stream) error {
	bytes := []byte{}
	if e := s.Decode(&bytes); e != nil {
		return e
	} else {
		i := big.Int{}
		if e := i.GobDecode(bytes); e != nil {
			return e
		} else {
			*b = I256(i)
			return nil
		}
	}
}

func (b I256) MarshalText() ([]byte, error) {
	i := big.Int(b)
	return i.MarshalText()
}

func (b *I256) UnmarshalJSON(input []byte) error {
	i := big.Int{}
	if e := i.UnmarshalJSON(input); e != nil {
		return e
	} else {
		*b = I256(i)
		return nil
	}
}

func (b *I256) UnmarshalText(input []byte) error {
	i := big.Int{}
	if e := i.UnmarshalText(input); e != nil {
		return e
	} else {
		*b = I256(i)
		return nil
	}
}

func (self I256) IsPositive() bool {
	l := big.Int(self)
	return l.Sign() >= 0
}

func (self I256) Abs() (ret U256) {
	l := big.Int(*self.ToRef())
	l.Abs(&l)
	ret = U256(l)
	return
}

func (self I256) ToRef() (ret *I256) {
	s := big.Int(self)
	r := big.NewInt(0)
	r.Set(&s)
	t := I256(*r)
	ret = &t
	return
}

func (self *I256) Reverse() {
	l := big.Int(*self.ToRef())
	z := big.NewInt(0)
	z.Sub(z, &l)
	*self = I256(*z)
}

func (self *I256) AddI(a *I256) {
	l := big.Int(*self.ToRef())
	r := big.Int(*a)
	l.Add(&l, &r)
	*self = I256(l)
	return
}

func (self *I256) AddU(a *U256) {
	l := big.Int(*self.ToRef())
	r := big.Int(*a)
	l.Add(&l, &r)
	*self = I256(l)
	return
}

func (self *I256) SubI(a *I256) {
	l := big.Int(*self.ToRef())
	r := big.Int(*a)
	l.Sub(&l, &r)
	*self = I256(l)
	return
}

func (self *I256) SubU(a *U256) {
	l := big.Int(*self.ToRef())
	r := big.Int(*a)
	l.Sub(&l, &r)
	*self = I256(l)
	return
}

func (self *I256) ToIntRef() (ret *big.Int) {
	r := big.Int(*self.ToRef())
	ret = &r
	return
}

func (self *I256) ToUint256() (ret c_type.Uint256) {
	i := big.Int(*self.ToRef())
	bytes := i.Bytes()
	for i := 0; i < len(bytes)/2; i++ {
		bytes[i], bytes[len(bytes)-1-i] = bytes[len(bytes)-1-i], bytes[i]
	}
	copy(ret[:], bytes[:])
	if i.Sign() < 0 {
		ret[31] = 1
	} else {
		ret[31] = 0
	}
	return
}
func (self *I256) Cmp(a *I256) int {
	l := big.Int(*self)
	r := big.Int(*a)
	return l.Cmp(&r)
}
