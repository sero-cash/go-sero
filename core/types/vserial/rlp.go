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
)

type Version struct {
	V VersionType
}

type VSerial struct {
	Versions []interface{}
	V        Version
}

func (self *VSerial) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	if size == 0 {
		self.V.V = VERSION_NIL
	} else {
		if size > 10 {
			self.V.V = VERSION_0
		} else {
			if e := s.Decode(&self.V); e != nil {
				return e
			}
		}
	}
	if int(self.V.V) > len(self.Versions) {
		log.Error("VSerial DecodeRLP ERROR: version num is error", "version", self.V.V, "len", len(self.Versions))
		return errors.New("VSerial DecodeRLP ERROR: version num is error")
	}
	if self.V.V == VERSION_NIL {
		if e := s.Decode(self.Versions[0]); e != nil {
			return e
		}
	} else {
		for i := 0; i <= int(self.V.V); i++ {
			if e := s.Decode(self.Versions[i]); e != nil {
				return e
			}
		}
	}
	return nil
}

func (self *VSerial) EncodeRLP(w io.Writer) error {
	v := Version{}
	v.V = VersionType(len(self.Versions) - 1)
	if v.V == VERSION_NIL {
		e := errors.New("encode header rlp error: version is nil")
		panic(e)
		return e
	}
	if v.V >= VERSION_1 {
		if e := rlp.Encode(w, &v); e != nil {
			return e
		}
	}
	for _, it := range self.Versions {
		if e := rlp.Encode(w, it); e != nil {
			return e
		}
	}
	return nil
}
