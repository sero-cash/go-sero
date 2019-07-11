package exchange

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"testing"

	"github.com/sero-cash/mine-pool/util"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/base58"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/wallet/light/light_issi"
	"github.com/sero-cash/go-sero/zero/wallet/light/light_types"
)

var (
	host = "http://192.168.15.220:8545"
	//host = "http://127.0.0.1:8545"
)

type Kr struct {
	SKr keys.PKr
	PKr keys.PKr
}

type Reception1 struct {
	Addr     string
	Currency string
	Value    *big.Int
}
type TxParam1 struct {
	From       string
	Gas        uint64
	GasPrice   uint64
	Receptions []Reception1
	Roots      []keys.Uint256
}

func TestExchange_GenTx(t *testing.T) {
	pkr1 := CreatePkr("cKDHK7pH8uG3HNuqNoixnHY2GZAqZWPkcXVQ2vDu9eJ79DxFTFkpWMJY6BxgnzG41ZUXB5F5xLnfGNRHy9VeY6a", 100)
	amount1, _ := new(big.Int).SetString("1000000000000000000", 10)
	param := TxParam1{"38AZd6GH7QCQav2BLfLmcJSu8cGWMDCtGafvN3vmmzraK2e8sQUE2NGk7f2hz3j5Nb6koQffVycHTBBZftZcbAjJ", 25000, 1000000000, []Reception1{{pkr1, "SERO", amount1}}, []keys.Uint256{}}
	GenTx(param)
	fmt.Println()

	//
	//  }
	//
	//wait := sync.WaitGroup{}
	//for i := 0; i < 20; i++ {
	// wait.Add(2)
	// go func() {
	//
	//  pkr1 := CreatePkr("3vA1bUyr7boSojrHve8FDpiZjAJtxfChM8akgznXsMhxoPiqvRPLadoKEkYnec46YVKcFWh4aknKxNRAXWYavUB3", 100+uint64(i))
	//
	//  amount1, _ := new(big.Int).SetString("1000000000000000000", 10)
	//  param := TxParam{"bgRgAisuE7fEreG2JSAKhz8b5QHY6JdBikpGzWwKY5AKTot1UP8hjAQGL8LoonSEpGaJa1bHjTz2dAWnD8DZRf1", 25000, 1000000000, []Reception{{pkr1, "SERO", amount1}}, []keys.Uint256{}}
	//  tx, err := GenTxWithSign(param)
	//  if err == nil {
	//   CommitTransaction(tx)
	//
	//  }
	//  wait.Done()
	// }()
	//
	// go func() {
	//
	//  amount2, _ := new(big.Int).SetString("5000000000", 10)
	//  pkr2 := CreatePkr("2pYYPeCEgqqtQ5hVY9WFpXfsuF1yJhA99ZkhZmb4pnmB6Lt68mjWRpZJeZhEzhZrwsPKtT6gATKuwzXHp3A6hJGa", 100+uint64(i))
	//  param := TxParam{"3vA1bUyr7boSojrHve8FDpiZjAJtxfChM8akgznXsMhxoPiqvRPLadoKEkYnec46YVKcFWh4aknKxNRAXWYavUB3", 45000, 1000000000, []Reception{{pkr2, "SERO", amount2}}, []keys.Uint256{}}
	//  tx, err := GenTxWithSign(param)
	//  if err == nil {
	//   CommitTransaction(tx)
	//
	//  }
	//  wait.Done()
	// }()
	//
	// wait.Wait()
	// fmt.Println("finished :", i)
	//}

	//
	//GenTx(param)
	//GetBalances("bgRgAisuE7fEreG2JSAKhz8b5QHY6JdBikpGzWwKY5AKTot1UP8hjAQGL8LoonSEpGaJa1bHjTz2dAWnD8DZRf1")
	//GetBalances("3vA1bUyr7boSojrHve8FDpiZjAJtxfChM8akgznXsMhxoPiqvRPLadoKEkYnec46YVKcFWh4aknKxNRAXWYavUB3")
	//GetBalances("2pYYPeCEgqqtQ5hVY9WFpXfsuF1yJhA99ZkhZmb4pnmB6Lt68mjWRpZJeZhEzhZrwsPKtT6gATKuwzXHp3A6hJGa")
	////GetBalances(CreatePkr("3vA1bUyr7boSojrHve8FDpiZjAJtxfChM8akgznXsMhxoPiqvRPLadoKEkYnec46YVKcFWh4aknKxNRAXWYavUB3",100))
	//GetBalances(CreatePkr("3vA1bUyr7boSojrHve8FDpiZjAJtxfChM8akgznXsMhxoPiqvRPLadoKEkYnec46YVKcFWh4aknKxNRAXWYavUB3",101))

	GetBalances("cKDHK7pH8uG3HNuqNoixnHY2GZAqZWPkcXVQ2vDu9eJ79DxFTFkpWMJY6BxgnzG41ZUXB5F5xLnfGNRHy9VeY6a")
	GetBalances("38AZd6GH7QCQav2BLfLmcJSu8cGWMDCtGafvN3vmmzraK2e8sQUE2NGk7f2hz3j5Nb6koQffVycHTBBZftZcbAjJ")
	//GetBalances("2vUQSffbmfUescZtpU4L2y2ihCJiyXyU7mYkQLK7m6DqemTz8HvSyMrQobKC4KfZ9a7TFx5kaiu9Pg4cKrHf6nG2")
	//GetBalances("fpn3qNUDEfC2s5j6v9dQqkzn4kPWtFJL2e474qLd4uSV6Ef3TpFjfrWNiKrLnDdZJi86iDWvmyn4ie8czj4DyWm")
	//GetBalances("5aVMAteeAQkjEC8bptMaUqW9ipwbVrz6Qg3Qb7idworuNihhpbbmRnhb5cUWsMwomZWY1PbVqRqVivpmDKaE6NT2")
	//GetBalances("5qi8KvZzqQdg8UUBFtY7iiG2w6iVBg7QkJ6F8sJFTpGZ3hNg8zRgnBnSdSEfLb77D7gYjhzSRBbt1AvbaWEpAZ7L")
	//GetBalances("5fxEtkKDnKwKi8n8V2GQqybH9fPsDh83wex6S9RxjmPnVHTbjrVatqdL3jajxPAiquq6g3XQNx58cpFhcvfYmpzr")
	//GetBalances("3dpiRseha1cLaNJTXoSTmW9ZMJnf3eU9zLXne5vqqeirUvMRQiPss4G19YaTeytzRYskAPc33VdzXjoVFJV6evZB")
	//GetBalances("3s8LRmddPTptrDkCFpeFgWCUSSezsYu3c2WCnVaYvuBXHtGQFka3EstkxQLpcYRxP4PhxfZr4cY1uuMhPuhP6Kwt")

	//GetRecords(pkr, 1, 100)
	//str := "{\"From\":{\"PKr\":\"0x51d921e7ef535d49c8fbe5532c63a6a57436c0feca5ed759eff1f2d49f12502e76c7a31945c4c192ebfc7d56db0c2a1d521f1250f6ee1bcc118a896d6908ecaa03f5b5c60e7455099c3a1b1ea7d5f78318c9f777cf37878a990ffc21533ce288\",\"SKr\":\"0xe5e6eb7107d0975128c7a17a81ea99eb5dc39249e24527150b9ea73be36d2904d00f37e6af2335f3c1351af3b31ca642ea5c069f50e1b3a6e929386a8b39170049df2ec0be67c60114685a35119e58719e5b63b1ea63aa843a59775c98074ab7\"},\"Gas\":25000,\"GasPrice\":1000000000,\"Ins\":[{\"Root\":\"0x2a8e9f898e74ddf5f711701db651bcef8dbf9541014863315aac07a71c060b08\",\"SKr\":\"0xe5e6eb7107d0975128c7a17a81ea99eb5dc39249e24527150b9ea73be36d2904d00f37e6af2335f3c1351af3b31ca642ea5c069f50e1b3a6e929386a8b39170049df2ec0be67c60114685a35119e58719e5b63b1ea63aa843a59775c98074ab7\"}],\"Outs\":[{\"Asset\":{\"Tkn\":{\"Currency\":\"0x000000000000000000000000000000000000000000000000000000005345524f\",\"Value\":599999975000000000000}},\"Memo\":\"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\",\"PKr\":\"0x639a4685f6b06523a00ff4d1467b915fc90e8945432e75d0f42eab5ade49e122515c60c72f48d3b023c6ce58684ff559e5c546d97e99370c0fe4f5b40c152a8149281340a624ddcacf711359ee4e8a7ae0bdcc27f3e4a594485338251ba26da8\"}]}"
	//var param light_issi.GenTxParam
	//err := json.Unmarshal([]byte(str), &param)
	//log.Printf("%v", err)

	//kr := CreateKr()
	//SendTx(kr.PKr, 1000000000000000000)
	//time.Sleep(time.Second * 30)
	//blocks := GetBlocksInfo()
	//var param light_issi.GenTxParam
	//for _, block := range blocks {
	// for _, out := range block.Outs {
	//  if out.PKr == kr.PKr {
	//   param.Ins = append(param.Ins, light_issi.GIn{kr.SKr, out.Root})
	//  }
	// }
	//}
	//param.Outs = []light_types.GOut{light_types.GOut{PKr: kr.PKr, Asset: assets.Asset{Tkn: &assets.Token{
	// Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
	// Value:    utils.U256(*big.NewInt(10)),
	//}}}}
	//
	//param.From = light_types.Kr{kr.SKr, kr.PKr}
	//param.Gas = 25000
	//param.GasPrice = 1000000000
	//txhash := GenTx(param)
	//fmt.Println(common.Bytes2Hex(txhash[:]))

	//hex2Bytes := common.Hex2Bytes("de4ad42c3b66f6baf6a9c145890406a8a2f5a9687b5dc2f6161a374db56ec153")
	//txhash := keys.Uint256{}
	//copy(txhash[:], hex2Bytes)
	//CommitTx(&txhash)nil

	//GetDetailInfo([]string{"0x7e1b85423f3d5b95aa93625178425a49ee3b09e2727292f9d1a464d5377ffbac"}, "0x51d921e7ef535d49c8fbe5532c63a6a57436c0feca5ed759eff1f2d49f12502e76c7a31945c4c192ebfc7d56db0c2a1d521f1250f6ee1bcc118a896d6908ecaa03f5b5c60e7455099c3a1b1ea7d5f78318c9f777cf37878a990ffc21533ce288")
}

func SendTx(to keys.PKr, amount uint64) {
	params := map[string]string{
		"from":  "bgRgAisuE7fEreG2JSAKhz8b5QHY6JdBikpGzWwKY5AKTot1UP8hjAQGL8LoonSEpGaJa1bHjTz2dAWnD8DZRf1",
		"to":    base58.EncodeToString(to[:]),
		"value": hexutil.EncodeUint64(amount),
	}
	resp, e := doPost("http://127.0.0.1:8545", "sero_sendTransaction", []interface{}{params})

	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	fmt.Println(string(message))
}

func CreateKr() (kr Kr) {
	resp, e := doPost(host, "ssi_createKr", nil)
	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	e = json.Unmarshal(message[:], &kr)
	fmt.Println(string(message))
	return kr
}

func GetBlocksInfo() (blocks []light_issi.Block) {
	resp, e := doPost(host, "ssi_getBlocksInfo", []string{hexutil.EncodeUint64(0), hexutil.EncodeUint64(100)})
	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	e = json.Unmarshal(message[:], &blocks)
	fmt.Println(string(message[:]))
	return blocks
}

func GetDetailInfo(outs []string, skr string) (blocks []light_issi.Block) {
	resp, e := doPost(host, "ssi_detail", []interface{}{outs, skr})
	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	e = json.Unmarshal(message[:], &blocks)
	log.Printf(string(message[:]))
	return blocks
}

func CreatePkr(addr string, index uint64) string {
	resp, e := doPost(host, "sero_getPkr", []interface{}{addr, index})
	if e != nil {
		fmt.Println(e)
		return ""
	}
	message := *resp.Result
	var str string
	e = json.Unmarshal(message[:], &str)
	fmt.Println(str)
	datas, e := hexutil.Decode(str)
	s := base58.EncodeToString(datas)
	log.Println(s)
	return s
}

func GenTx(param TxParam1) (hash keys.Uint256) {
	resp, e := doPost(host, "sero_genTx", []interface{}{param})
	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	//var tx light_types.GTx
	//e = json.Unmarshal(message[:], &tx)
	log.Printf(string(message))
	return
}

func GenTxWithSign(param TxParam1) (tx light_types.GTx, err error) {
	resp, e := doPost(host, "sero_genTxWithSign", []interface{}{param})
	if e != nil {
		err = e
		fmt.Println(e)
		return
	}
	message := *resp.Result
	e = json.Unmarshal(message[:], &tx)
	log.Printf(string(message))
	return
}

func GetBalances(pkr string) (balances map[string]*big.Int) {
	resp, e := doPost(host, "sero_getBalances", []interface{}{pkr})
	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	e = json.Unmarshal(message[:], &balances)
	log.Printf(string(message))
	return
}

func GetRecords(pkr string, begin, end uint64) {
	resp, e := doPost(host, "sero_getRecords", []interface{}{pkr, begin, end})
	if e != nil {
		fmt.Println(e)
		return
	}
	message := *resp.Result
	if message == nil {
		return
	}
	//e = json.Unmarshal(message[:], &balances)
	log.Printf(string(message))
	return
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

func doPost(url string, method string, params interface{}) (*JSONRpcResp, error) {
	client := &http.Client{
		Timeout: util.MustParseDuration("900s"),
	}
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": method, "params": params, "id": 0}
	data, err := json.Marshal(jsonReq)
	if err != nil {
		log.Printf(err.Error())
		return nil, err
	}
	log.Printf(string(data))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp *JSONRpcResp
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {

		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, errors.New(rpcResp.Error["message"].(string))

	}
	return rpcResp, err
}
