package consensus

import (
	"bytes"
	"fmt"
	"testing"
)

func s2u(str string) (ret []byte) {
	return []byte(str)
}

func TestConsSetValue(t *testing.T) {
	db := NewFakeDB()

	cmap := NewCons(&db)

	tree := cmap.CreatePoint("tree", "", true)

	tree.SetValue(s2u("k0"), s2u("v0"))
	v := tree.GetValue(s2u("k0"))
	if v == nil || bytes.Compare(v, s2u("v0")) != 0 {
		t.FailNow()
	}
	fmt.Println(v)
}

func TestConsSnapshot(t *testing.T) {
	db := NewFakeDB()
	cmap := NewCons(&db)
	tree := cmap.CreatePoint("tree", "", true)

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
	cmap := NewCons(&db)
	tree := cmap.CreatePoint("tree$", "treestate$", true)

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
}
