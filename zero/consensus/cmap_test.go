package consensus

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-czero-import/keys"
)

func TestCMap(t *testing.T) {
	db := NewFakeDB()

	cmap := NewCons(&db)

	tree := cmap.CreatePoint("tree", "", true)

	tree.SetValue(&keys.Uint256{'k', '0'}, &keys.Uint256{'v', '0'})
	v := tree.GetValue(&keys.Uint256{'k', '0'})
	fmt.Println(v)
}
