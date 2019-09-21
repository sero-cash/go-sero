package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txtool/flight"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-sero/zero/txtool"
)

func Dec(tk_str string, out_str string) {
	stdin := bufio.NewReader(os.Stdin)
	if len(tk_str) == 0 {
		fmt.Println("input tk:")
		var err error
		tk_str, err = stdin.ReadString('\n')
		if err != nil {
			OUTPUT_ERROR("TK READ ERROR", nil)
			return
		}
		tk_str = strings.Trim(tk_str, "\n")
		fmt.Println(tk_str)
	}
	if len(out_str) == 0 {
		fmt.Println("input out:")
		var err error
		out_str, err = stdin.ReadString('\n')
		if err != nil {
			OUTPUT_ERROR("OUT READ ERROR", nil)
			return
		}
		out_str = strings.Trim(out_str, "\n")
		fmt.Println(out_str)
	}

	tk_str = strings.Trim(tk_str, "'")
	out_str = strings.Trim(out_str, "'")

	if tk_str[1] != 'x' {
		tk_str = "0x" + tk_str
	}
	if tk_bs, e := hexutil.Decode(tk_str); e == nil {
		if len(tk_bs) == 64 {
			tk := c_type.Tk{}
			copy(tk[:], tk_bs)
			var outs []txtool.Out
			if e := json.Unmarshal([]byte(out_str), &outs); e == nil {
				douts := flight.DecOut(&tk, outs)
				if douts_bs, e := json.Marshal(douts); e == nil {
					OUTPUT_RESULT(string(douts_bs))
				} else {
					OUTPUT_ERROR("Marshal-", e)
				}
			} else {
				OUTPUT_ERROR("Unmarshal-", e)
			}
		} else {
			OUTPUT_ERROR("tk must 64 bytes", nil)
		}
	} else {
		OUTPUT_ERROR("TKDecode-", e)
	}
}
