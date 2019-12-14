package main

import (
	"flag"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/internal/ethapi"
	"github.com/sero-cash/go-sero/rpc"
	"github.com/sero-cash/go-sero/zero/proofservice"
	"github.com/sero-cash/go-sero/zero/utils"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	endpoint := flag.String("addr", ":9765", "listen address")
	rpc, config, tomeOut := initConfig()
	if err := startNode(*endpoint, rpc, config, tomeOut); err != nil {
		os.Exit(0)
	}
	select {}
}

func startNode(endpoint, rpcAddr string, config *proofservice.Config, timeout rpc.HTTPTimeouts) (error) {
	if endpoint == "" {
		return nil
	}
	proofservice.NewProofService(rpcAddr, nil, config)

	apis := []rpc.API{
		{
			Namespace: "proof",
			Version:   "1.0",
			Service:   ethapi.NewProofServiceApi(),
			Public:    true,
		}}

	_, _, err := rpc.StartHTTPEndpoint(endpoint, apis, []string{"proof"}, []string{}, []string{}, timeout)
	if err != nil {
		return err
	}
	log.Printf("HTTP endpoint opened, url %s", fmt.Sprintf("http://%s", endpoint))
	return nil
}

func initConfig() (string, *proofservice.Config, rpc.HTTPTimeouts) {

	var (
		maxWorkNumber  = flag.Int("maxWorkNumber", 10, "max work count")
		maxQueueNumber = flag.Int("maxQueueNumber", 20, "max work pending count")
		readTimeout    = flag.Duration("readTimeout", 120*time.Second, "readTimeout")
		writeTimeout   = flag.Duration("writeTimeout", 120*time.Second, "writeTimeout")
		idleTimeout    = flag.Duration("idleTimeout", 180*time.Second, "idleTimeout")

		rpcAddr = flag.String("rpcAddr", "127.0.0.1:8545", "idleTimeout")

		zinFee   = flag.String("zinFee", "0sero", "zinFee")
		oinFee   = flag.String("oinFee", "0sero", "oinFee")
		outFee   = flag.String("outFee", "0sero", "outFee")
		fixedFee = flag.String("fee", "0sero", "outFee")

		pkrString = flag.String("pkr", "0sero", "outFee")

		endpoint = flag.String("redis", "127.0.0.1:6379", "redis endpoint")
		password = flag.String("password", "", "redis password")
		database = flag.Int64("database", 0, "redis database")
		poolSize = flag.Int("poolSize", 10, "redis poolSize")
	)

	if strings.TrimSpace(*pkrString) == "" {
		panic("pkr is empty")
	}

	zinFeeAmount, err := utils.ParseAmount(*zinFee)
	if err != nil {
		panic(err);
	}
	oinFeeAmount, err := utils.ParseAmount(*oinFee)
	if err != nil {
		panic(err);
	}
	outFeeAmount, err := utils.ParseAmount(*outFee)
	if err != nil {
		panic(err);
	}
	fixedFeeAmount, err := utils.ParseAmount(*fixedFee)
	if err != nil {
		panic(err);
	}

	pkr := c_type.NewPKrByBytes(base58.Decode(*pkrString))
	timeout := rpc.HTTPTimeouts{*readTimeout, *writeTimeout, *idleTimeout}
	fee := proofservice.ServiceFee{zinFeeAmount, oinFeeAmount, outFeeAmount, fixedFeeAmount}
	return *rpcAddr, &proofservice.Config{pkr, *maxWorkNumber, *maxQueueNumber, fee}, timeout
}