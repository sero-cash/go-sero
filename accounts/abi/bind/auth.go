// Copyright 2016 The go-ethereum Authors
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

package bind

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/sero-cash/go-sero/crypto"
)

// NewTransactor is a utility method to easily create a transaction signer from
// an encrypted json key stream and the associated passphrase.
func NewTransactor(keyin io.Reader, passphrase string, value *big.Int) (*TransactOpts, error) {
	superzk.ZeroInit_NoCircuit()
	json, err := ioutil.ReadAll(keyin)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(json, passphrase)
	if err != nil {
		return nil, err
	}
	fromPkr := getMainPkr(key)

	return NewKeyedTransactor(key, &fromPkr, value), nil
}

func encodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func getMainPkr(key *keystore.Key) c_type.PKr {

	salt := encodeNumber(1)
	//log.Info("GenIndexPKr", "salt", hexutil.Encode(salt))
	random := append(key.Tk[:], salt...)
	r := crypto.Keccak256Hash(random).HashToUint256()
	pk := key.Address.ToUint512()
	return superzk.Pk2PKr(&pk, r)
}

// NewKeyedTransactor is a utility method to easily create a transaction signer
// from a single private key.
func NewKeyedTransactor(key *keystore.Key, refundTo *c_type.PKr, value *big.Int) *TransactOpts {
	tk := crypto.PrivkeyToTk(key.PrivateKey, key.Version)
	return &TransactOpts{
		From:     tk.ToPk(),
		Value:    value,
		RefundTo: refundTo,
		Encrypter: func(txParam *txtool.GTxParam) (*txtool.GTx, error) {
			priKey := crypto.FromECDSA(key.PrivateKey)
			var seed c_type.Uint256
			copy(seed[:], priKey[:])
			sk := superzk.Seed2Sk(&seed, key.Version)
			gtx, err := flight.SignTx(&sk, txParam)
			if err != nil {
				return nil, err
			}
			return &gtx, nil
		},
	}
}
