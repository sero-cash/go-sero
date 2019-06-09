package accounts

import (
	"github.com/sero-cash/go-sero/common/address"

	"github.com/tyler-smith/go-bip39"
)

func NewMnemonic() (string, error) {

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

func NewSeed(mnemonic, password string) (*address.Seed, error) {
	s, err := bip39.NewSeedWithErrorChecking(mnemonic, password)
	if err != nil {
		return nil, err
	}
	seed := &address.Seed{}
	copy(seed[:], s[:])
	return seed, nil
}

func PublicExtendPrikey(mnemonic, password string) []byte {
	return bip39.NewSeed(mnemonic, password)
}
