package consensus

import (
	"github.com/sero-cash/go-sero/serodb"
)

type DB interface {
	CurrentTri() serodb.Tri
	GlobalGetter() serodb.Getter
}

type CItem interface {
	CopyTo() (ret CItem)
	CopyFrom(CItem)
}

type PItem interface {
	CItem
	Id() (ret []byte)
	State() (ret []byte)
}
