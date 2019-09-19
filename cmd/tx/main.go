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
	"flag"
	"fmt"
	"runtime"

	"github.com/sero-cash/go-czero-import/c_czero"

	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/seroparam"
)

var method = ""
var txParam = ""
var sk = ""
var tk = ""
var out = ""
var key = ""

func init() {
	flag.StringVar(&method, "method", "", "tx method")
	flag.StringVar(&txParam, "tx", "", "txparam for sign")
	flag.StringVar(&sk, "sk", "", "sk for sign")
	flag.StringVar(&tk, "tk", "", "tk for dec")
	flag.StringVar(&out, "out", "", "out for dec")
}

func OUTPUT_RESULT(result interface{}) {
	fmt.Println("[OUTPUT-BEGIN]", result)
}

func OUTPUT_ERROR(result interface{}, e error) {
	fmt.Println("[OUTPUT-BEGIN] ERROR:", result, e)
}

func main() {
	fmt.Println("hello", "ok")
	seroparam.InitExchangeValueStr(true)
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("PThread: %v \n", zconfig.G_p_thread_num)
	flag.Parse()

	if method == "sign" {
		c_czero.ZeroInit_OnlyInOuts()
		Sign(sk, txParam)
		return
	}
	if method == "dec" {
		c_czero.ZeroInit_NoCircuit()
		Dec(tk, out)
		return
	}
	if method == "confirm" {
		c_czero.ZeroInit_NoCircuit()
		Confirm(key, out)
		return
	}
	OUTPUT_ERROR("METHOD-MUST-[sign,dec,confirm]", nil)
}
