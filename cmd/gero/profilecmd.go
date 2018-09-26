package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/trace"

	"github.com/sero-cash/go-sero/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var profileCommand = cli.Command{
	Action:    utils.MigrateFlags(startPprof),
	Name:      "profile",
	Usage:     "turn on performance analysis tools",
	ArgsUsage: "<profilePort> ",
	Category:  "MISCELLANEOUS COMMANDS",
}

//运行pprof分析器
func startPprof(ctx *cli.Context) error {
	node := makeFullNode(ctx)
	startNode(ctx, node)
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf(`Usage: gero profile <profilePort>`)
	}
	HTTPHost := "127.0.0.1"
	if ctx.GlobalIsSet(utils.RPCListenAddrFlag.Name) {
		HTTPHost = ctx.GlobalString(utils.RPCListenAddrFlag.Name)
	}
	go func(port string) {
		http.HandleFunc("/start", traces)
		http.HandleFunc("/stop", traceStop)
		http.HandleFunc("/trace", StaticServer)
		http.ListenAndServe(HTTPHost+":"+port, nil)
	}(args[0])
	node.Wait()
	return nil
}

func traces(w http.ResponseWriter, r *http.Request) {

	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}

	err = trace.Start(f)
	if err != nil {
		w.Write([]byte("trace enable or shutdown"))
	}
	w.Write([]byte("TrancStart"))
}

func traceStop(w http.ResponseWriter, r *http.Request) {
	trace.Stop()
	w.Write([]byte("TrancStop"))
}

func StaticServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=trace.out")
	path := "trace.out"
	http.ServeFile(w, req, path)
}
