package transport

type PageResponse struct {

	Count uint8 `json:"count,omitempty"`
	//页号
	PageNo uint8 `json:"page_no,omitempty"`
	// 页面显示的条数
	PageSize uint8 `json:"page_size,omitempty"`
	//
	OrderBy string `json:"order_by,omitempty"`
}