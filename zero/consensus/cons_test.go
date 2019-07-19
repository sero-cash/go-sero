package consensus

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/sero-cash/go-sero/core/types"
)

func s2u(str string) (ret []byte) {
	return []byte(str)
}

func TestConsSetValue(t *testing.T) {
	db := NewFakeDB()

	cmap := NewCons(&db, "block")

	tree := NewKVPt(&cmap, "tree", "test")

	tree.SetValue(s2u("k0"), s2u("v0"))
	tree.SetValue(s2u("k1"), s2u("v1"))
	v := tree.GetValue(s2u("k0"))
	if v == nil || bytes.Compare(v, s2u("v0")) != 0 {
		t.FailNow()
	}
	v = tree.GetValue(s2u("k1"))
	if v == nil || bytes.Compare(v, s2u("v1")) != 0 {
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

func NewTestObj2(id string, name string) (ret *TestObj) {
	ret = &TestObj{}
	ret.I = id
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
	blocklist := cmap.fetchBlockRecords()
	if len(blocklist) != 1 && len(blocklist[0].Pairs) != 1 {
		t.FailNow()
	}
	dblist := cmap.fetchDBPairs()
	if len(dblist) != 1 {
		t.FailNow()
	}

	conslist = cmap.fetchConsPairs(true)
	if len(conslist) != 0 {
		t.FailNow()
	}
	blocklist = cmap.fetchBlockRecords()
	if len(blocklist) != 1 && len(blocklist[0].Pairs) != 1 {
		t.FailNow()
	}
	dblist = cmap.fetchDBPairs()
	if len(dblist) != 1 {
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

	tree.AddObj(NewTestObj2("obj1", "3"))
	v = tree.GetObj(s2u("obj1"), &TestObj{})
	if v.(*TestObj).S != "3" {
		t.FailNow()
	}
	fmt.Println(v)

	header := types.Header{Number: big.NewInt(0)}
	cmap.Record(&header, &db.db)

	testObj := TestObj{}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("1"), &testObj)
	if v == nil {
		t.FailNow()
	}
	if testObj.S != "1" {
		t.FailNow()
	}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("0"), &testObj)
	if v != nil {
		t.FailNow()
	}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("2"), &testObj)
	if v != nil {
		t.FailNow()
	}

	hash := header.Hash()
	records := dbcons.GetBlockRecords(db.GlobalGetter(), 0, &hash)
	if len(records) != 1 && len(records[0].Pairs) != 2 {
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
	v = tree1.GetObj(s2u("obj1"), &TestObj{})
	if v.(*TestObj).S != "3" {
		t.FailNow()
	}
	fmt.Println(v)

}

func TestConsRecord2(t *testing.T) {
	db := NewFakeDB()
	dbcons := DBObj{"BLOCK$CONS$INDEX$"}
	c := NewCons(&db, dbcons.Pre)
	cmap := &c
	dbobj := DBObj{"treestate$"}
	tree := NewObjPt(cmap, "tree$", dbobj.Pre, "test")

	cmap.CreateSnapshot(0)

	tree.AddObj(NewTestObj2("obj1", "11"))
	tree.AddObj(NewTestObj2("obj1", "12"))
	tree.AddObj(NewTestObj2("obj1", "13"))
	tree.AddObj(NewTestObj2("obj2", "21"))
	tree.AddObj(NewTestObj2("obj2", "22"))
	tree.AddObj(NewTestObj2("obj3", "33"))

	cmap.CreateSnapshot(1)
	tree.AddObj(NewTestObj2("obj1", "14"))
	tree.AddObj(NewTestObj2("obj2", "25"))
	tree.AddObj(NewTestObj2("obj3", "36"))

	v := tree.GetObj(s2u("obj1"), &TestObj{})
	if v.(*TestObj).S != "14" {
		t.FailNow()
	}
	v = tree.GetObj(s2u("obj2"), &TestObj{})
	if v.(*TestObj).S != "25" {
		t.FailNow()
	}
	v = tree.GetObj(s2u("obj3"), &TestObj{})
	if v.(*TestObj).S != "36" {
		t.FailNow()
	}

	cmap.RevertToSnapshot(1)

	v = tree.GetObj(s2u("obj1"), &TestObj{})
	if v.(*TestObj).S != "13" {
		t.FailNow()
	}
	v = tree.GetObj(s2u("obj2"), &TestObj{})
	if v.(*TestObj).S != "22" {
		t.FailNow()
	}
	v = tree.GetObj(s2u("obj3"), &TestObj{})
	if v.(*TestObj).S != "33" {
		t.FailNow()
	}

	cmap = cmap.Copy(cmap.db)

	header := types.Header{Number: big.NewInt(0)}
	cmap.Record(&header, &db.db)

	testObj := TestObj{}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("13"), &testObj)
	if v == nil {
		t.FailNow()
	}
	if testObj.I != "obj1" {
		t.FailNow()
	}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("22"), &testObj)
	if v == nil {
		t.FailNow()
	}
	if testObj.I != "obj2" {
		t.FailNow()
	}
	v = dbobj.GetObject(db.GlobalGetter(), []byte("33"), &testObj)
	if v == nil {
		t.FailNow()
	}
	if testObj.I != "obj3" {
		t.FailNow()
	}

	//========

	v = dbobj.GetObject(db.GlobalGetter(), []byte("11"), &testObj)
	if v != nil {
		t.FailNow()
	}

	v = dbobj.GetObject(db.GlobalGetter(), []byte("21"), &testObj)
	if v != nil {
		t.FailNow()
	}

	hash := header.Hash()
	records := dbcons.GetBlockRecords(db.GlobalGetter(), 0, &hash)
	if len(records) != 1 && len(records[0].Pairs) != 3 {
		t.FailNow()
	}
	fmt.Println(records)

	cmap.Update()

	cmap1 := NewCons(&db, dbcons.Pre)
	tree1 := NewObjPt(&cmap1, "tree$", dbobj.Pre, "test")
	v = tree1.GetObj(s2u("obj1"), &TestObj{})
	if v.(*TestObj).S != "13" {
		t.FailNow()
	}
	v = tree1.GetObj(s2u("obj2"), &TestObj{})
	if v.(*TestObj).S != "22" {
		t.FailNow()
	}
	v = tree1.GetObj(s2u("obj3"), &TestObj{})
	if v.(*TestObj).S != "33" {
		t.FailNow()
	}
	fmt.Println(v)

}
