package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
)

type Current struct {
	Index int64
}

func NewCur() (ret Current) {
	ret.Index = -1
	return
}

func (self *Current) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type CurrentGet struct {
	Out Current
}

func (self *CurrentGet) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.Out = NewCur()
		return
	} else {
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}

type StateBlock struct {
	Roots []keys.Uint256
	Dels  []keys.Uint256
}

func (self *StateBlock) Serial() (ret []byte, e error) {
	if self != nil {
		if bytes, err := rlp.EncodeToBytes(self); err != nil {
			e = err
			return
		} else {
			ret = bytes
			return
		}
	} else {
		return
	}
}

type State0BlockGet struct {
	Out StateBlock
}

func (self *State0BlockGet) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		return
	} else {
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
