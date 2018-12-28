package state

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/serodb"
)

func TestOutTree(t *testing.T) {
	// Create an empty state database
	cpt.ZeroInit("", 0)
	db := serodb.NewMemDatabase()
	statedb, _ := New(common.Hash{}, NewDatabase(db), 0)

	outState := NewOutState(statedb)

	for i := 1; i <= 15; i++ {
		uint256s := keys.Uint256{uint8(i)}
		root := outState.AppendLeaf(uint256s)
		fmt.Println("root : ", root)
	}

	for i := 1; i <= 15; i++ {
		fmt.Println(i, ":", outState.db.GetState(EmptyAddress, indexPathKey(uint64(i), uint64(0))))
	}

	fmt.Println("--------------------------------------------------")
	for i := 1; i <= 15; i++ {
		fmt.Println(i, ":", outState.db.GetState(EmptyAddress, indexPathKey(uint64(i), uint64(1))))
	}

	for i := 9; i <= 15; i++ {
		current := keys.Uint256{uint8(i)}
		index, getPaths := outState.GetPaths(current)
		for _, each := range getPaths {
			if index%2 == 0 {
				current = cpt.Combine(&current, each)
			} else {
				current = cpt.Combine(each, &current)
			}
			index = index / 2

		}
		fmt.Println(current)
	}

}
