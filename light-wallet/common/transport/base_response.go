package transport

type BaseResponse struct {
	Code string `json:"code,omitempty"`
	Desc string `json:"desc,omitempty"`
}
