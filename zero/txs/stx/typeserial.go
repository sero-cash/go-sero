package stx

import (
	"errors"
	"io"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/core/types/vserial"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

type ZtxVersion_0 struct {
	Ehash    keys.Uint256
	From     keys.PKr
	Fee      assets.Token
	Sign     keys.Uint512
	Bcr      keys.Uint256
	Bsign    keys.Uint512
	Desc_Z   Desc_Z
	Desc_O   Desc_O
	Desc_Pkg PkgDesc_Z
}

type ZtxVersion_1 struct {
	BuyShare   *BuyShareCmd   `rlp:"nil"`
	RegistPool *RegistPoolCmd `rlp:"nil"`
	ClosePool  *ClosePoolCmd  `rlp:"nil"`
	Contract   *ContractCmd   `rlp:"nil"`
}

type ZtxRlp struct {
	Version   vserial.Version
	Version_0 ZtxVersion_0
	Version_1 ZtxVersion_1
}

func (self *ZtxRlp) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	if size == 0 {
		self.Version.V = vserial.VERSION_NIL
	} else {
		if size > 10 {
			self.Version.V = vserial.VERSION_0
		} else {
			if e := s.Decode(&self.Version); e != nil {
				return e
			}
		}
	}
	if e := s.Decode(&self.Version_0); e != nil {
		return e
	}
	if self.Version.V >= vserial.VERSION_1 {
		if e := s.Decode(&self.Version_1); e != nil {
			return e
		}
	}
	return nil
}

func (self *ZtxRlp) EncodeRLP(w io.Writer) error {
	if self.Version.V == vserial.VERSION_NIL {
		e := errors.New("encode header rlp error: version is nil")
		panic(e)
		return e
	}
	if self.Version.V >= vserial.VERSION_1 {
		if e := rlp.Encode(w, &self.Version); e != nil {
			return e
		}
	}
	if e := rlp.Encode(w, &self.Version_0); e != nil {
		return e
	}
	if self.Version.V >= vserial.VERSION_1 {
		if e := rlp.Encode(w, &self.Version_1); e != nil {
			return e
		}
	}
	return nil
}

func (b *T) DecodeRLP(s *rlp.Stream) error {
	hr := ZtxRlp{}
	if e := s.Decode(&hr); e != nil {
		return e
	}

	b.Ehash = hr.Version_0.Ehash
	b.From = hr.Version_0.From
	b.Fee = hr.Version_0.Fee
	b.Sign = hr.Version_0.Sign
	b.Bcr = hr.Version_0.Bcr
	b.Bsign = hr.Version_0.Bsign
	b.Desc_Z = hr.Version_0.Desc_Z
	b.Desc_O = hr.Version_0.Desc_O
	b.Desc_Pkg = hr.Version_0.Desc_Pkg
	b.Desc_Cmd.BuyShare = hr.Version_1.BuyShare
	b.Desc_Cmd.RegistPool = hr.Version_1.RegistPool
	b.Desc_Cmd.ClosePool = hr.Version_1.ClosePool
	b.Desc_Cmd.Contract = hr.Version_1.Contract

	return nil
}

// EncodeRLP serializes b into the Ethereum RLP block format.
func (b *T) EncodeRLP(w io.Writer) error {
	hr := ZtxRlp{}
	if b.Desc_Cmd.Count() > 0 {
		hr.Version.V = vserial.VERSION_1
	} else {
		hr.Version.V = vserial.VERSION_0
	}
	hr.Version_0.Ehash = b.Ehash
	hr.Version_0.From = b.From
	hr.Version_0.Fee = b.Fee
	hr.Version_0.Sign = b.Sign
	hr.Version_0.Bcr = b.Bcr
	hr.Version_0.Bsign = b.Bsign
	hr.Version_0.Desc_Z = b.Desc_Z
	hr.Version_0.Desc_O = b.Desc_O
	hr.Version_0.Desc_Pkg = b.Desc_Pkg
	hr.Version_1.BuyShare = b.Desc_Cmd.BuyShare
	hr.Version_1.RegistPool = b.Desc_Cmd.RegistPool
	hr.Version_1.ClosePool = b.Desc_Cmd.ClosePool
	hr.Version_1.Contract = b.Desc_Cmd.Contract
	return rlp.Encode(w, &hr)
}
