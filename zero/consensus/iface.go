package consensus

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/serodb"
)

type Tri interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
}

type DB interface {
	Num() uint64
	CurrentTri() Tri
	GlobalGetter() serodb.Getter
}

type CItem interface {
	CopyTo() (ret CItem)
	CopyFrom(CItem)
	State() (ret *keys.Uint256)
}
