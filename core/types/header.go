package types

import (
	"encoding/binary"
	"io"
	"math/big"
	"unsafe"

	"github.com/sero-cash/go-sero/core/types/typeserial"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/rlp"
)

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() ([]byte, error) {
	return hexutil.Bytes(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *BlockNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("BlockNonce", input, n[:])
}

//go:generate gencodec -type Header -field-override headerMarshaling -out gen_header_json.go

// Header represents a block header in the Ethereum blockchain.
type Header struct {
	//version0
	ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
	Coinbase    common.Address `json:"miner"            gencodec:"required"`
	Licr        keys.LICr      `json:"licr"            	gencodec:"required"`
	Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int       `json:"difficulty"       gencodec:"required"`
	Number      *big.Int       `json:"number"           gencodec:"required"`
	GasLimit    uint64         `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64         `json:"gasUsed"          gencodec:"required"`
	Time        *big.Int       `json:"timestamp"        gencodec:"required"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash    `json:"mixHash"          gencodec:"required"`
	Nonce       BlockNonce     `json:"nonce"            gencodec:"required"`
	//version1
	Test uint64
}

// field type overrides for gencodec
type headerMarshaling struct {
	Difficulty *hexutil.Big
	Number     *hexutil.Big
	GasLimit   hexutil.Uint64
	GasUsed    hexutil.Uint64
	Time       *hexutil.Big
	Extra      hexutil.Bytes
	Hash       common.Hash `json:"hash"` // adds call to Hash() in MarshalJSON
}

func (h *Header) Valid() bool {
	if h.Licr.H == 0 {
		return h.Number.Uint64() >= h.Licr.L
	} else {
		return h.Number.Uint64() >= h.Licr.L && h.Number.Uint64() <= h.Licr.H
	}
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash {
	//test
	return rlpHash(h)
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() common.Hash {
	return rlpHash([]interface{}{
		h.ParentHash,
		h.Coinbase,
		h.Root,
		h.TxHash,
		h.ReceiptHash,
		h.Bloom,
		h.Difficulty,
		h.Number,
		h.GasLimit,
		h.GasUsed,
		h.Time,
		h.Extra,
	})
}

func (h *Header) ActualDifficulty() *big.Int {
	if h.Valid() {
		c := new(big.Int).SetUint64(h.Licr.C)
		if h.Difficulty.Cmp(c) > 0 {
			return new(big.Int).Sub(h.Difficulty, c)
		} else {
			return big.NewInt(1)
		}
	} else {
		return maxUint256
	}
}

// Size returns the approximate memory used by all internal contents. It is used
// to approximate and limit the memory consumption of various caches.
func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)+(h.Difficulty.BitLen()+h.Number.BitLen()+h.Time.BitLen())/8)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(big.Int); h.Time != nil {
		cpy.Time.Set(h.Time)
	}
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

// DecodeRLP decodes the Ethereum
func (b *Header) DecodeRLP(s *rlp.Stream) error {
	hr := typeserial.HeaderRLP{}
	if e := s.Decode(&hr); e != nil {
		return e
	}
	b.ParentHash = hr.Header.ParentHash
	b.Coinbase = hr.Header.Coinbase
	b.Licr = hr.Header.Licr
	b.Root = hr.Header.Root
	b.TxHash = hr.Header.TxHash
	b.ReceiptHash = hr.Header.ReceiptHash
	b.Bloom = hr.Header.Bloom
	b.Difficulty = hr.Header.Difficulty
	b.Number = hr.Header.Number
	b.GasLimit = hr.Header.GasLimit
	b.GasUsed = hr.Header.GasUsed
	b.Time = hr.Header.Time
	b.Extra = hr.Header.Extra
	b.MixDigest = hr.Header.MixDigest
	b.Nonce = hr.Header.Nonce
	return nil
}

// EncodeRLP serializes b into the Ethereum RLP block format.
func (b *Header) EncodeRLP(w io.Writer) error {
	hr := typeserial.HeaderRLP{}
	hr.Header = &typeserial.Header{}
	hr.Header.ParentHash = b.ParentHash
	hr.Header.Coinbase = b.Coinbase
	hr.Header.Licr = b.Licr
	hr.Header.Root = b.Root
	hr.Header.TxHash = b.TxHash
	hr.Header.ReceiptHash = b.ReceiptHash
	hr.Header.Bloom = b.Bloom
	hr.Header.Difficulty = b.Difficulty
	hr.Header.Number = b.Number
	hr.Header.GasLimit = b.GasLimit
	hr.Header.GasUsed = b.GasUsed
	hr.Header.Time = b.Time
	hr.Header.Extra = b.Extra
	hr.Header.MixDigest = b.MixDigest
	hr.Header.Nonce = b.Nonce
	return rlp.Encode(w, &hr)
}
