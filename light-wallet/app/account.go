package app

import (
	"os"
	"crypto/ecdsa"
	"github.com/sero-cash/go-sero/crypto"
	"path/filepath"
	"io/ioutil"
	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/pborman/uuid"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	"github.com/sero-cash/go-czero-import/keys"
	"math/big"
	"github.com/sero-cash/go-sero/accounts"
)

type Account struct {
	wallet       accounts.Wallet
	pk           *keys.Uint512
	tk           *keys.Uint512
	skr          keys.PKr
	mainPkr      keys.PKr
	balances     map[string]*big.Int
	utxoNums     map[string]uint64

	isChanged bool
	keyPath   string
}

func (account *Account) Create(passphrase string) error {

	var privateKey *ecdsa.PrivateKey
	// If not loaded, generate random.
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		logex.Fatalf("Failed to generate random private key: %v", err)
		return err
	}
	// Create the keyfile object with a random UUID.
	id := uuid.NewRandom()
	address := crypto.PrivkeyToAddress(privateKey)
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PrivkeyToAddress(privateKey),
		Tk:         crypto.PrivkeyToTk(privateKey),
		PrivateKey: privateKey,
	}
	// Encrypt key with passphrase.
	keyjson, err := keystore.EncryptKey(key, passphrase, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		logex.Fatalf("Error encrypting key: %v", err)
		return err
	}
	// Store the file to disk.
	if err := os.MkdirAll(filepath.Dir(GetKeystorePath()+"/"+address.String()), 0700); err != nil {
		logex.Fatalf("Could not create directory %s", filepath.Dir(GetKeystorePath()))
		return err
	}
	if err := ioutil.WriteFile(GetKeystorePath()+"/"+address.String(), keyjson, 0600); err != nil {
		logex.Fatalf("Failed to write keyfile to %s: %v", GetKeystorePath(), err)
		return err
	}
	// Output some information.
	logex.Infof("Create account successful. address =[%s]", key.Address)
	return nil

}

func (account *Account) Import(passphrase, keyPath string) error {
	keyJson, err := ioutil.ReadFile(keyPath)
	if err != nil {
		logex.Errorf("Failed to read the keyfile at '%s': %v", keyPath, err)
		return err
	}
	// Decrypt key with passphrase.
	key, err := keystore.DecryptKey(keyJson, passphrase)
	if err != nil {
		logex.Errorf("Error decrypting key: %v", err)
		return err
	}
	// Then write the new keyfile in place of the old one.
	if err := ioutil.WriteFile(GetKeystorePath()+"/"+key.Address.String(), keyJson, 0600); err != nil {
		logex.Errorf("Error writing new keyFile to disk: %v", err)
		return err
	}
	logex.Infof("Import account successful. address=[%s]", key.Address)
	return nil
}

func (account *Account) UpdatePass(oldPas, newPass string) error {
	keyJson, err := ioutil.ReadFile(account.keyPath)
	if err != nil {
		logex.Errorf("Failed to read the keyfile at '%s': %v", account.keyPath, err)
		return err
	}
	// Decrypt key with passphrase.
	key, err := keystore.DecryptKey(keyJson, oldPas)
	if err != nil {
		logex.Errorf("Error decrypting key: %v", err)
		return err
	}

	// Encrypt the key with the new passphrase.
	newJson, err := keystore.EncryptKey(key, newPass, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		logex.Errorf("Error encrypting with new passphrase: %v", err)
	}

	// Then write the new keyfile in place of the old one.
	if err := ioutil.WriteFile(GetKeystorePath()+"/"+key.Address.String(), newJson, 0600); err != nil {
		logex.Errorf("Error writing new keyFile to disk: %v", err)
		return err
	}
	logex.Infof("Update account pass successful. address=[%s]", key.Address)
	return nil
}

func (account *Account) Export() {

}
