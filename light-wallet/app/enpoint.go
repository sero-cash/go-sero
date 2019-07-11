package app

import (
	"github.com/go-kit/kit/endpoint"
	"context"
	"github.com/sero-cash/go-sero/light-wallet/common/transport"
	"github.com/sero-cash/go-sero/light-wallet/common/validator"
	"github.com/sero-cash/go-sero/light-wallet/common/errorcode"
	"github.com/sero-cash/go-sero/light-wallet/common/utils"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	"fmt"
)

type AccountCreateReq struct {
	Passphrase string `json:"passphrase"`
}

func MakeAccountCreateEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}

		accountCreateReq := AccountCreateReq{}
		utils.Convert(req.Biz, &accountCreateReq)

		if accountCreateReq.Passphrase == "" {
			response.SetBaseResponse(errorcode.FAIL_CODE, "passphrase is nil")
			return response, nil
		}

		resp, err := service.NewAccountWithMnemonic(accountCreateReq.Passphrase)
		logex.Info(resp, err)
		if err != nil {
			response.SetBaseResponse(errorcode.FAIL_CODE, err.Error())
		} else {
			response.SetBizResponse(resp)
		}
		return response, nil
	}
}

type accountImportWithMnemonicReq struct {
	Mnemonic   string `json:"mnemonic"`
	Passphrase string `json:"passphrase"`
}

func MakeAccountImportWithMnemonicEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}
		aimq := accountImportWithMnemonicReq{}
		utils.Convert(req.Biz, &aimq)

		resp, err := service.ImportAccountFromMnemonic(aimq.Mnemonic, aimq.Passphrase)
		if err != nil {
			response.SetBaseResponse(errorcode.FAIL_CODE, err.Error())
		} else {
			response.SetBizResponse(resp)
		}

		return response, nil
	}
}

type accountImportWithPrivateKeyReq struct {
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}

func MakeAccountImportWithPrivateKeyEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}

		aipq := accountImportWithPrivateKeyReq{}
		utils.Convert(req.Biz, &aipq)

		resp, err := service.ImportAccountFromRawKey(aipq.PrivateKey, aipq.Passphrase)
		if err != nil {
			response.SetBaseResponse(errorcode.FAIL_CODE, err.Error())
		} else {
			response.SetBizResponse(resp)
		}

		return response, nil
	}
}

type accountExportMnemonic struct {
	Passphrase string `json:"passphrase"`
	Address    string `json:"address"`
}

func MakeAccountExportMnemonicEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}

		aem := accountExportMnemonic{}
		utils.Convert(req.Biz, &aem)

		resp, err := service.ExportMnemonic(aem.Address, aem.Passphrase)
		if err != nil {
			response.SetBaseResponse(errorcode.FAIL_CODE, err.Error())
		} else {
			response.SetBizResponse(resp)
		}

		return response, nil
	}
}

// assets main
func MakeAccountListEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}
		response.SetBizResponse(service.AccountList())

		return response, nil
	}
}

type pk struct {
	PK string
}

// assets main
func MakeAccountDetailEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}
		pk := pk{}
		utils.Convert(req.Biz, &pk)
		ac := service.AccountDetail(pk.PK)
		response.SetBizResponse(ac)
		return response, nil
	}
}

func MakeAccountBalanceEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}
		pk := pk{}
		utils.Convert(req.Biz, &pk)
		balance := service.AccountBalance(pk.PK)
		for key, v := range balance {
			fmt.Println("balance == ", key, v.String())
		}
		response.SetBizResponse(balance)

		return response, nil
	}
}

func MakeTxListEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}
		pk := pk{}
		utils.Convert(req.Biz, &pk)
		if records, err := service.TXList(pk.PK, req.Page); err != nil {
			response.SetBaseResponse(errorcode.FAIL_CODE, err.Error())
			return response, nil
		} else {
			response.SetBizResponse(records)
		}
		return response, nil
	}
}

func MakeTxNumEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}
		pk := pk{}
		utils.Convert(req.Biz, &pk)
		response.SetBizResponse(service.TXNum(pk.PK))
		return response, nil
	}
}

func MakeTxSendEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transport.Request)
		response := transport.Response{}
		response.SetBaseResponseSuccess()

		if ok, err := validator.ValidateBaseRequestParam(req.Base); !ok {
			response.SetBaseResponse(errorcode.InvalidBaseParameters, err.Error())
			return response, nil
		}

		transferReq := transferReq{}
		utils.Convert(req.Biz, &transferReq)

		hash, err := service.Transfer(transferReq.From, transferReq.To, transferReq.Currency, transferReq.Amount, transferReq.GasPrice)
		if err != nil {
			response.SetBaseResponse(errorcode.FAIL_CODE, err.Error())
			return response, nil
		}
		response.SetBizResponse(hash)
		return response, nil
	}
}

type transferReq struct {
	From     string
	To       string
	Currency string
	Amount   string
	GasPrice string
}
