package localdb

import (
	"io"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/core/types/vserial"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v2"
)

type OutState_Version0 struct {
	Index  uint64
	Out_O  *stx_v1.Out_O   `rlp:"nil"`
	Out_Z  *stx_v1.Out_Z   `rlp:"nil"`
	OutCM  *c_type.Uint256 `rlp:"nil"`
	RootCM *c_type.Uint256 `rlp:"nil"`
}

type OutState_Version1 struct {
	Out_P *stx_v2.Out_P `rlp:"nil"`
	Out_C *stx_v2.Out_C `rlp:"nil"`
}

func (b *OutState) DecodeRLP(s *rlp.Stream) error {
	vs := vserial.NewVSerial()
	v0 := OutState_Version0{}
	v1 := OutState_Version1{}

	vs.Add(&v0, vserial.VERSION_0)
	vs.Add(&v1, vserial.VERSION_1)
	if e := s.Decode(&vs); e != nil {
		return e
	}

	if vs.V() <= vserial.VERSION_1 {
		SetOSForVersion0(b, &v0)
		if vs.V() >= vserial.VERSION_1 {
			SetOSForVersion1(b, &v1)
		}
	}
	return nil
}

func SetVersion0ForOS(v0 *OutState_Version0, b *OutState) {
	v0.Index = b.Index
	v0.Out_O = b.Out_O
	v0.Out_Z = b.Out_Z
	v0.OutCM = b.OutCM
	v0.RootCM = b.RootCM
}

func SetOSForVersion0(b *OutState, v0 *OutState_Version0) {
	b.Index = v0.Index
	b.Out_O = v0.Out_O
	b.Out_Z = v0.Out_Z
	b.OutCM = v0.OutCM
	b.RootCM = v0.RootCM
}

func SetVersion1ForOS(v1 *OutState_Version1, b *OutState) {
	v1.Out_P = b.Out_P
	v1.Out_C = b.Out_C
}

func SetOSForVersion1(b *OutState, v1 *OutState_Version1) {
	b.Out_P = v1.Out_P
	b.Out_C = v1.Out_C
}

func (b *OutState) EncodeRLP(w io.Writer) error {
	vs := vserial.NewVSerial()

	v0 := OutState_Version0{}
	SetVersion0ForOS(&v0, b)
	vs.Add(&v0, vserial.VERSION_0)

	if b.IsSzk() {
		v1 := OutState_Version1{}
		SetVersion1ForOS(&v1, b)
		vs.Add(&v1, vserial.VERSION_1)
	}

	return rlp.Encode(w, &vs)
}
