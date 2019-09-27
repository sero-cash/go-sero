// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package keystore

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/txs/tx"
)

// keystoreWallet implements the accounts.Wallet interface for the original
// keystore.
type keystoreWallet struct {
	account  accounts.Account // Single account contained in this wallet
	keystore *KeyStore        // Keystore where the account originates from
}

// URL implements accounts.Wallet, returning the URL of the account within.
func (w *keystoreWallet) URL() accounts.URL {
	return w.account.URL
}

// Status implements accounts.Wallet, returning whether the account held by the
// keystore wallet is unlocked or not.
func (w *keystoreWallet) Status() (string, error) {
	w.keystore.mu.RLock()
	defer w.keystore.mu.RUnlock()

	if _, ok := w.keystore.unlocked[w.account.Key]; ok {
		return "Unlocked", nil
	}
	return "Locked", nil
}

// Open implements accounts.Wallet, but is a noop for plain wallets since there
// is no connection or decryption step necessary to access the list of accounts.
func (w *keystoreWallet) Open(passphrase string) error { return nil }

// Close implements accounts.Wallet, but is a noop for plain wallets since is no
// meaningful open operation.
func (w *keystoreWallet) Close() error { return nil }

// Accounts implements accounts.Wallet, returning an account list consisting of
// a single account that the plain kestore wallet contains.
func (w *keystoreWallet) Accounts() []accounts.Account {
	return []accounts.Account{w.account}
}

// Contains implements accounts.Wallet, returning whether a particular account is
// or is not wrapped by this wallet instance.
func (w *keystoreWallet) Contains(account accounts.Account) bool {
	return account.Key == w.account.Key && (account.URL == (accounts.URL{}) || account.URL == w.account.URL)
}

// Derive implements accounts.Wallet, but is a noop for plain wallets since there
// is no notion of hierarchical account derivation for plain keystore accounts.
func (w *keystoreWallet) Derive(path accounts.DerivationPath, pin bool) (accounts.Account, error) {
	return accounts.Account{}, accounts.ErrNotSupported
}

// SelfDerive implements accounts.Wallet, but is a noop for plain wallets since
// there is no notion of hierarchical account derivation for plain keystore accounts.
func (w *keystoreWallet) SelfDerive(base accounts.DerivationPath, chain sero.ChainStateReader) {}

func (w *keystoreWallet) AddressUnlocked(account accounts.Account) (bool, error) {
	if account.Key != w.account.Key {
		return false, accounts.ErrUnknownAccount
	}
	if account.URL != (accounts.URL{}) && account.URL != w.account.URL {
		return false, accounts.ErrUnknownAccount
	}
	_, err := w.keystore.GetSeed(account)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (w *keystoreWallet) EncryptTxWithSeed(seed address.Seed, btx *types.Transaction, txt *tx.T, state *state.StateDB) (tx *types.Transaction, e error) {
	return
}

func (w *keystoreWallet) GetSeed() (*address.Seed, error) {
	// Make sure the requested account is contained within
	seed, err := w.keystore.GetSeed(w.account)
	if err != nil {
		return nil, err
	}
	return seed, nil

}

func (w *keystoreWallet) GetSeedWithPassphrase(passphrase string) (*address.Seed, error) {
	// Make sure the requested account is contained within
	seed, err := w.keystore.GetSeedWithPassphrase(w.account, passphrase)
	if err != nil {
		return nil, err
	}
	return seed, nil

}

func (w *keystoreWallet) IsMine(pkr c_type.PKr) bool {
	tk := w.account.Tk.ToTk()
	succ := superzk.IsMyPKr(&tk, &pkr)
	return succ
}
