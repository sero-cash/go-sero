package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type IData interface {
	Clear()
	AddOut(root *keys.Uint256, out *localdb.OutState)
	AddNil(in *keys.Uint256)
	AddDel(in *keys.Uint256)

	LoadState(tr tri.Tri)
	SaveState(tr tri.Tri)

	HasIn(tr tri.Tri, hash *keys.Uint256) (exists bool)
	GetOut(tr tri.Tri, root *keys.Uint256) (src *localdb.OutState)

	GetRoots() (roots []keys.Uint256)
	GetDels() (dels []keys.Uint256)
	GetIndex() (index int64)
}
