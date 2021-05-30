package snapshot

import (
	"testing"
)


func TestSnapshot(t *testing.T) {
	sg,_:=NewSnapshotGen(
		"/Users/tangzhige/Documents/gero/prod/datadir/gero/chaindata_kk",
		"/Users/tangzhige/Documents/gero/prod/datadir/gero/chaindata",
	)
	sg.Process(int(sg.src_head_num+1))
}