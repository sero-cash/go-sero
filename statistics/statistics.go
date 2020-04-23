package statistics

import (
	"bytes"
	"fmt"
	"github.com/sero-cash/go-sero/log"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Method struct {
	Start    time.Time
	Interval int64
	Count    int64
}

var statistics sync.Map
var keys []string
var start int

func Start() {
	if start == 0 {
		start = 1
		// AddJob("0/20 * * * * ?", print)
	}
}
func Now() {
	if start == 0 {
		return
	}
	funName := runFuncName()
	if val, ok := statistics.Load(funName); ok {
		method := val.(*Method)
		method.Start = time.Now()
	} else {
		keys = append(keys, funName)
		statistics.Store(funName, &Method{Start: time.Now()})
	}

}
func Since() {
	if start == 0 {
		return
	}
	funName := runFuncName()
	if val, ok := statistics.Load(funName); ok {
		method := val.(*Method)
		method.Interval += time.Since(method.Start).Milliseconds()
		method.Count++
	} else {
		panic(funName + " error")
	}
	if funName == "core.(*BlockChain).insertChain" {
		print()
	}
}

func print() {
	var buffer bytes.Buffer

	for _, key := range keys {
		if val, ok := statistics.Load(key); ok {
			method := val.(*Method)
			buffer.WriteString(fmt.Sprintf("%s : %vms %v, ", key, method.Interval, method.Count))
		} else {
			panic(key + " error")
		}
	}
	log.Info("statistics", "methods", buffer.String())
}

func runFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])
	name := f.Name()
	last := strings.LastIndex(name, "/")
	if last == -1 {
		return name
	}
	return name[last+1:]
}
