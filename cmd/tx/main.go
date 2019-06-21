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
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-czero-import/cpt"
)

var txParam = ""
var sk = ""

func init() {
	flag.StringVar(&txParam, "tx", "", "param for gen tx")
	flag.StringVar(&sk, "sk", "", "sk for signing tx")
}

func main() {
	seroparam.InitExchangeValueStr(true)
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("PThread: %v \n", zconfig.G_p_thread_num)
	flag.Parse()
	cpt.ZeroInit_OnlyInOuts()

	stdin := bufio.NewReader(os.Stdin)
	if len(sk) == 0 {
		fmt.Println("input sk:")
		var err error
		sk, err = stdin.ReadString('\n')
		if err != nil {
			fmt.Println("[OUTPUT-BEGIN] ERROR: SK READ ERROR")
		}
		sk = strings.Trim(sk, "\n")
		fmt.Println(sk)
	}

	if len(txParam) == 0 {
		fmt.Println("input txParam:")
		var err error
		txParam, err = stdin.ReadString('\n')
		if err != nil {
			fmt.Println("[OUTPUT-BEGIN] ERROR: TXPARAM READ ERROR - ", err)
		}
		txParam = strings.Trim(txParam, "\n")
		fmt.Println(txParam)
	}

	sk = strings.Trim(sk, "'")
	txParam = strings.Trim(txParam, "'")

	sk_bytes := keys.Uint512{}
	if sk[1] != 'x' {
		sk = "0x" + sk
	}
	if bs, e := hexutil.Decode(sk); e != nil {
		fmt.Println("[OUTPUT-BEGIN] ERROR: DecodeSK-", e)
	} else {
		copy(sk_bytes[:], bs)
		if len(txParam) > 0 && len(sk) > 0 {
			var gtp txtool.GTxParam
			if e := json.Unmarshal([]byte(txParam), &gtp); e != nil {
				fmt.Println("[OUTPUT-BEGIN] ERROR: Unmarshal-", e)
			} else {
				if gtx, e := flight.SignTx(&sk_bytes, &gtp); e != nil {
					fmt.Println("[OUTPUT-BEGIN] ERROR: SignTx-", e)
				} else {
					if jtx, e := json.Marshal(&gtx); e != nil {
						fmt.Println("[OUTPUT-BEGIN] ERROR: Marshal-", e)
					} else {
						fmt.Println("[OUTPUT-BEGIN] " + string(jtx))
					}
				}
			}
		} else {
			fmt.Println("[OUTPUT-BEGIN] ERROR: Input params invalid")
		}
	}
}
