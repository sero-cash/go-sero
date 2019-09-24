package localdb

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/utils"
)

type OutState struct {
	Index  uint64
	Out_O  *stx_v0.Out_O   `rlp:"nil"`
	Out_Z  *stx_v0.Out_Z   `rlp:"nil"`
	Out_P  *stx_v1.Out_P   `rlp:"nil"`
	Out_C  *stx_v1.Out_C   `rlp:"nil"`
	OutCM  *c_type.Uint256 `rlp:"nil"`
	RootCM *c_type.Uint256 `rlp:"nil"`
}

func (self *OutState) genOutCM() {
	if self.OutCM == nil {
		if cm, err := genOutCM(self); err != nil {
			panic(err)
			return
		} else {
			self.OutCM = &cm
			return
		}
	} else {
		return
	}
}

func (self *OutState) GenRootCM() {
	if self.Out_O != nil || self.Out_Z != nil {
		self.genOutCM()
	}
	if self.RootCM == nil {
		if cm, err := genRootCM(self); err != nil {
			panic(err)
			return
		} else {
			self.RootCM = &cm
			return
		}
	} else {
		return
	}
}

func (out *OutState) TxType() string {
	if out.Out_O != nil {
		return "Out_O"
	}
	if out.Out_Z != nil {
		return "Out_Z"
	}
	if out.Out_P != nil {
		return "Out_P"
	}
	if out.Out_C != nil {
		return "Out_C"
	}
	return "EMPTY"
}

func (out *OutState) IsZero() bool {
	if out.Out_Z != nil || out.Out_C != nil {
		return true
	} else {
		return false
	}
}
func (out *OutState) IsSzk() bool {
	if out.Out_P != nil || out.Out_C != nil {
		return true
	}
	return false
}

func (self *OutState) Clone() (ret OutState) {
	utils.DeepCopy(&ret, self)
	return
}

func (self *OutState) ToPKr() *c_type.PKr {
	if self.Out_O != nil {
		return &self.Out_O.Addr
	} else if self.Out_Z != nil {
		return &self.Out_Z.PKr
	} else if self.Out_P != nil {
		return &self.Out_P.PKr
	} else if self.Out_C != nil {
		return &self.Out_C.PKr
	}
	return nil
}

func (self *OutState) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type OutState0Get struct {
	Out *OutState
}

func (self *OutState0Get) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.Out = nil
		return
	} else {
		self.Out = &OutState{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
