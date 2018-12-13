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
	"errors"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/txs"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/lstate"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
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

	if _, ok := w.keystore.unlocked[w.account.Address]; ok {
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
	return account.Address == w.account.Address && (account.URL == (accounts.URL{}) || account.URL == w.account.URL)
}

// Derive implements accounts.Wallet, but is a noop for plain wallets since there
// is no notion of hierarchical account derivation for plain keystore accounts.
func (w *keystoreWallet) Derive(path accounts.DerivationPath, pin bool) (accounts.Account, error) {
	return accounts.Account{}, accounts.ErrNotSupported
}

// SelfDerive implements accounts.Wallet, but is a noop for plain wallets since
// there is no notion of hierarchical account derivation for plain keystore accounts.
func (w *keystoreWallet) SelfDerive(base accounts.DerivationPath, chain sero.ChainStateReader) {}

func (w *keystoreWallet) EncryptTx(account accounts.Account, tx *types.Transaction, txt *tx.T, state *state.StateDB) (*types.Transaction, error) {
	// Make sure the requested account is contained within
	if account.Address != w.account.Address {
		return nil, accounts.ErrUnknownAccount
	}
	if account.URL != (accounts.URL{}) && account.URL != w.account.URL {
		return nil, accounts.ErrUnknownAccount
	}
	seed, err := w.keystore.GetSeed(account)
	if err != nil {
		return nil, err
	}
	return w.EncryptTxWithSeed(*seed, tx, txt, state)

}

func (w *keystoreWallet) EncryptTxWithSeed(seed common.Seed, btx *types.Transaction, txt *tx.T, state *state.StateDB) (*types.Transaction, error) {
	w.keystore.mu.Lock()
	defer w.keystore.mu.Unlock()
	ins := []tx.In{}
	costTkn := txt.TokenCost()
	costTkt := txt.TikectCost()
	tk := keys.Seed2Tk(seed.SeedToUint256())
	outs, tknMap, tktMap, err := txs.GetRoots(&tk, costTkn, costTkt)
	if err != nil {
		return nil, err
	}
	for _, out := range outs {
		ins = append(ins, tx.In{Root: out})
	}
	for cy, value := range tknMap {
		token := &assets.Token{
			Currency: cy,
			Value:    value,
		}
		asset := assets.Asset{
			Tkn: token,
		}
		selfOut := tx.Out{
			Addr:  keys.Seed2Addr(seed.SeedToUint256()),
			Asset: asset,
			Z:     tx.TYPE_Z,
		}
		txt.Outs = append(txt.Outs, selfOut)
	}
	for catg, value := range tktMap {
		for _, v := range value {
			ticket := &assets.Ticket{
				Category: catg,
				Value:    v,
			}
			asset := assets.Asset{
				Tkt: ticket,
			}
			selfOut := tx.Out{
				Addr:  keys.Seed2Addr(seed.SeedToUint256()),
				Asset: asset,
				Z:     tx.TYPE_Z,
			}
			txt.Outs = append(txt.Outs, selfOut)
		}

	}

	if txt.PkgClose != nil {
		zpkg := lstate.CurrentState1().State.Pkgs.GetPkg(&txt.PkgClose.Id)
		if zpkg == nil {
			return nil, errors.New("PkgClose Id is not exists!")
		}
		pkg_o, err := pkg.DePkg(&txt.PkgClose.Key, &zpkg.Pack.Pkg)
		if err != nil {
			return nil, err
		}
		selfOut := tx.Out{
			Addr:  zpkg.Pack.Pkr,
			Asset: pkg_o.Asset,
			Z:     tx.TYPE_Z,
		}
		txt.Outs = append(txt.Outs, selfOut)
	}

	txt.Ins = ins

	log.Info("EncryptTxWithSeed : ", "in_num", len(txt.Ins), "out_num", len(txt.Outs))

	for i, in := range txt.Ins {
		log.Info("    ctx_in : ", "index", i, "root", in.Root)
	}
	for i, out := range txt.Outs {
		log.Info("    ctx_out : ", "index", i, "to", hexutil.Encode(out.Addr[:]))
	}

	stx, err := txs.Gen(seed.SeedToUint256(), txt)
	if err != nil {
		return nil, err
	}

	for i, in := range stx.Desc_Z.Ins {
		log.Info("    desc_z : ", "index", i, "nil", hexutil.Encode(in.Nil[:]), "trace", hexutil.Encode(in.Trace[:]))
	}

	return btx.WithEncrypt(&stx)

}

func (w *keystoreWallet) EncryptTxWithPassphrase(account accounts.Account, passphrase string, tx *types.Transaction, txt *tx.T, state *state.StateDB) (*types.Transaction, error) {
	// Make sure the requested account is contained within
	if account.Address != w.account.Address {
		return nil, accounts.ErrUnknownAccount
	}
	if account.URL != (accounts.URL{}) && account.URL != w.account.URL {
		return nil, accounts.ErrUnknownAccount
	}

	seed, err := w.keystore.GetSeedWithPassphrase(account, passphrase)
	if err != nil {
		return nil, err
	}
	return w.EncryptTxWithSeed(*seed, tx, txt, state)

}

/*func (w *keystoreWallet) GetSeed() (*common.Seed, error) {
	// Make sure the requested account is contained within
	seed ,err:=w.keystore.GetSeed(w.account)
	if err != nil{
		return nil, err
	}
	return seed,nil

}*/

func (w *keystoreWallet) IsMine(onceAddress common.Address) bool {
	tk := w.account.Tk.ToUint512()
	succ, _ := keys.IsMyPKr(tk, onceAddress.ToUint512())
	return succ
}
