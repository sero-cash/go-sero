package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/rand"
	"testing"
	"time"
)

var newprint = func(node *Node) {
	node.Print()
}

// func newState() (State, *state.StateDB) {
//	db := serodb.NewMemDatabase()
//	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
//	//state.GetZState()
//	//return state
//	//db := consensus.NewFakeDB()
//	return NewStakeState(state), state
// }

func initAVLTree(state State, n int) (*AVLTree, map[common.Hash]uint32) {
	all := map[common.Hash]uint32{}
	tree := NewAVLTree(state)
	for i := 1; i <= n; i++ {

		u := uint64(rand.Intn(1000))
		num := uint32(u%10 + 1)
		tree.Insert(initAVLNode(uint64(i), num, all))
	}

	return tree, all
}

func initAVLNode(seed uint64, num uint32, all map[common.Hash]uint32) *Node {
	hash := crypto.Keccak256Hash(utils.EncodeNumber(seed))
	// for all[hash] > 0 {
	//	u = uint64(rand.Intn(1000000))
	//	hash = crypto.Keccak256Hash(append(utils.EncodeNumber(u), uint8(i)))
	// }
	if _, ok := all[hash]; ok {
		panic("initAVLNode")
	}
	all[hash] = num

	return &Node{key: hash, total: num, num: num, factor: 1}
}

func TestAVLTree(t *testing.T) {
	db, _ := newState()
	tree, all := initAVLTree(db, 10)
	tree.Midtraverse()
	fmt.Println(tree.Size(), len(all), tree.newRootNode().factor)
}

func TestAVLTreeCopy(t *testing.T) {
	db1, _ := newState()
	tree, all := initAVLTree(db1, 20)
	tree.Midtraverse()
	fmt.Println(tree.Size(), len(all), tree.newRootNode().factor)
	db2, _ := newState()
	newTree := Copy(db2, tree)
	newTree.Midtraverse()
	fmt.Println(newTree.Size(), newTree.newRootNode().factor)
}

func TestAVLTreeFindByIndex(t *testing.T) {
	db, _ := newState()
	tree, _ := initAVLTree(db, 10)
	tree.Midtraverse()
	fmt.Println()
	node, _ := tree.FindByIndex(0)
	node.Print()
}

func TestAVLDelByIndex(t *testing.T) {

	db, _ := newState()
	tree, _ := initAVLTree(db, 100)

	count := uint32(0)
	all := tree.newRootNode().total
	fmt.Println()
	for {

		if tree.Size() == 0 || count >= all {
			break
		}
		count++
		index := rand.Uint32() % tree.Size()
		node, err := tree.FindByIndex(index)
		if err != nil {
			panic(err)
		}
		ret := tree.Delete(node.key, 1)
		if ret == nil || node.num != ret.num {
			panic("delete err")
		}
		tree.Midtraverse()
		fmt.Println()
		// tree.newRootNode().Print()
		fmt.Println("------------------------------------------")
	}
	if all != count {
		t.Error("error")
	}
}

func TestAVLDelByHash(t *testing.T) {
	db, _ := newState()
	tree, all := initAVLTree(db, 1000)
	fmt.Println()
	start := 100000
	for {
		var key common.Hash
		var num uint32
		for key, num = range all {
			if num == 0 {
				t.Error("error", num)
				return
			}

			if num >= 5 {
				num = 5
			} else {
				num = 1
			}

			// ddb, _ := newState()
			// dTree := Copy(ddb, tree)
			fmt.Println("delete ", common.Bytes2Hex(key[:]), num)
			tree.Delete(key, num)
			tree.midtraverse(tree.newRootNode(), func(node *Node) {
				// node.Print()
			}, func(node *Node) {
				if !check(node, node.left(tree.state), node.right(tree.state)) {
					panic("")
					// node.Print()
					// dTree.Delete(key, num)
				}
			})

			start++
			u := uint64(rand.Intn(10))
			node := initAVLNode(uint64(start), uint32(u%10+1), all)
			fmt.Println()
			fmt.Print("insert: ")
			node.Print()

			// idb, _ := newState()
			// iTree := Copy(idb, tree)
			tree.Insert(node)

			tree.midtraverse(tree.newRootNode(), func(node *Node) {
				// node.Print()
			}, func(node *Node) {
				if !check(node, node.left(tree.state), node.right(tree.state)) {
					panic("")
					// node.Print()
					// iTree.Insert(&Node{key: node.key, num: node.num, height: 1})
				}
			})

			fmt.Println("-------------------")
			break
		}
		all[key] -= num
		if all[key] == 0 {
			delete(all, key)
			fmt.Println("remove", key.String())
		}
		if len(all) == 0 {
			break
		}
		rootNode := tree.newRootNode()
		fmt.Println("tree", len(all), rootNode.total, rootNode.factor)
		if len(all) < 2^(rootNode.factor-3) {
			panic("len(all) < 2^(rootNode.height-3)")
		}
		time.Sleep(time.Microsecond * 100)
	}

	rootNode := tree.newRootNode()

	fmt.Println("rootNode", rootNode.key)
	tree.Midtraverse()
	fmt.Println()
}

func TestOldAndNew(t *testing.T) {
	state1, _ := newState()
	oldTree, all := initTree(state1, 10)
	fmt.Println()
	oldTree.newRootNode().Print()

	newTree := CopyFromOldV1(oldTree)
	newTree.Midtraverse()
	fmt.Println()
	newTree.newRootNode().Print()
	fmt.Println()
	for i := 11; i <= 20; i++ {
		u := uint64(rand.Intn(10))
		node := initAVLNode(uint64(i), uint32(u%10+1), all)
		fmt.Print("insert: ")
		node.Print()
		newTree.Insert(node)
	}

	newTree.Midtraverse()
	fmt.Println()
	newTree.newRootNode().Print()
	fmt.Println()
	// idb, _ := newState()
	// iTree := Copy(idb, tree)

}

func check(node, left, right *Node) bool {
	leftH := 0
	rightH := 0
	leftT := uint32(0)
	rightT := uint32(0)

	if left != nil {
		leftH = left.factor
		leftT = left.total
	}
	if right != nil {
		rightH = right.factor
		rightT = right.total
	}
	if node.factor != max(leftH, rightH)+1 {
		return false
	}

	if node.total != leftT+rightT+node.num {
		return false
	}
	return true
}
