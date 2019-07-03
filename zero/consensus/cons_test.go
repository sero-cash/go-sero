package consensus

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sero-cash/go-czero-import/keys"
)

func s2u(str string) (ret []byte) {
	return []byte(str)
}

func TestConsSetValue(t *testing.T) {
	db := NewFakeDB()

	cmap := NewCons(&db, "block")

	tree := NewKVPt(&cmap, "tree", "test")

	tree.SetValue(s2u("k0"), s2u("v0"))
	v := tree.GetValue(s2u("k0"))
	if v == nil || bytes.Compare(v, s2u("v0")) != 0 {
		t.FailNow()
	}
	fmt.Println(v)
}

func TestConsSnapshot(t *testing.T) {
	db := NewFakeDB()
	cmap := NewCons(&db, "block")
	tree := NewKVPt(&cmap, "tree", "test")

	cmap.CreateSnapshot(0)

	tree.SetValue(s2u("k0"), s2u("v0"))
	v := tree.GetValue(s2u("k0"))
	if v == nil || bytes.Compare(v, s2u("v0")) != 0 {
		t.FailNow()
	}

	cmap.CreateSnapshot(1)

	tree.SetValue(s2u("k0"), s2u("v1"))
	v = tree.GetValue(s2u("k0"))
	if v == nil || bytes.Compare(v, s2u("v1")) != 0 {
		t.FailNow()
	}

	cmap.RevertToSnapshot(1)

	v = tree.GetValue(s2u("k0"))
	if v == nil || bytes.Compare(v, s2u("v0")) != 0 {
		t.FailNow()
	}

	cmap.RevertToSnapshot(0)

	v = tree.GetValue(s2u("k0"))
	if v != nil {
		t.FailNow()
	}
}

type TestObj struct {
	I string
	S string
}

func (self *TestObj) CopyTo() CItem {
	r := &TestObj{}
	r.I = self.I
	r.S = self.S
	return r
}
func (self *TestObj) CopyFrom(item CItem) {
	i := item.(*TestObj)
	self.I = i.I
	self.S = i.S
	return
}

func (self *TestObj) Id() (ret []byte) {
	return append([]byte{}, []byte(self.I)...)
}

func (self *TestObj) State() (ret []byte) {
	return append([]byte{}, []byte(self.S)...)
}

func NewTestObj(name string) (ret *TestObj) {
	ret = &TestObj{}
	ret.I = "obj0"
	ret.S = name
	return
}

func TestConsSetObj(t *testing.T) {
	db := NewFakeDB()
	cmap := NewCons(&db, "block")
	tree := NewObjPt(&cmap, "tree$", "treestate$", "test")

	cmap.CreateSnapshot(0)

	tree.AddObj(NewTestObj("0"))
	v := tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "0" {
		t.FailNow()
	}
	fmt.Println(v)

	cmap.CreateSnapshot(1)

	tree.AddObj(NewTestObj("1"))
	v = tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "1" {
		t.FailNow()
	}
	fmt.Println(v)

	cmap.RevertToSnapshot(1)

	v = tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "0" {
		t.FailNow()
	}
	fmt.Println(v)

	cmap.RevertToSnapshot(0)

	v = tree.GetObj(s2u("obj0"), &TestObj{})
	if v != nil {
		t.FailNow()
	}
	fmt.Println(v)

	conslist := cmap.fetchConsPairs(true)
	blocklist := cmap.fetchConsPairs(true)
	dblist := cmap.fetchConsPairs(true)
	fmt.Println(conslist)
	fmt.Println(blocklist)
	fmt.Println(dblist)
}

func TestConsFetch(t *testing.T) {
	db := NewFakeDB()
	cmap := NewCons(&db, "block")
	dbobj := DBObj{"treestate$"}
	tree := NewObjPt(&cmap, "tree$", dbobj.Pre, "test")

	cmap.CreateSnapshot(0)

	tree.AddObj(NewTestObj("0"))
	v := tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "0" {
		t.FailNow()
	}
	fmt.Println(v)

	cmap.CreateSnapshot(1)

	tree.AddObj(NewTestObj("1"))
	v = tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "1" {
		t.FailNow()
	}
	fmt.Println(v)

	conslist := cmap.fetchConsPairs(false)
	if len(conslist) != 1 {
		t.FailNow()
	}
	blocklist := cmap.fetchBlockRecords(false)
	if len(blocklist) != 1 && len(blocklist[0].Hashes) != 2 {
		t.FailNow()
	}
	dblist := cmap.fetchDBPairs(false)
	if len(dblist) != 2 {
		t.FailNow()
	}

	conslist = cmap.fetchConsPairs(true)
	if len(conslist) != 0 {
		t.FailNow()
	}
	blocklist = cmap.fetchBlockRecords(true)
	if len(blocklist) != 0 {
		t.FailNow()
	}
	dblist = cmap.fetchDBPairs(true)
	if len(dblist) != 0 {
		t.FailNow()
	}

}

func TestConsRecord(t *testing.T) {
	db := NewFakeDB()
	dbcons := DBObj{"BLOCK$CONS$INDEX$"}
	cmap := NewCons(&db, dbcons.Pre)
	dbobj := DBObj{"treestate$"}
	tree := NewObjPt(&cmap, "tree$", dbobj.Pre, "test")

	cmap.CreateSnapshot(0)

	tree.AddObj(NewTestObj("0"))
	v := tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "0" {
		t.FailNow()
	}
	fmt.Println(v)

	cmap.CreateSnapshot(1)

	tree.AddObj(NewTestObj("1"))
	v = tree.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "1" {
		t.FailNow()
	}
	fmt.Println(v)

	blockhash := common.Hash(keys.RandUint256())
	cmap.Record(&blockhash, &db.db)

	testObj := TestObj{}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("1"), &testObj)
	if v == nil {
		t.FailNow()
	}
	if testObj.S != "1" {
		t.FailNow()
	}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("0"), &testObj)
	if v == nil {
		t.FailNow()
	}
	if testObj.S != "0" {
		t.FailNow()
	}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("2"), &testObj)
	if v != nil {
		t.FailNow()
	}

	records := dbcons.GetBlockRecords(db.GlobalGetter(), 0, &blockhash)
	if len(records) != 1 && len(records[0].Hashes) != 2 {
		t.FailNow()
	}
	fmt.Println(records)

	cmap.Update()

	cmap1 := NewCons(&db, dbcons.Pre)
	tree1 := NewObjPt(&cmap1, "tree$", dbobj.Pre, "test")
	v = tree1.GetObj(s2u("obj0"), &TestObj{})
	if v.(*TestObj).S != "1" {
		t.FailNow()
	}
	fmt.Println(v)

}
