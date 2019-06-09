package hdwallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"sync"

	sero "github.com/sero-cash/go-sero"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/txs/tx"

	"github.com/sero-cash/go-sero/crypto"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"

	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	bip39 "github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	mnemonic        string
	masterKey       *hdkeychain.ExtendedKey
	seed            []byte
	url             accounts.URL
	deriveNextPaths map[address.AccountAddress]accounts.DerivationPath
	accounts        []accounts.Account
	stateLock       sync.RWMutex
}

func newWallet(seed []byte) (*Wallet, error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		masterKey:       masterKey,
		seed:            seed,
		accounts:        []accounts.Account{},
		deriveNextPaths: map[address.AccountAddress]accounts.DerivationPath{},
	}, nil
}

// NewFromMnemonic returns a new wallet from a BIP-39 mnemonic.
func NewFromMnemonic(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}

	seed, err := NewSeedFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	wallet, err := newWallet(seed)
	if err != nil {
		return nil, err
	}
	wallet.mnemonic = mnemonic

	return wallet, nil
}

// NewFromSeed returns a new wallet from a BIP-39 seed.
func NewFromSeed(seed []byte) (*Wallet, error) {
	if len(seed) == 0 {
		return nil, errors.New("seed is required")
	}

	return newWallet(seed)
}

// URL implements accounts.Wallet, returning the URL of the device that
// the wallet is on, however this does nothing since this is not a hardware device.
func (w *Wallet) URL() accounts.URL {
	return w.url
}

// Status implements accounts.Wallet, returning a custom status message
// from the underlying vendor-specific hardware wallet implementation,
// however this does nothing since this is not a hardware device.
func (w *Wallet) Status() (string, error) {
	return "ok", nil
}

// Open implements accounts.Wallet, however this does nothing since this
// is not a hardware device.
func (w *Wallet) Open(passphrase string) error {
	return nil
}

// Close implements accounts.Wallet, however this does nothing since this
// is not a hardware device.
func (w *Wallet) Close() error {
	return nil
}

// Accounts implements accounts.Wallet, returning the list of accounts pinned to
// the wallet. If self-derivation was enabled, the account list is
// periodically expanded based on current chain state.
func (w *Wallet) Accounts() []accounts.Account {
	// Attempt self-derivation if it's running
	// Return whatever account list we ended up with
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	cpy := make([]accounts.Account, len(w.accounts))
	copy(cpy, w.accounts)
	return cpy
}

// Contains implements accounts.Wallet, returning whether a particular account is
// or is not pinned into this wallet instance.
func (w *Wallet) Contains(account accounts.Account) bool {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	_, exists := w.deriveNextPaths[account.Address]
	return exists
}

// Unpin unpins account from list of pinned accounts.
func (w *Wallet) Unpin(account accounts.Account) error {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	for i, acct := range w.accounts {
		if acct.Address.String() == account.Address.String() {
			w.accounts = removeAtIndex(w.accounts, i)
			delete(w.deriveNextPaths, account.Address)
			return nil
		}
	}

	return errors.New("account not found")
}

// Derive implements accounts.Wallet, deriving a new account at the specific
// derivation path. If pin is set to true, the account will be added to the list
// of tracked accounts.
func (w *Wallet) Derive(path accounts.DerivationPath, pin bool) (accounts.Account, error) {
	// Try to derive the actual account and update its URL if successful
	w.stateLock.RLock() // Avoid device disappearing during derivation

	a, err := w.deriveAccount(path)

	w.stateLock.RUnlock()

	// If an error occurred or no pinning was requested, return
	if err != nil {
		return accounts.Account{}, err
	}

	account := accounts.Account{
		Address: a.Address,
		Tk:      a.Tk,
		URL: accounts.URL{
			Scheme: "",
			Path:   path.String(),
		},
	}

	if !pin {
		return account, nil
	}

	// Pinning needs to modify the state
	w.stateLock.Lock()
	defer w.stateLock.Unlock()

	if _, ok := w.deriveNextPaths[a.Address]; !ok {
		w.accounts = append(w.accounts, account)
		w.deriveNextPaths[a.Address] = path
	}

	return account, nil
}

// SelfDerive implements accounts.Wallet, trying to discover accounts that the
// user used previously (based on the chain state), but ones that he/she did not
// explicitly pin to the wallet manually. To avoid chain head monitoring, self
// derivation only runs during account listing (and even then throttled).
func (w *Wallet) SelfDerive(base accounts.DerivationPath, chain sero.ChainStateReader) {

}

func (w *Wallet) EncryptTx(account accounts.Account, tx *types.Transaction, txt *tx.T, state *state.StateDB) (*types.Transaction, error) {
	return nil, nil

}

func (w *Wallet) EncryptTxWithPassphrase(account accounts.Account, passphrase string, tx *types.Transaction, txt *tx.T, state *state.StateDB) (*types.Transaction, error) {
	return nil, nil
}

// IsMine return whether an once address is mine or not
func (w *Wallet) IsMine(onceAddress common.Address) bool {
	return true
}

func (w *Wallet) AddressUnlocked(account accounts.Account) (bool, error) {
	return false, nil
}

// PrivateKey returns the ECDSA private key of the account.
func (w *Wallet) PrivateKey(account accounts.Account) (*ecdsa.PrivateKey, error) {
	path, err := ParseDerivationPath(account.URL.Path)
	if err != nil {
		return nil, err
	}

	return w.derivePrivateKey(path)
}

// PrivateKeyBytes returns the ECDSA private key in bytes format of the account.
func (w *Wallet) PrivateKeyBytes(account accounts.Account) ([]byte, error) {
	privateKey, err := w.PrivateKey(account)
	if err != nil {
		return nil, err
	}

	return crypto.FromECDSA(privateKey), nil
}

// PrivateKeyHex return the ECDSA private key in hex string format of the account.
func (w *Wallet) PrivateKeyHex(account accounts.Account) (string, error) {
	privateKeyBytes, err := w.PrivateKeyBytes(account)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(privateKeyBytes)[2:], nil
}

// Address returns the address of the account.
func (w *Wallet) Address(account accounts.Account) (address.AccountAddress, error) {
	privateKey, err := w.PrivateKey(account)
	if err != nil {
		return address.AccountAddress{}, err
	}

	return crypto.PrivkeyToAddress(privateKey), nil
}

// AddressBytes returns the address in bytes format of the account.
func (w *Wallet) AddressBytes(account accounts.Account) ([]byte, error) {
	address, err := w.Address(account)
	if err != nil {
		return nil, err
	}
	return address.Bytes(), nil
}

// AddressHex returns the address in base58 string format of the account.
func (w *Wallet) AddressBase58(account accounts.Account) (string, error) {
	address, err := w.Address(account)
	if err != nil {
		return "", err
	}
	return address.String(), nil
}

// Path return the derivation path of the account.
func (w *Wallet) Path(account accounts.Account) (string, error) {
	return account.URL.Path, nil
}

// ParseDerivationPath parses the derivation path in string format into []uint32
func ParseDerivationPath(path string) (accounts.DerivationPath, error) {
	return accounts.ParseDerivationPath(path)
}

// MustParseDerivationPath parses the derivation path in string format into
// []uint32 but will panic if it can't parse it.
func MustParseDerivationPath(path string) accounts.DerivationPath {
	parsed, err := accounts.ParseDerivationPath(path)
	if err != nil {
		panic(err)
	}

	return parsed
}

// NewMnemonic returns a randomly generated BIP-39 mnemonic using 128-256 bits of entropy.
func NewMnemonic(bits int) (string, error) {
	entropy, err := bip39.NewEntropy(bits)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// NewSeed returns a randomly generated BIP-39 seed.
func NewSeed() ([]byte, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	return b, err
}

// NewSeedFromMnemonic returns a BIP-39 seed based on a BIP-39 mnemonic.
func NewSeedFromMnemonic(mnemonic string) ([]byte, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	return bip39.NewSeedWithErrorChecking(mnemonic, "")
}

// DerivePrivateKey derives the private key of the derivation path.
func (w *Wallet) derivePrivateKey(path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	var err error
	key := w.masterKey
	for _, n := range path {
		key, err = key.Child(n)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := key.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}

	return privateKeyECDSA, nil
}

// DeriveAddress derives the account address of the derivation path.
func (w *Wallet) deriveAccount(path accounts.DerivationPath) (accounts.Account, error) {
	privateKeyECDSA, err := w.derivePrivateKey(path)
	if err != nil {
		return accounts.Account{}, err
	}

	address := crypto.PrivkeyToAddress(privateKeyECDSA)
	tk := crypto.PrivkeyToTk(privateKeyECDSA)
	return accounts.Account{Address: address, Tk: tk}, nil
}

// removeAtIndex removes an account at index.
func removeAtIndex(accts []accounts.Account, index int) []accounts.Account {
	return append(accts[:index], accts[index+1:]...)
}
