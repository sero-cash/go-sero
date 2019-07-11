package transport

type PageRequest struct {

	//页号
	PageNo uint8 `json:"page_no,omitempty"`
	// 页面显示的条数
	PageSize uint8 `json:"page_size,omitempty"`

	//排序方式，如果不指定则默认按主键倒序排列
	OrderBy string `json:"order_by,omitempty"`
}