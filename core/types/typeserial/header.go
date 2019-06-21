package typeserial

import (
	"io"
	"math/big"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/rlp"
)

const (
	BloomByteLength = 256
)

type Header struct {
	ParentHash  common.Hash
	Coinbase    common.Address
	Licr        keys.LICr
	Root        common.Hash
	TxHash      common.Hash
	ReceiptHash common.Hash
	Bloom       [BloomByteLength]byte
	Difficulty  *big.Int
	Number      *big.Int
	GasLimit    uint64
	GasUsed     uint64
	Time        *big.Int
	Extra       []byte
	MixDigest   common.Hash
	Nonce       [8]byte
}

type HeaderRLP struct {
	Header  *Header
	Version uint64
}

func (self *HeaderRLP) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	if size > 300 || size == 0 {
		self.Header = &Header{}
		if e := s.Decode(self.Header); e != nil {
			return e
		}
		return nil
	} else {
		return errors.New("headerRLP decode error: unknow version")
	}
}

func (self *HeaderRLP) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, self.Header)
}
