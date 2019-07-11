package utils

import (
	"github.com/satori/go.uuid"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
)

func UUID() string {
	u,e:=uuid.NewV4()
	if e != nil {
		logex.Error("gen UUID fail,err:",e)
		return ""
	}
	return u.String()
}