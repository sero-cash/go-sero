package stake

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/utils"
)

// func newState() (State, *state.StateDB) {
//	db := serodb.NewMemDatabase()
//	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
//	//state.GetZState()
//	//return state
//	//db := consensus.NewFakeDB()
//	return NewStakeState(state), state
// }
func TestTree(t *testing.T) {
	db, stateDB := newState()
	root := stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
	tree, _ := initTree(db, 1000)
	root = stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
	fmt.Println(tree.Size())
}

func initNode(seed uint64, num uint32, all map[common.Hash]uint32) *Node {
	hash := crypto.Keccak256Hash(utils.EncodeNumber(seed))
	// for all[hash] > 0 {
	//	u = uint64(rand.Intn(1000000))
	//	hash = crypto.Keccak256Hash(append(utils.EncodeNumber(u), uint8(i)))
	// }

	all[hash] = num

	return &Node{key: hash, total: num, num: num, factor: 1}
}

func initTree(state State, n int) (*STree, map[common.Hash]uint32) {
	all := map[common.Hash]uint32{}
	tree := &STree{state: state}
	for i := 1; i <= n; i++ {

		u := uint64(rand.Intn(100))
		num := uint32(u%10 + 1)
		tree.Insert(initNode(uint64(i), num, all))
	}
	// hash := state.GetStakeState(rootKey)
	// root := &Node{key: hash}
	// root.load(tree.state)
	// tree.midtraverse(root, func(node *Node) {
	// 	node.Print()
	// }, nil)
	return tree, all
}

func TestTreeFindByIndex(t *testing.T) {
	db, _ := newState()
	tree, _ := initTree(db, 10)

	fmt.Println()
	node, _ := tree.FindByIndex(0)
	node.Print()
}

func TestDelByIndex(t *testing.T) {

	state, stateDB := newState()
	tree, _ := initTree(state, 10)

	fmt.Println()
	for {
		size := tree.Size()

		if size == 0 {
			break
		}

		// snapshot := stateDB.Snapshot()
		fmt.Println("size : ", size)
		index := rand.Uint32() % size
		node, err := tree.FindByIndex(index)
		if err != nil {
			panic(err)
		}
		node = tree.Delete(node.key, 1)
		node.Print()

		// stateDB.RevertToSnapshot(snapshot)

		tree.midtraverse(tree.newRootNode(), func(node *Node) {
			node.Print()
		}, nil)

		root := stateDB.IntermediateRoot(true)
		fmt.Println("root:", root.String())
		fmt.Println()
	}

	tree.midtraverse(tree.newRootNode(), func(node *Node) {
		node.Print()
	}, nil)

	root := stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
}

func TestDelByHash(t *testing.T) {
	state, _ := newState()
	tree, all := initTree(state, 1000)

	start := uint64(100000 + 1)
	fmt.Println()
	for {
		var key common.Hash
		var num uint32
		for key, num = range all {
			if num == 0 {
				log.Error("error", num)
				return
			}
			fmt.Println("delete ", common.Bytes2Hex(key[:]))

			if num >= 5 {
				num = 5
			} else {
				num = 1
			}
			// snapshot := stateDB.Snapshot()
			// fmt.Println("snapshot : ", snapshot)
			tree.Delete(key, num)

			if num != 1 {
				start++
				tree.Insert(initNode(start, 1, all))
			}

			// if index == 2 {
			//	root := stateDB.IntermediateRoot(true)
			//	fmt.Println("root : ", root.String())
			// } else {
			//	stateDB.RevertToSnapshot(snapshot)
			// }

			rootNode := &Node{key: state.GetStakeState(rootKey)}
			rootNode.load(state)
			rootNode.Print()
			// rootNode.MiddleOrder(state)
			fmt.Println()
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

	hash := state.GetStakeState(rootKey)
	root := &Node{key: hash}
	fmt.Println("rootNode", common.Bytes2Hex(hash[:]))

	tree.midtraverse(root, func(node *Node) {
		node.Print()
	}, nil)
	fmt.Println()
}
