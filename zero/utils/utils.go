// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
)

func Uint64ToBytes(r uint64) []byte {
	value := new(big.Int).SetUint64(r)
	return value.Bytes()
}
func Int64ToBytes(r int64) []byte {
	value := new(big.Int).SetInt64(r)
	return value.Bytes()
}
func Uint256SliceCut(is []keys.Uint256, l int) (ret []keys.Uint256) {
	is_l := len(is)
	if is_l < l {
		l = is_l
	}
	ret = is[:l]
	return
}

func DeepCopy(dst, src interface{}) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		panic(fmt.Sprintf("deepCopy encode error for : %v", src))
	}
	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst); err != nil {
		panic(fmt.Sprintf("deepCopy decode error for : %v", src))
	}
}

type Uint256s []keys.Uint256

func (self Uint256s) Len() int {
	return len(self)
}
func (self Uint256s) Less(i, j int) bool {
	return bytes.Compare(self[i][:], self[j][:]) < 0
}
func (self Uint256s) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type I256 big.Int

var I256_0 I256 = I256(*big.NewInt(0))

func NewI256(i int64) (ret I256) {
	ret = I256(*big.NewInt(i))
	return
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

func (self *I256) ToUint256() (ret keys.Uint256) {
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

func (b U256) MarshalText() ([]byte, error) {
	i := big.Int(b)
	return i.MarshalText()
}

func (b *U256) UnmarshalJSON(input []byte) error {
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

type TimeRecord struct {
	start time.Time
	name  string
}

func TR_enter(name string) (tr TimeRecord) {
	//fmt.Printf("\n{{\nStart (" + name + ") >>>>>> \n")
	//tr.start = time.Now()
	//tr.name = name
	return
}

//func (tr *TimeRecord) Renter(name string) {
//	fmt.Printf(" ...... [[ Ren ("+tr.name+":"+name+")     s=%v ]]\n", time.Since(tr.start))
//	tr.start = time.Now()
//}

func (tr *TimeRecord) Leave() {
	//fmt.Printf("End ("+tr.name+")     s=%v  <<<<<<<\n}}\n", time.Since(tr.start))
}

func StringToUint256(str string) keys.Uint256 {
	var ret keys.Uint256
	b := []byte(strings.ToUpper(str))
	if len(b) > len(ret) {
		b = b[len(b)-len(ret):]
	}
	copy(ret[len(ret)-len(b):], b)
	return ret

}

type Proc interface {
	Run() bool
}

type Procs struct {
	ch   chan int
	wait sync.WaitGroup
	Runs []Proc
	succ bool
}

func NewProcs(num int) (ret Procs) {
	ret = Procs{
		make(chan int, num),
		sync.WaitGroup{},
		nil,
		true,
	}
	return
}

func (self *Procs) StartProc(run Proc) {
	self.Runs = append(self.Runs, run)
	if cpt.Is_czero_debug() {
		if !run.Run() {
			self.succ = false
		}
	} else {
		self.wait.Add(1)
		go func(run Proc) {
			self.ch <- 0
			defer func() {
				<-self.ch
				self.wait.Done()
			}()
			if !run.Run() {
				self.succ = false
			}
		}(run)
	}
}

func (self *Procs) Wait() []Proc {
	self.wait.Wait()
	if self.succ {
		p := self.Runs
		self.Runs = nil
		return p
	} else {
		return nil
	}
}
