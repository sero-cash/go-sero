package snapshot

import (
	"testing"
)

func TestSnapshot(t *testing.T) {
	sg,_:=NewSnapshotGen(
		"/Users/tangzhige/Documents/Env/gero/data0/datadir/gero/chaindata",
		"/Users/tangzhige/Documents/Env/gero/data0/datadir/gero/chaindata_bk",
	)
	sg.Run()
	sg.Close()
}