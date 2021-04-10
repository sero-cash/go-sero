// Copyright 2020 The sero Authors
// This file is part of the sero library.
//
// The sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the sero library. If not, see <http://www.gnu.org/licenses/>.

package sero

import (
	"encoding/json"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txtool"
)

var committedTxPrefix = []byte("COMMITEDTX")

type Backup struct {
	db *serodb.LDBDatabase
}

func NewBackup(dbpath string) *Backup {
	db, err := serodb.NewLDBDatabase(dbpath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	return &Backup{db: db}
}

func committedTxKey(txhash c_type.Uint256) []byte {
	return append(committedTxPrefix, txhash[:]...)
}

func (self *Backup) PutCommittedTx(args *txtool.GTx) error {
	blob, err := json.Marshal(args)
	if err != nil {
		return err
	}
	err = self.db.Put(committedTxKey(args.Hash), blob)
	if err != nil {
		return err
	}
	return nil
}

func (self *Backup) GetCommittedTx(txHash c_type.Uint256) (*txtool.GTx, error) {
	value, err := self.db.Get(committedTxKey(txHash))
	if err != nil {
		return nil, err
	}
	var gtx txtool.GTx
	err = json.Unmarshal(value, &gtx)
	if err != nil {
		return nil, err
	}
	return &gtx, nil
}
