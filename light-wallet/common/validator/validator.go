package validator

import (
	"errors"
	"strconv"
	"time"
	"github.com/sero-cash/go-sero/light-wallet/common/transport"
	"github.com/sero-cash/go-sero/light-wallet/common/utils"
	"github.com/sero-cash/go-sero/light-wallet/common/config"
)

func ValidateBaseRequestParam(base transport.BaseRequest) (bool bool, err error) {
	timestamp := strconv.FormatInt(base.Timestamp, 10)
	nowTime := time.Now().UnixNano() / 1e6

	if nowTime -  base.Timestamp > 300*1000{
		//return false, errors.New("Invalid Timestamp ")
	}
	sign := base.Sign
	baseSign, err := utils.ComputeHmac256(base.Token+base.AppId+timestamp, config.SignKey)

	if err != nil {
		return false, err
	}
	if baseSign != sign {
		return false, errors.New("Signature validate failed. ")
	}
	return true, nil
}