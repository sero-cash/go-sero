package vserial

import (
	"errors"
	"io"

	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-sero/rlp"
)

type VersionType int8

const (
	VERSION_NIL = VersionType(-1)
	VERSION_0   = VersionType(0)
	VERSION_1   = VersionType(1)
	VERSION_2   = VersionType(2)
)

type Version struct {
	V VersionType
}

type vSerial struct {
	versions []interface{}
	v        Version
}

func NewVSerial() (ret vSerial) {
	ret.v.V = VERSION_NIL
	return
}

func (self *vSerial) V() VersionType {
	return self.v.V
}
func (self *vSerial) Add(data interface{}, ver VersionType) {
	if ver <= self.v.V {
		panic(errors.New("vserial add version error"))
	}
	self.v.V = ver
	self.versions = append(self.versions, data)
}

func (self *vSerial) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	if size == 0 {
		self.v.V = VERSION_NIL
	} else {
		if size > 10 {
			self.v.V = VERSION_0
		} else {
			if e := s.Decode(&self.v); e != nil {
				return e
			}
		}
	}
	if int(self.v.V) >= len(self.versions) {
		log.Error("VSerial DecodeRLP ERROR: version num is error", "version", self.v.V, "len", len(self.versions))
		return errors.New("VSerial DecodeRLP ERROR: version num is error")
	}
	if self.v.V == VERSION_NIL {
		if e := s.Decode(self.versions[0]); e != nil {
			return e
		}
	} else {
		if self.v.V <= VERSION_1 {
			for i := 0; i <= int(self.v.V); i++ {
				if e := s.Decode(self.versions[i]); e != nil {
					return e
				}
			}
		} else {
			if e := s.Decode(self.versions[self.v.V]); e != nil {
				return e
			}
		}
	}
	return nil
}

func (self *vSerial) EncodeRLP(w io.Writer) error {
	if self.v.V == VERSION_NIL {
		e := errors.New("encode header rlp error: version is nil")
		panic(e)
		return e
	}
	if self.v.V >= VERSION_1 {
		if e := rlp.Encode(w, &self.v); e != nil {
			return e
		}
	}
	if self.v.V <= VERSION_1 {
		for _, it := range self.versions {
			if e := rlp.Encode(w, it); e != nil {
				return e
			}
		}
	} else {
		if e := rlp.Encode(w, self.versions[len(self.versions)-1]); e != nil {
			return e
		}
	}
	return nil
}
