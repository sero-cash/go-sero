package consensus

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-czero-import/keys"
)

func s2u(str string) (ret *keys.Uint256) {
	ret = &keys.Uint256{}
	copy(ret[:], str[:])
	return
}

func u2s(u *keys.Uint256) (ret string) {
	ret = string(u[:])
	return
}

func TestConsSetValue(t *testing.T) {
	db := NewFakeDB()

	cmap := NewCons(&db)

	tree := cmap.CreatePoint("tree", "", true)

	tree.SetValue(s2u("k0"), s2u("v0"))
	v := tree.GetValue(s2u("k0"))
	if v == nil || *v != *s2u("v0") {
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
	if v == nil || *v != *s2u("v0") {
		t.FailNow()
	}

	cmap.CreateSnapshot(1)

	tree.SetValue(s2u("k0"), s2u("v1"))
	v = tree.GetValue(s2u("k0"))
	if v == nil || *v != *s2u("v1") {
		t.FailNow()
	}

	cmap.RevertToSnapshot(1)

	v = tree.GetValue(s2u("k0"))
	if v == nil || *v != *s2u("v0") {
		t.FailNow()
	}

	cmap.RevertToSnapshot(0)

	v = tree.GetValue(s2u("k0"))
	if v != nil {
		t.FailNow()
	}
}

func TestConsSetObj(t *testing.T) {

}
