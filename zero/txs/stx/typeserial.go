package stx

import (
	"io"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/core/types/vserial"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

type ZtxVersion_0 struct {
	Ehash    c_type.Uint256
	From     c_type.PKr
	Fee      assets.Token
	Sign     c_type.Uint512
	Bcr      c_type.Uint256
	Bsign    c_type.Uint512
	Desc_Z   stx_v0.Desc_Z
	Desc_O   stx_v0.Desc_O
	Desc_Pkg PkgDesc_Z
}

type ZtxVersion_1 struct {
	BuyShare   *BuyShareCmd   `rlp:"nil"`
	RegistPool *RegistPoolCmd `rlp:"nil"`
	ClosePool  *ClosePoolCmd  `rlp:"nil"`
	Contract   *ContractCmd   `rlp:"nil"`
}

type ZtxVersion_2 struct {
	Version0 ZtxVersion_0
	Version1 ZtxVersion_1
	Tx2      stx_v1.Tx
}

func SetTxForVersion0(b *T, v0 *ZtxVersion_0) {
	b.Ehash = v0.Ehash
	b.From = v0.From
	b.Fee = v0.Fee
	b.Sign = v0.Sign
	b.Bcr = v0.Bcr
	b.Bsign = v0.Bsign
	b.Desc_Pkg = v0.Desc_Pkg
	b.Desc_Z = v0.Desc_Z
	b.Desc_O = v0.Desc_O
}

func SetTxForVersion1(b *T, v1 *ZtxVersion_1) {
	b.Desc_Cmd.BuyShare = v1.BuyShare
	b.Desc_Cmd.RegistPool = v1.RegistPool
	b.Desc_Cmd.ClosePool = v1.ClosePool
	b.Desc_Cmd.Contract = v1.Contract
}

func SetTxForVersion2(b *T, v2 *ZtxVersion_2) {
	SetTxForVersion0(b, &v2.Version0)
	SetTxForVersion1(b, &v2.Version1)
	b.Tx1 = &v2.Tx2
}

func (b *T) DecodeRLP(s *rlp.Stream) error {
	vs := vserial.NewVSerial()
	v0 := ZtxVersion_0{}
	v1 := ZtxVersion_1{}
	v2 := ZtxVersion_2{}

	vs.Add(&v0, vserial.VERSION_0)
	vs.Add(&v1, vserial.VERSION_1)
	vs.Add(&v2, vserial.VERSION_2)
	if e := s.Decode(&vs); e != nil {
		return e
	}

	if vs.V() <= vserial.VERSION_1 {
		SetTxForVersion0(b, &v0)
		if vs.V() >= vserial.VERSION_1 {
			SetTxForVersion1(b, &v1)
		}
	} else if vs.V() >= vserial.VERSION_2 {
		SetTxForVersion2(b, &v2)
	}
	return nil
}

func SetVersion0ForTx(v0 *ZtxVersion_0, b *T) {
	v0.Ehash = b.Ehash
	v0.From = b.From
	v0.Fee = b.Fee
	v0.Sign = b.Sign
	v0.Bcr = b.Bcr
	v0.Bsign = b.Bsign
	v0.Desc_Z = b.Desc_Z
	v0.Desc_O = b.Desc_O
	v0.Desc_Pkg = b.Desc_Pkg
}

func SetVersion1ForTx(v1 *ZtxVersion_1, b *T) {
	v1.BuyShare = b.Desc_Cmd.BuyShare
	v1.RegistPool = b.Desc_Cmd.RegistPool
	v1.ClosePool = b.Desc_Cmd.ClosePool
	v1.Contract = b.Desc_Cmd.Contract
}

func SetVersion2ForTx(v2 *ZtxVersion_2, b *T) {
	SetVersion0ForTx(&v2.Version0, b)
	SetVersion1ForTx(&v2.Version1, b)
	v2.Tx2 = *b.Tx1
}

// EncodeRLP serializes b into the Ethereum RLP block format.
func (b *T) EncodeRLP(w io.Writer) error {
	vs := vserial.NewVSerial()

	if b.Tx1 != nil {
		v2 := ZtxVersion_2{}
		SetVersion2ForTx(&v2, b)
		vs.Add(&v2, vserial.VERSION_2)
		return rlp.Encode(w, &vs)
	}

	v0 := ZtxVersion_0{}
	SetVersion0ForTx(&v0, b)
	vs.Add(&v0, vserial.VERSION_0)

	if b.Desc_Cmd.Count() > 0 {
		v1 := ZtxVersion_1{}
		SetVersion1ForTx(&v1, b)
		vs.Add(&v1, vserial.VERSION_1)
	}

	return rlp.Encode(w, &vs)
}
