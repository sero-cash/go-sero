package stx

import (
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

func (b *T) DecodeRLP(s *rlp.Stream) error {
	vs := vserial.VSerial{}
	v0 := ZtxVersion_0{}
	v1 := ZtxVersion_1{}

	vs.Versions = append(vs.Versions, &v0)
	vs.Versions = append(vs.Versions, &v1)
	if e := s.Decode(&vs); e != nil {
		return e
	}

	b.Ehash = v0.Ehash
	b.From = v0.From
	b.Fee = v0.Fee
	b.Sign = v0.Sign
	b.Bcr = v0.Bcr
	b.Bsign = v0.Bsign
	b.Desc_Z = v0.Desc_Z
	b.Desc_O = v0.Desc_O
	b.Desc_Pkg = v0.Desc_Pkg
	b.Desc_Cmd.BuyShare = v1.BuyShare
	b.Desc_Cmd.RegistPool = v1.RegistPool
	b.Desc_Cmd.ClosePool = v1.ClosePool
	b.Desc_Cmd.Contract = v1.Contract

	return nil
}

// EncodeRLP serializes b into the Ethereum RLP block format.
func (b *T) EncodeRLP(w io.Writer) error {
	vs := vserial.VSerial{}

	v0 := ZtxVersion_0{}
	v0.Ehash = b.Ehash
	v0.From = b.From
	v0.Fee = b.Fee
	v0.Sign = b.Sign
	v0.Bcr = b.Bcr
	v0.Bsign = b.Bsign
	v0.Desc_Z = b.Desc_Z
	v0.Desc_O = b.Desc_O
	v0.Desc_Pkg = b.Desc_Pkg
	vs.Versions = append(vs.Versions, &v0)

	if b.Desc_Cmd.Count() > 0 {
		v1 := ZtxVersion_1{}
		v1.BuyShare = b.Desc_Cmd.BuyShare
		v1.RegistPool = b.Desc_Cmd.RegistPool
		v1.ClosePool = b.Desc_Cmd.ClosePool
		v1.Contract = b.Desc_Cmd.Contract
		vs.Versions = append(vs.Versions, &v1)
	}

	return rlp.Encode(w, &vs)
}
