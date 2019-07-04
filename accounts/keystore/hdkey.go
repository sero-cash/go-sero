package keystore

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"math/big"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/address"

	"github.com/pborman/uuid"
	"github.com/tyler-smith/go-bip39"

	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/crypto"

	"github.com/btcsuite/btcd/btcec"

	"golang.org/x/crypto/ripemd160"
)

const (

	// HardenedKeyStart is the index at which a hardended key starts.  Each
	// extended key has 2^31 normal child keys and 2^31 hardned child keys.
	// Thus the range for normal child keys is [0, 2^31 - 1] and the range
	// for hardened child keys is [2^31, 2^32 - 1].
	HardenedKeyStart = 0x80000000 // 2^31

	// MinSeedBytes is the minimum number of bytes allowed for a seed to
	// a master node.
	MinSeedBytes = 16 // 128 bits

	// MaxSeedBytes is the maximum number of bytes allowed for a seed to
	// a master node.
	MaxSeedBytes = 64 // 512 bits

	// serializedKeyLen is the length of a serialized public or private
	// extended key.  It consists of 4 bytes version, 1 byte depth, 4 bytes
	// fingerprint, 4 bytes child number, 32 bytes chain code, and 33 bytes
	// public/private key data.
	serializedKeyLen = 4 + 1 + 4 + 4 + 32 + 33 // 78 bytes

	// maxUint8 is the max positive integer which can be serialized in a uint8
	maxUint8 = 1<<8 - 1
)

var (
	// ErrDeriveHardFromPublic describes an error in which the caller
	// attempted to derive a hardened extended key from a public key.
	ErrDeriveHardFromPublic = errors.New("cannot derive a hardened key " +
		"from a public key")

	// ErrDeriveBeyondMaxDepth describes an error in which the caller
	// has attempted to derive more than 255 keys from a root key.
	ErrDeriveBeyondMaxDepth = errors.New("cannot derive a key with more than " +
		"255 indices in its path")

	// ErrNotPrivExtKey describes an error in which the caller attempted
	// to extract a private key from a public extended key.
	ErrNotPrivExtKey = errors.New("unable to create private keys from a " +
		"public extended key")

	// ErrInvalidChild describes an error in which the child at a specific
	// index is invalid due to the derived key falling outside of the valid
	// range for secp256k1 private keys.  This error indicates the caller
	// should simply ignore the invalid child extended key at this index and
	// increment to the next index.
	ErrInvalidChild = errors.New("the extended key at this index is invalid")

	// ErrUnusableSeed describes an error in which the provided seed is not
	// usable due to the derived key falling outside of the valid range for
	// secp256k1 private keys.  This error indicates the caller must choose
	// another seed.
	ErrUnusableSeed = errors.New("unusable seed")

	// ErrInvalidSeedLen describes an error in which the provided seed or
	// seed length is not in the allowed range.
	ErrInvalidSeedLen = fmt.Errorf("seed length must be between %d and %d "+
		"bits", MinSeedBytes*8, MaxSeedBytes*8)

	// ErrBadChecksum describes an error in which the checksum encoded with
	// a serialized extended key does not match the calculated value.
	ErrBadChecksum = errors.New("bad extended key checksum")

	// ErrInvalidKeyLen describes an error in which the provided serialized
	// key is not the expected length.
	ErrInvalidKeyLen = errors.New("the provided serialized extended key " +
		"length is invalid")
)

// masterKey is the master key used along with a random seed used to generate
// the master node in the hierarchical tree.
var masterKey = []byte("Bitcoin seed")

var HDPrivateKeyID = [4]byte{0x04, 0x88, 0xad, 0xe4}
var HDPublicKeyID = [4]byte{0x04, 0x88, 0xb2, 0x1e} // starts with xpub

// HDKey houses all the information needed to support a hierarchical
// deterministic extended key.  See the package overview documentation for
// more details on how to use extended keys.
type HDKey struct {
	key       []byte // This will be the pubkey for extended pub keys
	pubKey    []byte
	chainCode []byte
	depth     uint8
	parentFP  []byte
	childNum  uint32
	version   []byte
	isPrivate bool
}

// NewHDKey returns a new instance of an extended key with the given
// fields.  No error checking is performed here as it's only intended to be a
// convenience method used to create a populated struct. This function should
// only by used by applications that need to create custom HDKeys. All
// other applications should just use NewMaster, Child, or Neuter.
func NewHDKey(version, key, chainCode, parentFP []byte, depth uint8,
	childNum uint32, isPrivate bool) *HDKey {

	// NOTE: The pubKey field is intentionally left nil so it is only
	// computed and memoized as required.
	return &HDKey{
		key:       key,
		chainCode: chainCode,
		depth:     depth,
		parentFP:  parentFP,
		childNum:  childNum,
		version:   version,
		isPrivate: isPrivate,
	}
}

func (k *HDKey) pubKeyBytes() []byte {
	// Just return the key if it's already an extended public key.
	if !k.isPrivate {
		return k.key
	}

	// This is a private extended key, so calculate and memoize the public
	// key if needed.
	if len(k.pubKey) == 0 {
		pkx, pky := btcec.S256().ScalarBaseMult(k.key)
		pubKey := btcec.PublicKey{Curve: btcec.S256(), X: pkx, Y: pky}
		k.pubKey = pubKey.SerializeCompressed()
	}

	return k.pubKey
}

// IsPrivate returns whether or not the extended key is a private extended key.
//
// A private extended key can be used to derive both hardened and non-hardened
// child private and public extended keys.  A public extended key can only be
// used to derive non-hardened child public extended keys.
func (k *HDKey) IsPrivate() bool {
	return k.isPrivate
}

// Depth returns the current derivation level with respect to the root.
//
// The root key has depth zero, and the field has a maximum of 255 due to
// how depth is serialized.
func (k *HDKey) Depth() uint8 {
	return k.depth
}

// ParentFingerprint returns a fingerprint of the parent extended key from which
// this one was derived.
func (k *HDKey) ParentFingerprint() uint32 {
	return binary.BigEndian.Uint32(k.parentFP)
}

// Child returns a derived child extended key at the given index.  When this
// extended key is a private extended key (as determined by the IsPrivate
// function), a private extended key will be derived.  Otherwise, the derived
// extended key will be also be a public extended key.
//
// When the index is greater to or equal than the HardenedKeyStart constant, the
// derived extended key will be a hardened extended key.  It is only possible to
// derive a hardended extended key from a private extended key.  Consequently,
// this function will return ErrDeriveHardFromPublic if a hardened child
// extended key is requested from a public extended key.
//
// A hardened extended key is useful since, as previously mentioned, it requires
// a parent private extended key to derive.  In other words, normal child
// extended public keys can be derived from a parent public extended key (no
// knowledge of the parent private key) whereas hardened extended keys may not
// be.
//
// NOTE: There is an extremely small chance (< 1 in 2^127) the specific child
// index does not derive to a usable child.  The ErrInvalidChild error will be
// returned if this should occur, and the caller is expected to ignore the
// invalid child and simply increment to the next index.
func (k *HDKey) Child(i uint32) (*HDKey, error) {
	// Prevent derivation of children beyond the max allowed depth.
	if k.depth == maxUint8 {
		return nil, ErrDeriveBeyondMaxDepth
	}

	// There are four scenarios that could happen here:
	// 1) Private extended key -> Hardened child private extended key
	// 2) Private extended key -> Non-hardened child private extended key
	// 3) Public extended key -> Non-hardened child public extended key
	// 4) Public extended key -> Hardened child public extended key (INVALID!)

	// Case #4 is invalid, so error out early.
	// A hardened child extended key may not be created from a public
	// extended key.
	isChildHardened := i >= HardenedKeyStart
	if !k.isPrivate && isChildHardened {
		return nil, ErrDeriveHardFromPublic
	}

	// The data used to derive the child key depends on whether or not the
	// child is hardened per [BIP32].
	//
	// For hardened children:
	//   0x00 || ser256(parentKey) || ser32(i)
	//
	// For normal children:
	//   serP(parentPubKey) || ser32(i)
	keyLen := 33
	data := make([]byte, keyLen+4)
	if isChildHardened {
		// Case #1.
		// When the child is a hardened child, the key is known to be a
		// private key due to the above early return.  Pad it with a
		// leading zero as required by [BIP32] for deriving the child.
		copy(data[1:], k.key)
	} else {
		// Case #2 or #3.
		// This is either a public or private extended key, but in
		// either case, the data which is used to derive the child key
		// starts with the secp256k1 compressed public key bytes.
		copy(data, k.pubKeyBytes())
	}
	binary.BigEndian.PutUint32(data[keyLen:], i)

	// Take the HMAC-SHA512 of the current key's chain code and the derived
	// data:
	//   I = HMAC-SHA512(Key = chainCode, Data = data)
	hmac512 := hmac.New(sha512.New, k.chainCode)
	hmac512.Write(data)
	ilr := hmac512.Sum(nil)

	// Split "I" into two 32-byte sequences Il and Ir where:
	//   Il = intermediate key used to derive the child
	//   Ir = child chain code
	il := ilr[:len(ilr)/2]
	childChainCode := ilr[len(ilr)/2:]

	// Both derived public or private keys rely on treating the left 32-byte
	// sequence calculated above (Il) as a 256-bit integer that must be
	// within the valid range for a secp256k1 private key.  There is a small
	// chance (< 1 in 2^127) this condition will not hold, and in that case,
	// a child extended key can't be created for this index and the caller
	// should simply increment to the next index.
	ilNum := new(big.Int).SetBytes(il)
	if ilNum.Cmp(btcec.S256().N) >= 0 || ilNum.Sign() == 0 {
		return nil, ErrInvalidChild
	}

	// The algorithm used to derive the child key depends on whether or not
	// a private or public child is being derived.
	//
	// For private children:
	//   childKey = parse256(Il) + parentKey
	//
	// For public children:
	//   childKey = serP(point(parse256(Il)) + parentKey)
	var isPrivate bool
	var childKey []byte
	if k.isPrivate {
		// Case #1 or #2.
		// Add the parent private key to the intermediate private key to
		// derive the final child key.
		//
		// childKey = parse256(Il) + parenKey
		keyNum := new(big.Int).SetBytes(k.key)
		ilNum.Add(ilNum, keyNum)
		ilNum.Mod(ilNum, btcec.S256().N)
		childKey = ilNum.Bytes()
		isPrivate = true
	} else {
		// Case #3.
		// Calculate the corresponding intermediate public key for
		// intermediate private key.
		ilx, ily := btcec.S256().ScalarBaseMult(il)
		if ilx.Sign() == 0 || ily.Sign() == 0 {
			return nil, ErrInvalidChild
		}

		// Convert the serialized compressed parent public key into X
		// and Y coordinates so it can be added to the intermediate
		// public key.
		pubKey, err := btcec.ParsePubKey(k.key, btcec.S256())
		if err != nil {
			return nil, err
		}

		// Add the intermediate public key to the parent public key to
		// derive the final child key.
		//
		// childKey = serP(point(parse256(Il)) + parentKey)
		childX, childY := btcec.S256().Add(ilx, ily, pubKey.X, pubKey.Y)
		pk := btcec.PublicKey{Curve: btcec.S256(), X: childX, Y: childY}
		childKey = pk.SerializeCompressed()
	}

	// The fingerprint of the parent for the derived child is the first 4
	// bytes of the RIPEMD160(SHA256(parentPubKey)).
	parentFP := Hash160(k.pubKeyBytes())[:4]
	return NewHDKey(k.version, childKey, childChainCode, parentFP,
		k.depth+1, i, isPrivate), nil
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

func (k *HDKey) DriverPath(path string) (*HDKey, error) {

	if path == "m" || path == "M" || path == "m'" || path == "M'" {
		return k, nil
	}

	driverpath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}

	ck := NewHDKey(k.version, k.key, k.chainCode, k.parentFP, k.depth, k.childNum, k.isPrivate)

	for _, n := range driverpath {
		ck, err = ck.Child(n)
		if err != nil {
			return nil, err
		}
	}
	return ck, nil

}

func (k *HDKey) DriverChild(index uint32) (*HDKey, error) {
	ck, err := k.Child(index)
	if err != nil {
		return nil, err
	}
	return ck, nil

}

func (k *HDKey) PrivateHDkey() string {

	if !k.isPrivate {
		return "This is a public key only"
	}
	if len(k.key) == 0 {
		return "zeroed private hd key"
	}

	var childNumBytes [4]byte
	binary.BigEndian.PutUint32(childNumBytes[:], k.childNum)

	// The serialized format is:
	//   version (4) || depth (1) || parent fingerprint (4)) ||
	//   child num (4) || chain code (32) || key data (33) || checksum (4)
	serializedBytes := make([]byte, 0, serializedKeyLen+4)
	serializedBytes = append(serializedBytes, k.version...)
	serializedBytes = append(serializedBytes, k.depth)
	serializedBytes = append(serializedBytes, k.parentFP...)
	serializedBytes = append(serializedBytes, childNumBytes[:]...)
	serializedBytes = append(serializedBytes, k.chainCode...)

	serializedBytes = append(serializedBytes, 0x00)
	serializedBytes = paddedAppend(32, serializedBytes, k.key)

	checkSum := DoubleHashB(serializedBytes)[:4]
	serializedBytes = append(serializedBytes, checkSum...)
	return base58.Encode(serializedBytes)
}

func (k *HDKey) PublicHDkey() string {
	if len(k.pubKeyBytes()) == 0 {
		return "zeroed public HD key"
	}

	var childNumBytes [4]byte
	binary.BigEndian.PutUint32(childNumBytes[:], k.childNum)

	// The serialized format is:
	//   version (4) || depth (1) || parent fingerprint (4)) ||
	//   child num (4) || chain code (32) || key data (33) || checksum (4)
	serializedBytes := make([]byte, 0, serializedKeyLen+4)
	serializedBytes = append(serializedBytes, HDPublicKeyID[:]...)
	serializedBytes = append(serializedBytes, k.depth)
	serializedBytes = append(serializedBytes, k.parentFP...)
	serializedBytes = append(serializedBytes, childNumBytes[:]...)
	serializedBytes = append(serializedBytes, k.chainCode...)

	serializedBytes = append(serializedBytes, k.pubKeyBytes()...)

	checkSum := DoubleHashB(serializedBytes)[:4]
	serializedBytes = append(serializedBytes, checkSum...)
	return base58.Encode(serializedBytes)
}

// zero sets all bytes in the passed slice to zero.  This is used to
// explicitly clear private key material from memory.
func zero(b []byte) {
	lenb := len(b)
	for i := 0; i < lenb; i++ {
		b[i] = 0
	}
}

// Zero manually clears all fields and bytes in the extended key.  This can be
// used to explicitly clear key material from memory for enhanced security
// against memory scraping.  This function only clears this particular key and
// not any children that have already been derived.
func (k *HDKey) Zero() {
	zero(k.key)
	zero(k.pubKey)
	zero(k.chainCode)
	zero(k.parentFP)
	k.version = nil
	k.key = nil
	k.depth = 0
	k.childNum = 0
	k.isPrivate = false
}

// NewMaster creates a new master node for use in creating a hierarchical
// deterministic key chain.  The seed must be between 128 and 512 bits and
// should be generated by a cryptographically secure random generation source.
//
// NOTE: There is an extremely small chance (< 1 in 2^127) the provided seed
// will derive to an unusable secret key.  The ErrUnusable error will be
// returned if this should occur, so the caller must check for it and generate a
// new seed accordingly.
func NewMaster(seed []byte) (*HDKey, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if len(seed) < MinSeedBytes || len(seed) > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	// First take the HMAC-SHA512 of the master key and the seed data:
	//   I = HMAC-SHA512(Key = "Bitcoin seed", Data = S)
	hmac512 := hmac.New(sha512.New, masterKey)
	hmac512.Write(seed)
	lr := hmac512.Sum(nil)

	// Split "I" into two 32-byte sequences Il and Ir where:
	//   Il = master secret key
	//   Ir = master chain code
	secretKey := lr[:len(lr)/2]
	chainCode := lr[len(lr)/2:]

	// Ensure the key in usable.
	secretKeyNum := new(big.Int).SetBytes(secretKey)
	if secretKeyNum.Cmp(btcec.S256().N) >= 0 || secretKeyNum.Sign() == 0 {
		return nil, ErrUnusableSeed
	}

	parentFP := []byte{0x00, 0x00, 0x00, 0x00}
	return NewHDKey(HDPrivateKeyID[:], secretKey, chainCode,
		parentFP, 0, 0, true), nil
}

func NewFromMnemonic(mnemonic string) (*HDKey, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}

	seed := bip39.NewSeed(mnemonic, "")
	return NewMaster(seed)

}

func NewFromMnemonicWithPassword(mnemonic, password string) (*HDKey, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}

	seed := bip39.NewSeed(mnemonic, password)
	return NewMaster(seed)
}

func CreateBip39HDKey() (string, *HDKey, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", nil, err
	}

	seed := bip39.NewSeed(mnemonic, "")
	hdkey, err := NewMaster(seed)
	if err != nil {
		return "", nil, err
	}
	return mnemonic, hdkey, nil
}

func CreateBip39HDKeyWithPassword(password string) (string, *HDKey, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", nil, err
	}

	seed := bip39.NewSeed(mnemonic, password)
	hdkey, err := NewMaster(seed)
	if err != nil {
		return "", nil, err
	}
	return mnemonic, hdkey, nil
}

func (k *HDKey) DriverPathKey(path string) (*Key, error) {
	ck, err := k.DriverPath(path)
	if err != nil {
		return nil, err
	}
	if ck.isPrivate {
		privateKeyECDSA, err := crypto.ToECDSA(ck.key)
		if err != nil {
			return nil, err
		}
		id := uuid.NewRandom()
		var seed keys.Uint256
		copy(seed[:], ck.key)
		pubBytes := keys.Seed2Addr(&seed)
		tkBytes := keys.Seed2Tk(&seed)
		key := &Key{
			Id:         id,
			Address:    address.BytesToAccount(pubBytes[:]),
			Tk:         address.BytesToAccount(tkBytes[:]),
			PrivateKey: privateKeyECDSA,
		}
		return key, nil
	} else {
		return nil, errors.New("there is no privateKey")
	}

}

func (k *HDKey) DriverChildKey(index uint32) (*Key, error) {
	ck, err := k.DriverChild(index)
	if err != nil {
		return nil, err
	}
	if ck.isPrivate {
		privateKey := ck.key
		privateKeyECDSA, err := crypto.ToECDSA(privateKey)
		if err != nil {
			return nil, err
		}
		id := uuid.NewRandom()
		key := &Key{
			Id:         id,
			Address:    crypto.PrivkeyToAddress(privateKeyECDSA),
			Tk:         crypto.PrivkeyToTk(privateKeyECDSA),
			PrivateKey: privateKeyECDSA,
		}
		return key, nil
	} else {
		return nil, errors.New("there is no privateKey")
	}

}

// NewKeyFromString returns a new extended key instance from a base58-encoded
// extended key.
func NewKeyFromString(key string) (*HDKey, error) {
	// The base58-decoded extended key must consist of a serialized payload
	// plus an additional 4 bytes for the checksum.
	decoded := base58.Decode(key)
	if len(decoded) != 82 {
		return nil, ErrInvalidKeyLen
	}
	if len(decoded) != serializedKeyLen+4 {
		return nil, ErrInvalidKeyLen
	}

	// The serialized format is:
	//   version (4) || depth (1) || parent fingerprint (4)) ||
	//   child num (4) || chain code (32) || key data (33) || checksum (4)

	// Split the payload and checksum up and ensure the checksum matches.
	payload := decoded[:len(decoded)-4]
	checkSum := decoded[len(decoded)-4:]
	expectedCheckSum := DoubleHashB(payload)[:4]
	if !bytes.Equal(checkSum, expectedCheckSum) {
		return nil, ErrBadChecksum
	}

	// Deserialize each of the payload fields.
	version := payload[:4]
	depth := payload[4:5][0]
	parentFP := payload[5:9]
	childNum := binary.BigEndian.Uint32(payload[9:13])
	chainCode := payload[13:45]
	keyData := payload[45:78]

	// The key data is a private key if it starts with 0x00.  Serialized
	// compressed pubkeys either start with 0x02 or 0x03.
	isPrivate := keyData[0] == 0x00
	if isPrivate {
		// Ensure the private key is valid.  It must be within the range
		// of the order of the secp256k1 curve and not be 0.
		keyData = keyData[1:]
		keyNum := new(big.Int).SetBytes(keyData)
		if keyNum.Cmp(btcec.S256().N) >= 0 || keyNum.Sign() == 0 {
			return nil, ErrUnusableSeed
		}
	} else {
		// Ensure the public key parses correctly and is actually on the
		// secp256k1 curve.
		_, err := btcec.ParsePubKey(keyData, btcec.S256())
		if err != nil {
			return nil, err
		}
	}

	return NewHDKey(version, keyData, chainCode, parentFP, depth,
		childNum, isPrivate), nil
}

// GenerateSeed returns a cryptographically secure random seed that can be used
// as the input for the NewMaster function to generate a new master node.
//
// The length is in bytes and it must be between 16 and 64 (128 to 512 bits).
// The recommended length is 32 (256 bits) as defined by the RecommendedSeedLen
// constant.
func GenerateSeed(length uint8) ([]byte, error) {
	// Per [BIP32], the seed must be in range [MinSeedBytes, MaxSeedBytes].
	if length < MinSeedBytes || length > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func DoubleHashB(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}
