package snapshot

import (
	"testing"
)

func TestSnapshot(t *testing.T) {
	sg,_:=NewSnapshotGen(
		"/Users/tangzhige/Documents/gero/prod/datadir/gero/chaindata_bk_2",
		"/Users/tangzhige/Documents/gero/prod/datadir/gero/chaindata_test",
	)
	sg.Run()
	sg.Close()
}