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

//version 0
type Version_0 struct {
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

//version 1
type Vote struct {
	Index uint8
	Sign  keys.Uint512
}

type Version_1 struct {
	CurrentVotes []Vote
	ParentVotes  []Vote
}

func (self *Version_1) Valid() bool {
	if len(self.CurrentVotes) > 0 {
		return true
	}
	if len(self.ParentVotes) > 0 {
		return true
	}
	return false
}

type VersionType int8

const (
	VERSION_NIL = VersionType(-1)
	VERSION_0   = VersionType(0)
	VERSION_1   = VersionType(1)
)

type Version struct {
	V VersionType
}

type HeaderRLP struct {
	Version Version
	Version_0
	Version_1
}

func (self *HeaderRLP) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	if size == 0 {
		self.Version.V = VERSION_NIL
	} else {
		if size > 10 {
			self.Version.V = VERSION_0
		} else {
			if e := s.Decode(&self.Version); e != nil {
				return e
			}
		}
	}
	if e := s.Decode(&self.Version_0); e != nil {
		return e
	}
	if self.Version.V >= VERSION_1 {
		if e := s.Decode(&self.Version_1); e != nil {
			return e
		}
	}
	return nil
}

func (self *HeaderRLP) EncodeRLP(w io.Writer) error {
	if self.Version.V == VERSION_NIL {
		e := errors.New("encode header rlp error: version is nil")
		panic(e)
		return e
	}
	if self.Version.V >= VERSION_1 {
		if e := rlp.Encode(w, &self.Version); e != nil {
			return e
		}
	}
	if e := rlp.Encode(w, &self.Version_0); e != nil {
		return e
	}
	if self.Version.V >= VERSION_1 {
		if e := rlp.Encode(w, &self.Version_1); e != nil {
			return e
		}
	}
	return nil
}
