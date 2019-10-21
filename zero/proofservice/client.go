package proofservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/wallet/light"
	"log"
	"net/http"
	"time"
)

type LocalClient struct {
	backend Backend
}

func NewLocalClient(backend Backend) *LocalClient {
	return &LocalClient{backend}
}

func (self *LocalClient) CommitTx(tx *txtool.GTx) error {
	return self.backend.CommitTx(tx)
}
func (self *LocalClient) CheckNils(nils []c_type.Uint256) bool {
	if nilResps, err := self.backend.CheckNil(nils); err == nil {
		return len(nilResps) == 0
	}
	return false;
}

type RemoteClient struct {
	host string
}

func NewRemoteClient(host string) *RemoteClient {
	return &RemoteClient{host}
}

func (self *RemoteClient) CommitTx(tx *txtool.GTx) error {
	if _, err := commitTx(self.host, tx); err != nil {
		return err;
	}
	return nil
}

func (self *RemoteClient) CheckNils(nils []c_type.Uint256) bool {

	return checkNils(self.host, nils);
}

func checkNils(host string, nils []c_type.Uint256) bool {
	resp, err := doPost(host, "light_checkNil", []interface{}{nils})
	if err != nil {
		return false
	}

	nilResps := []light.NilValue{}
	message := *resp.Result
	err = json.Unmarshal(message[:], &nilResps)
	if err != nil {
		return false
	}
	return len(nilResps) == 0
}

func commitTx(host string, tx *txtool.GTx) (string, error) {

	hash := tx.Tx.ToHash()
	_, err := doPost(host, "sero_commitTx", []interface{}{tx})
	if err != nil {
		return common.Bytes2Hex(hash[:]), err
	}

	return common.Bytes2Hex(hash[:]), nil
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

func doPost(url string, method string, params interface{}) (*JSONRpcResp, error) {
	client := &http.Client{
		Timeout: MustParseDuration("900s"),
	}
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": method, "params": params, "id": 0}
	data, err := json.Marshal(jsonReq)
	if err != nil {
		log.Printf(err.Error())
		return nil, err
	}

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
	data, _ = json.Marshal(rpcResp)
	return rpcResp, err
}

func MustParseDuration(s string) time.Duration {
	value, err := time.ParseDuration(s)
	if err != nil {
		panic("util: Can't parse duration `" + s + "`: " + err.Error())
	}
	return value
}
