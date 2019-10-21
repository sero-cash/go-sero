package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_1"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txtool/generate/generate_0"

	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/common/hexutil"
)

func Confirm(key_str string, out_str string) {
	stdin := bufio.NewReader(os.Stdin)
	if len(key_str) == 0 {
		fmt.Println("input key:")
		var err error
		key_str, err = stdin.ReadString('\n')
		if err != nil {
			OUTPUT_ERROR("TK READ ERROR", nil)
			return
		}
		key_str = strings.Trim(key_str, "\n")
		fmt.Println(key_str)
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

	key_str = strings.Trim(key_str, "'")
	out_str = strings.Trim(out_str, "'")

	if key_str[1] != 'x' {
		key_str = "0x" + key_str
	}
	if key_bs, e := hexutil.Decode(key_str); e == nil {
		if len(key_bs) == 32 {
			key := c_type.Uint256{}
			copy(key[:], key_bs)
			if strings.Contains(out_str, "\"OutCM\":") {
				var out stx_v0.Out_Z
				if e := json.Unmarshal([]byte(out_str), &out); e == nil {
					if dout := generate_0.ConfirmOutZ(&key, true, &out); dout != nil {
						if dout_bs, e := json.Marshal(dout); e == nil {
							OUTPUT_RESULT(string(dout_bs))
						} else {
							OUTPUT_ERROR("Marshal-", e)
						}
					} else {
						OUTPUT_ERROR("Confirm OutZ Failed", nil)
					}
				} else {
					OUTPUT_ERROR("Out_Z Unmarshal-", e)
				}
			} else {
				var out stx_v1.Out_C
				if e := json.Unmarshal([]byte(out_str), &out); e == nil {
					if dout, _ := generate_1.ConfirmOutC(&key, &out); dout != nil {
						if dout_bs, e := json.Marshal(dout); e == nil {
							OUTPUT_RESULT(string(dout_bs))
						} else {
							OUTPUT_ERROR("Out_C Marshal-", e)
						}
					}
				} else {
					OUTPUT_ERROR("Out_C Unmarshal-", e)
				}
			}
		} else {
			OUTPUT_ERROR("key must 32 bytes", nil)
		}
	} else {
		OUTPUT_ERROR("KeyDecode-", e)
	}
}
