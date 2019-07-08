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

//go:generate gencodec -type Version_0 -field-override headerMarshaling -out gen_header_json.go

// Version_0 represents a block header in the Ethereum blockchain.
type Header struct {
	//Data
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
	//POW
	MixDigest common.Hash `json:"mixHash"          gencodec:"required"`
	Nonce     BlockNonce  `json:"nonce"            gencodec:"required"`
	//POS
	CurrentVotes []typeserial.Vote
	ParentVotes  []typeserial.Vote
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
	return rlpHash(h)
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashPow() common.Hash {
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

func (h *Header) HashPos() (ret common.Hash) {
	m := sha3.NewKeccak256()
	m.Write(h.MixDigest[:])
	m.Write(h.Nonce[:])
	hp := m.Sum(nil)
	copy(ret[:], hp)
	return
}

func StakeHash(currentPosHash *common.Hash, parentPosHash *common.Hash) (ret common.Hash) {
	m := sha3.NewKeccak256()
	m.Write(currentPosHash[:])
	m.Write(parentPosHash[:])
	sh := m.Sum(nil)
	copy(ret[:], sh)
	return
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
	if len(h.CurrentVotes) > 0 {
		cpy.CurrentVotes = append([]typeserial.Vote{}, h.CurrentVotes...)
	}
	if len(h.ParentVotes) > 0 {
		cpy.CurrentVotes = append([]typeserial.Vote{}, h.ParentVotes...)
	}
	return &cpy
}

func HeaderToTypeSerial(b *Header) (hr typeserial.HeaderRLP) {
	//Version_0
	hr.Version_0.ParentHash = b.ParentHash
	hr.Version_0.Coinbase = b.Coinbase
	hr.Version_0.Licr = b.Licr
	hr.Version_0.Root = b.Root
	hr.Version_0.TxHash = b.TxHash
	hr.Version_0.ReceiptHash = b.ReceiptHash
	hr.Version_0.Bloom = b.Bloom
	hr.Version_0.Difficulty = b.Difficulty
	hr.Version_0.Number = b.Number
	hr.Version_0.GasLimit = b.GasLimit
	hr.Version_0.GasUsed = b.GasUsed
	hr.Version_0.Time = b.Time
	hr.Version_0.Extra = b.Extra
	hr.Version_0.MixDigest = b.MixDigest
	hr.Version_0.Nonce = b.Nonce

	//Version_1
	if len(b.CurrentVotes) > 0 || len(b.ParentVotes) > 0 {
		hr.Version_1.CurrentVotes = b.CurrentVotes
		hr.Version_1.ParentVotes = b.ParentVotes
		//Version Number
		hr.Version.V = typeserial.VERSION_1
	} else {
		hr.Version.V = typeserial.VERSION_0
	}
	return
}

func TypeSerialToHeader(hr *typeserial.HeaderRLP) (b Header) {
	b.ParentHash = hr.Version_0.ParentHash
	b.Coinbase = hr.Version_0.Coinbase
	b.Licr = hr.Version_0.Licr
	b.Root = hr.Version_0.Root
	b.TxHash = hr.Version_0.TxHash
	b.ReceiptHash = hr.Version_0.ReceiptHash
	b.Bloom = hr.Version_0.Bloom
	b.Difficulty = hr.Version_0.Difficulty
	b.Number = hr.Version_0.Number
	b.GasLimit = hr.Version_0.GasLimit
	b.GasUsed = hr.Version_0.GasUsed
	b.Time = hr.Version_0.Time
	b.Extra = hr.Version_0.Extra
	b.MixDigest = hr.Version_0.MixDigest
	b.Nonce = hr.Version_0.Nonce
	if hr.Version.V == typeserial.VERSION_1 {
		b.CurrentVotes = hr.Version_1.CurrentVotes
		b.ParentVotes = hr.Version_1.ParentVotes
	}
	return
}

// DecodeRLP decodes the Ethereum
func (b *Header) DecodeRLP(s *rlp.Stream) error {
	hr := typeserial.HeaderRLP{}
	if e := s.Decode(&hr); e != nil {
		return e
	}
	*b = TypeSerialToHeader(&hr)
	return nil
}

// EncodeRLP serializes b into the Ethereum RLP block format.
func (b *Header) EncodeRLP(w io.Writer) error {
	hr := HeaderToTypeSerial(b)
	return rlp.Encode(w, &hr)
}
