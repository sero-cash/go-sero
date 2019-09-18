package data

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type Log interface {
	Op(state IData)
}

type AddTxOutLog struct {
	Pkr *c_type.PKr
}

func (log AddTxOutLog) Op(state IData) {
	state.AddTxOut(log.Pkr)
}

type AddOutLog struct {
	Root   *c_type.Uint256
	Out    *localdb.OutState
	Txhash *c_type.Uint256
}

func (log AddOutLog) Op(state IData) {
	state.AddOut(log.Root, log.Out, log.Txhash)
}

type AddNilLog struct {
	In *c_type.Uint256
}

func (log AddNilLog) Op(state IData) {
	state.AddNil(log.In)
}

type AddDelLog struct {
	In *c_type.Uint256
}

func (log AddDelLog) Op(state IData) {
	state.AddDel(log.In)
}

type Revision struct {
	Id           int
	JournalIndex int
}

type IData interface {
	Clear()

	AddTxOut(pkr *c_type.PKr) int
	AddOut(root *c_type.Uint256, out *localdb.OutState, txhash *c_type.Uint256)
	AddNil(in *c_type.Uint256)
	AddDel(in *c_type.Uint256)

	LoadState(tr tri.Tri)
	SaveState(tr tri.Tri)
	RecordState(putter serodb.Putter, root *c_type.Uint256)

	HasIn(tr tri.Tri, hash *c_type.Uint256) (exists bool)
	GetOut(tr tri.Tri, root *c_type.Uint256) (src *localdb.OutState)

	GetRoots() (roots []c_type.Uint256)
	GetDels() (dels []c_type.Uint256)
}
