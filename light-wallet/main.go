package main

import (
	"github.com/sero-cash/go-sero/light-wallet/app"
	"runtime"
	"github.com/zserge/lorca"
	"net"
	"net/http"
	"fmt"
	"os"
	"os/signal"
	"github.com/sero-cash/go-sero/light-wallet/web"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/sero-cash/go-sero/light-wallet/common/transport"
	"github.com/sero-cash/go-czero-import/cpt"
)

func main() {

	cpt.ZeroInit_OnlyInOuts()

	// Setting global env
	lightApp := app.App{}
	if err := lightApp.Init(); err != nil {
		panic(err)
	}

	// setup log
	log := logex.Log{Name: "light-wallet", Path: app.GetLogPath()}
	if logFile, err := log.Setup(); err != nil {
		panic(err)
	} else {
		defer logFile.Close()
	}

	// init sero light client
	app.NewSeroLight()

	// init ui
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", 960, 768, args...)
	if err != nil {
		logex.Fatal(err)
	}
	defer ui.Close()

	// A simple way to know when UI is ready (uses body.onload event in JS)
	ui.Bind("start", func() {
		logex.Info("UI is ready")
	})

	//banding http handle
	privateAccountApi := app.NewPrivateAccountAPI()
	createAccountHandler := httptransport.NewServer(
		app.MakeAccountCreateEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/create", accessControl(createAccountHandler))

	//upload keystore
	http.HandleFunc("/account/import/keystore", privateAccountApi.UploadKeystoreHandler())

	importWithMnemonicHandler := httptransport.NewServer(
		app.MakeAccountImportWithMnemonicEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/import/mnemonic", accessControl(importWithMnemonicHandler))

	importWithPrivateHandler := httptransport.NewServer(
		app.MakeAccountImportWithPrivateKeyEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/import/private", accessControl(importWithPrivateHandler))

	exportMnemonicHandler := httptransport.NewServer(
		app.MakeAccountExportMnemonicEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/export/mnemonic", accessControl(exportMnemonicHandler))

	accountListHandler := httptransport.NewServer(
		app.MakeAccountListEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/list", accessControl(accountListHandler))

	accountDetailHandler := httptransport.NewServer(
		app.MakeAccountDetailEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/detail", accessControl(accountDetailHandler))

	accountBalanceHandler := httptransport.NewServer(
		app.MakeAccountBalanceEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/account/balance", accessControl(accountBalanceHandler))

	txListHandler := httptransport.NewServer(
		app.MakeTxListEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/tx/list", accessControl(txListHandler))

	txNumHandler := httptransport.NewServer(
		app.MakeTxNumEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/tx/num", accessControl(txNumHandler))

	txTransferHandler := httptransport.NewServer(
		app.MakeTxSendEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/tx/transfer", accessControl(txTransferHandler))

	keyPathandler := httptransport.NewServer(
		app.MakeDataPathEndpoint(privateAccountApi),
		transport.DecodeRequest,
		transport.EncodeResponse,
	)
	http.Handle("/path/keystore", accessControl(keyPathandler))


	//start file server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logex.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(web.FS))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))

	// You may use console.log to debug your JS code, it will be printed via
	// log.Println(). Also exceptions are printed in a similar manner.
	//start http server
	err = http.ListenAndServe(":2345", nil)
	if err != nil {
		logex.Fatal(err)
	}

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	logex.Info("exiting...")

}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}


