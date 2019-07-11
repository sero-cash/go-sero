package transport

type BaseRequest struct {
	Token     string `json:"token,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Language  string `json:"language"`
	Sign      string `json:"sign,omitempty"`
	AppId     string `json:"app_id,omitempty"`
}