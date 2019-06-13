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
	"encoding/json"
	"flag"
	"fmt"
	"runtime"

	"github.com/sero-cash/go-sero/zero/txs/generate"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/zero/light"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/light/light_types"
)

var txParam = ""
var sk = ""

func init() {
	flag.StringVar(&txParam, "tx", "", "param for gen tx")
	flag.StringVar(&sk, "sk", "", "sk for signing tx")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("PThread: %v \n", generate.G_p_thread_num)
	flag.Parse()
	cpt.ZeroInit_OnlyInOuts()

	fmt.Println("[OUTPUT-BEGIN]")
	sk_bytes := keys.Uint512{}
	if sk[1] != 'x' {
		sk = "0x" + sk
	}
	if bs, e := hexutil.Decode(sk); e != nil {
		fmt.Println("ERROR: DecodeSK-", e)
	} else {
		copy(sk_bytes[:], bs)
		if len(txParam) > 0 && len(sk) > 0 {
			var gtp light_types.GenTxParam
			if e := json.Unmarshal([]byte(txParam), &gtp); e != nil {
				fmt.Println("ERROR: Unmarshal-", e)
			} else {
				copy(gtp.From.SKr[:], sk_bytes[:])
				for i := range gtp.Ins {
					copy(gtp.Ins[i].SKr[:], sk_bytes[:])
				}
				if gtx, e := light.SLI_Inst.GenTx(&gtp); e != nil {
					fmt.Println("ERROR: GenTx-", e)
				} else {
					if jtx, e := json.Marshal(&gtx); e != nil {
						fmt.Println("ERROR: Marshal-", e)
					} else {
						fmt.Println(string(jtx))
					}
				}
			}
		} else {
			fmt.Println("ERROR: Input params invalid")
		}
	}
	fmt.Println("[OUTPUT-END]")
}
