package transport

import (
	"sero.cash/sero-go/wallet/common/errorcode"
	"net/http"
	"encoding/json"
	"context"
)

type BizResponse interface{}

type Response struct {
	Base BaseResponse `json:"base,omitempty"`
	Page PageResponse `json:"page,omitempty"`
	Biz  BizResponse  `json:"biz,omitempty"`
}

type GenResponse interface {
	SetBaseResponse(baseResponse BaseResponse)
	SetPageResponse(pageResponse PageResponse)
	SetBizResponse(bizResponse BizResponse)
}

func (response *Response) SetBaseResponse(code, desc string) {
	baseResponse := BaseResponse{
		Code: code,
		Desc: desc,
	}

	*response = Response{
		Base: baseResponse,
		Page: response.Page,
		Biz:  response.Biz,
	}
}

func (response *Response) SetBaseResponseSuccess() {
	baseResponse := BaseResponse{
		Code: errorcode.SUCCESS_CODE,
		Desc: errorcode.SUCCESS_DESC,
	}
	*response = Response{
		Base: baseResponse,
	}
}

func (response *Response) SetPageResponse(count, pageSize, pageNo uint8, orderBy string) {
	pageResponse := PageResponse{
		Count:    count,
		PageSize: pageSize,
		PageNo:   pageNo,
	}
	*response = Response{
		Base: response.Base,
		Page: pageResponse,
		Biz:  response.Biz,
	}
}

func (response *Response) SetBizResponse(bizResponse BizResponse) {
	*response = Response{
		Base: response.Base,
		Page: response.Page,
		Biz:  bizResponse,
	}
}


type BizResponseX struct {
	Bool bool `json:"bool"`
	Code string `json:"code"`
	Message string `json:"message"`
}
func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
