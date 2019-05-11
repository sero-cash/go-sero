package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)


type Log interface {
	Op(state IData);
}

type AddTxOutLog struct {
	Pkr *keys.PKr
}

func (log AddTxOutLog) Op(state IData) {
	state.AddTxOut(log.Pkr)
}

type AddOutLog struct {
	Root   *keys.Uint256
	Out    *localdb.OutState
	Txhash *keys.Uint256
}

func (log AddOutLog) Op(state IData) {
	state.AddOut(log.Root, log.Out, log.Txhash)
}

type AddNilLog struct {
	In *keys.Uint256
}

func (log AddNilLog) Op(state IData) {
	state.AddNil(log.In)
}

type AddDelLog struct {
	In *keys.Uint256
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

	AddTxOut(pkr *keys.PKr) int
	AddOut(root *keys.Uint256, out *localdb.OutState, txhash *keys.Uint256)
	AddNil(in *keys.Uint256)
	AddDel(in *keys.Uint256)

	LoadState(tr tri.Tri)
	SaveState(tr tri.Tri)
	RecordState(putter serodb.Putter, root *keys.Uint256)

	HasIn(tr tri.Tri, hash *keys.Uint256) (exists bool)
	GetOut(tr tri.Tri, root *keys.Uint256) (src *localdb.OutState)

	GetRoots() (roots []keys.Uint256)
	GetDels() (dels []keys.Uint256)
}