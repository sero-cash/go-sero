package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"time"
	"github.com/sero-cash/go-sero/light-wallet/common/config"
)

func GetTimestamp() (timestamp int64) {

	return time.Now().Unix()
}

func GetTimestampString() (timestamp string) {

	return strconv.FormatInt(time.Now().Unix(), 10)
}

func SignStr(args ... string) (string, error) {
	sortData := sort.StringSlice(args[0:])
	sort.Sort(sortData)
	var orgStr string
	for _, arg := range sortData {
		fmt.Println(arg)
		orgStr += arg
	}
	return ComputeHmac256(orgStr, config.SignKey)
}

func ComputeHmac256(message string, secret string) (string, error) {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)),nil
}
