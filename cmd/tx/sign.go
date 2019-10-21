package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
)

func Sign(sk string, txParam string) {
	stdin := bufio.NewReader(os.Stdin)
	if len(sk) == 0 {
		fmt.Println("input sk:")
		var err error
		sk, err = stdin.ReadString('\n')
		if err != nil {
			OUTPUT_ERROR("SK READ ERROR", nil)
			return
		}
		sk = strings.Trim(sk, "\n")
		fmt.Println(sk)
	}

	if len(txParam) == 0 {
		fmt.Println("input txParam:")
		var err error
		txParam, err = stdin.ReadString('\n')
		if err != nil {
			OUTPUT_ERROR("TXPARAM READ ERROR", nil)
			return
		}
		txParam = strings.Trim(txParam, "\n")
		fmt.Println(txParam)
	}

	sk = strings.Trim(sk, "'")
	txParam = strings.Trim(txParam, "'")

	sk_bytes := c_type.Uint512{}
	if sk[1] != 'x' {
		sk = "0x" + sk
	}
	if bs, e := hexutil.Decode(sk); e != nil {
		OUTPUT_ERROR("DecodeSK-", e)
	} else {
		copy(sk_bytes[:], bs)
		if len(txParam) > 0 && len(sk) > 0 {
			var gtp txtool.GTxParam
			if e := json.Unmarshal([]byte(txParam), &gtp); e != nil {
				OUTPUT_ERROR("Unmarshal-", e)
			} else {
				if gtx, e := flight.SignTx(&sk_bytes, &gtp); e != nil {
					OUTPUT_ERROR("SignTx-", e)
				} else {
					if jtx, e := json.Marshal(&gtx); e != nil {
						OUTPUT_ERROR("Marshal-", e)
					} else {
						OUTPUT_RESULT(string(jtx))
					}
				}
			}
		} else {
			OUTPUT_ERROR("Input params invalid", nil)
		}
	}
}
