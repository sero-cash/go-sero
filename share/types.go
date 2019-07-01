package share

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
)

type Lottery struct {
	ParentHash common.Hash
	PosHash    common.Hash
}

type Vote struct {
	PosHash common.Hash
	Sign    keys.Uint512
}
