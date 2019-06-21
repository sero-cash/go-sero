package transport

import (
	"encoding/json"
	"context"
	"net/http"
)

type BizRequest interface{}

type Request struct {
	Base BaseRequest `json:"base,omitempty"`
	Page PageRequest `json:"page,omitempty"`
	Biz  BizRequest  `json:"biz,omitempty"`
}

type GenRequest interface {
	SetBaseRequest(base BaseRequest) Request
	SetPageRequest(page PageRequest) Request
	SetBizRequest(biz BizRequest) Request
}

func (request *Request) SetBizRequest(biz BizRequest) error {
	*request = Request{
		Base: request.Base,
		Page: request.Page,
		Biz:  biz,
	}
	return nil
}

func (request *Request) SetBaseRequest(base BaseRequest) error {
	*request = Request{
		Base: base,
		Page: request.Page,
		Biz:  request.Biz,
	}
	return nil
}

func (request *Request) SetPageRequest(page PageRequest) error {
	*request = Request{
		Base: request.Base,
		Page: page,
		Biz:  request.Biz,
	}
	return nil
}

func DecodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}