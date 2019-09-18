// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/sero-cash/go-sero/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

type outputRegister struct {
	PKr string
}

var commandRegister = cli.Command{
	Name:        "register",
	Usage:       "generate PKr",
	ArgsUsage:   "<keyfile>",
	Description: `Generate PKr use the keyfile.`,
	Flags: []cli.Flag{
		passphraseFlag,
		jsonFlag,
	},
	Action: func(ctx *cli.Context) error {
		keyfilepath := ctx.Args().First()

		// Read key from file.
		keyjson, err := ioutil.ReadFile(keyfilepath)
		if err != nil {
			utils.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
		}

		// Get Address from keyjson
		kstr, err := keystore.GetAddress(keyjson)
		if err != nil {
			utils.Fatalf("Failed to get address '%v'", err)
		}

		address := c_type.Uint512{}

		if err := c_czero.Base58Decode(&kstr, address[:]); err != nil {
			utils.Fatalf("Failed to decode address '%v'", err)
		}

		// Output all relevant information we can retrieve.
		r := c_type.RandUint256()
		pkr := c_czero.Addr2PKr(&address, &r)
		out := outputRegister{
			*cpt.Base58Encode(pkr[:]),
		}

		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(out)
		} else {
			fmt.Println("PKr:       ", out.PKr)
		}
		return nil
	},
}
