package tree

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/stake"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/rand"
	"testing"
)

func newState() (*stake.StakeState, *state.StateDB) {
	db := serodb.NewMemDatabase()
	state, _ := state.New(state.NewDatabase(db), nil)
	// state.GetZState()
	// return state
	// db := consensus.NewFakeDB()
	return stake.NewStakeState(state), state
}

// func newState() (State, *state.StateDB) {
//	db := serodb.NewMemDatabase()
//	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
//	//state.GetZState()
//	//return state
//	//db := consensus.NewFakeDB()
//	return NewStakeState(state), state
// }

func initTree(state stake.State, n int) (*AVLTree, map[common.Hash]int) {
	all := map[common.Hash]int{}
	tree := NewTree(state)
	for i := 1; i <= n; i++ {

		u := uint64(rand.Intn(1000))
		num := int(u%10 + 1)
		tree.Insert(initNode(uint64(i), num, all))
	}

	hash := state.GetStakeState(rootKey)

	fmt.Println("roothash", hash.String())
	root := &Node{key: hash}
	tree.Midtraverse(root)

	return tree, all
}

func initNode(seed uint64, num int, all map[common.Hash]int) *Node {
	hash := crypto.Keccak256Hash(utils.EncodeNumber(seed))
	// for all[hash] > 0 {
	//	u = uint64(rand.Intn(1000000))
	//	hash = crypto.Keccak256Hash(append(utils.EncodeNumber(u), uint8(i)))
	// }

	all[hash] = num

	return &Node{key: hash, total: num, num: num, height: 1}
}

func TestTree(t *testing.T) {
	db, stateDB := newState()
	root := stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
	tree, _ := initTree(db, 100)
	root = stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
	fmt.Println(tree.size())
}

func TestTreeFindByIndex(t *testing.T) {
	db, _ := newState()
	tree, _ := initTree(db, 10)

	fmt.Println()
	node, _ := tree.Search(53)
	node.Print()
}

func TestDelByIndex(t *testing.T) {

	db, _ := newState()
	tree, _ := initTree(db, 100000)

	count := 0
	all := tree.newRootNode().total
	fmt.Println()
	for {
		rootNode := tree.newRootNode()
		size := rootNode.total
		if size == 0 || count >= all {
			fmt.Println("rootKey", rootNode.key.String(), 0)
			break
		}
		count++
		if size == 43 {
			fmt.Println("12345")
		}
		index := rand.Int() % size
		// fmt.Println("rootKey", rootNode.key.String())
		// fmt.Println("size :", size, "index :", index)

		node, err := tree.Search(index)
		if err != nil {
			panic(err)
		}
		tree.Delete(node.key, 1)
		// tree.Midtraverse(tree.newRootNode())

		// root := stateDB.IntermediateRoot(true)
		// fmt.Println("root:", root.String())
		// fmt.Println()
	}
	if all != count {
		t.Error("error")
	}
}

func TestDelByHash(t *testing.T) {
	db, _ := newState()
	tree, all := initTree(db, 100000)
	root := tree.newRootNode()
	left := root.left(tree.state)
	right := root.right(tree.state)
	fmt.Println(root.key.String(), root.height, root.total, root.num)
	fmt.Println(left.key.String(), left.height, left.total, left.num)
	fmt.Println(right.key.String(), right.height, right.total, right.num)
	fmt.Println()
	for {
		var key common.Hash
		var num int
		for key, num = range all {
			if num == 0 {
				t.Error("error", num)
				return
			}

			if num >= 5 {
				num = 3
			} else {
				num = 1
			}

			// fmt.Println("delete ", common.Bytes2Hex(key[:]), num)
			// snapshot := stateDB.Snapshot()
			// fmt.Println("snapshot : ", snapshot)
			tree.Delete(key, num)

			// rootNode := tree.newRootNode()
			// fmt.Println("rootNode", rootNode.key)
			// tree.Midtraverse(rootNode)

			// start++
			// num = num
			// tree.Insert(initNode(start, 3, all))

			// fmt.Println()
			break
		}
		all[key] -= num
		if all[key] == 0 {
			delete(all, key)
		}
		if len(all) == 0 {
			break
		}

	}

	rootNode := tree.newRootNode()

	fmt.Println("rootNode", rootNode.key)
	tree.Midtraverse(rootNode)
	fmt.Println()
}
