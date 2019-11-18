package exchange

import (
	"io"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/core/types/vserial"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/assets"
)

type Utxo_Version0 struct {
	Pkr    c_type.PKr
	Root   c_type.Uint256
	TxHash c_type.Uint256
	Nil    c_type.Uint256
	Num    uint64
	Asset  assets.Asset
	IsZ    bool
}

type Utxo_Version1 struct {
	Ignore bool
}

func (b *Utxo) DecodeRLP(s *rlp.Stream) error {
	vs := vserial.NewVSerial()
	v0 := Utxo_Version0{}
	v1 := Utxo_Version1{}

	vs.Add(&v0, vserial.VERSION_0)
	vs.Add(&v1, vserial.VERSION_1)
	if e := s.Decode(&vs); e != nil {
		return e
	}

	if vs.V() <= vserial.VERSION_1 {
		SetUtxoForVersion0(b, &v0)
		if vs.V() >= vserial.VERSION_1 {
			SetUtxoForVersion1(b, &v1)
		}
	}
	return nil
}

func SetVersion0ForUtxo(v0 *Utxo_Version0, b *Utxo) {
	v0.Pkr = b.Pkr
	v0.Root = b.Root
	v0.TxHash = b.TxHash
	v0.Nil = b.Nil
	v0.Num = b.Num
	v0.Asset = b.Asset
	v0.IsZ = b.IsZ
}

func SetUtxoForVersion0(b *Utxo, v0 *Utxo_Version0) {
	b.Pkr = v0.Pkr
	b.Root = v0.Root
	b.TxHash = v0.TxHash
	b.Nil = v0.Nil
	b.Num = v0.Num
	b.Asset = v0.Asset
	b.IsZ = v0.IsZ
}

func SetVersion1ForUtxo(v1 *Utxo_Version1, b *Utxo) {
	v1.Ignore = b.Ignore
}

func SetUtxoForVersion1(b *Utxo, v1 *Utxo_Version1) {
	b.Ignore = v1.Ignore
}

func (b *Utxo) EncodeRLP(w io.Writer) error {
	vs := vserial.NewVSerial()

	v0 := Utxo_Version0{}
	SetVersion0ForUtxo(&v0, b)
	vs.Add(&v0, vserial.VERSION_0)

	if b.Ignore {
		v1 := Utxo_Version1{}
		SetVersion1ForUtxo(&v1, b)
		vs.Add(&v1, vserial.VERSION_1)
	}

	return rlp.Encode(w, &vs)
}
