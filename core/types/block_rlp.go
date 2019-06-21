package types

import (
	"io"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/rlp"
)

// "external" block encoding. used for sero protocol, etc.
type Block_Version_0 struct {
	Header *Header
	Txs    []*Transaction
}

// DecodeRLP decodes the Ethereum
func (b *Block) DecodeRLP(s *rlp.Stream) error {
	b0 := Block_Version_0{}

	_, size, _ := s.Kind()
	if err := s.Decode(&b0); err != nil {
		return err
	}

	b.header = b0.Header
	b.transactions = b0.Txs

	b.size.Store(common.StorageSize(rlp.ListSize(size)))
	return nil
}

// EncodeRLP serializes b into the Ethereum RLP block format.
func (b *Block) EncodeRLP(w io.Writer) error {
	b0 := Block_Version_0{}
	b0.Header = b.header
	b0.Txs = b.transactions
	return rlp.Encode(w, b0)
}
