package consensus

import (
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
	Id() (ret []byte)
}

type PItem interface {
	CItem
	State() (ret []byte)
}
