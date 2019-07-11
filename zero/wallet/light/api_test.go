package light

import (
	"testing"
	"fmt"
	"log"
	"encoding/json"
	"time"
	"net/http"
	"bytes"
)

var (
	host = "http://127.0.0.1:8545"
)

/**
curl -i -H 'Content-Type:application/json' -d '{"jsonrpc":"2.0","id":2,"method":"light_getOutsByPKr","params":["2HVVk8CN2qTSTKUd4GgqKN4nkMMA4KjE97PwczEv9nkz5VEKG6P5kiLbYMHM7ihKcY36MSxMpCvxGt8zqzwpS7ag5pq4EweHzNpUtjYBGJov4hgcEGf26DRGrTTiR2nSeNm6",0,200]}' http://127.0.0.1:8545
 */

func TestLightNode_GetOutsByPKr(t *testing.T) {
	GetOutsByPKr("2HVVk8CN2qTSTKUd4GgqKN4nkMMA4KjE97PwczEv9nkz5VEKG6P5kiLbYMHM7ihKcY36MSxMpCvxGt8zqzwpS7ag5pq4EweHzNpUtjYBGJov4hgcEGf26DRGrTTiR2nSeNm6",0,100)
}

func GetOutsByPKr(pkr string, begin, end uint64) {
	resp, e := doPost(host, "light_getOutsByPKr", []interface{}{pkr, begin, end})
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
		Timeout: 900 * time.Second,
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
		return nil, fmt.Errorf(rpcResp.Error["message"].(string))

	}
	return rpcResp, err
}
