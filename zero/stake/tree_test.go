package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/rand"
	"testing"
)

func (node *SNode) println() {
	fmt.Printf("%s, %v, %v, %v\n", common.Bytes2Hex(node.key[:]), node.total, node.num, node.nodeNum)
}

//func newState() (State, *state.StateDB) {
//	db := serodb.NewMemDatabase()
//	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
//	//state.GetZState()
//	//return state
//	//db := consensus.NewFakeDB()
//	return NewStakeState(state), state
//}
func TestTree(t *testing.T) {
	db, stateDB := newState()
	root := stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
	tree, _ := initTree(db)
	root = stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())
	fmt.Println(tree.size())
}

func initNode(seed uint64, num uint32, all map[common.Hash]uint32) *SNode {
	hash := crypto.Keccak256Hash(utils.EncodeNumber(seed))
	//for all[hash] > 0 {
	//	u = uint64(rand.Intn(1000000))
	//	hash = crypto.Keccak256Hash(append(utils.EncodeNumber(u), uint8(i)))
	//}

	all[hash] = num

	return &SNode{key: hash, total: num, num: num, nodeNum: 1}
}

func initTree(state State) (*STree, map[common.Hash]uint32) {
	all := map[common.Hash]uint32{}
	tree := NewTree(state)
	for i := 1; i <= 1000; i++ {

		u := uint64(rand.Intn(100))
		num := uint32(u%10 + 1)
		tree.insert(initNode(uint64(i), num, all))
	}
	hash := state.GetStakeState(rootKey)
	root := &SNode{key: hash}
	root.MiddleOrder(state)
	return tree, all
}

//func TestTreeFindByIndex(t *testing.T) {
//	db := serodb.NewMemDatabase()
//	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
//	tree, _ := initTree(state)
//
//	fmt.Println()
//	node := tree.findByIndex(492)
//	node.Print(state)
//}
//
//func TestTreeFindByHash(t *testing.T) {
//	db := serodb.NewMemDatabase()
//	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
//	tree, all := initTree(state)
//
//	for key, num := range all {
//		node := tree.findNodeByHash(key)
//		fmt.Println(num)
//		node.Print(tree.state)
//		fmt.Println(node)
//
//	}
//}
//
//func TestDelByIndex(t *testing.T) {
//
//	state, stateDB := newState()
//	tree, _ := initTree(state)
//
//	fmt.Println()
//	for {
//		size := tree.size()
//
//		if size == 0 {
//			break
//		}
//
//		snapshot := stateDB.Snapshot()
//		fmt.Println("size : ", size)
//		index := rand.Uint32() % size
//		node := tree.deletNodeByIndex(index)
//		node.println()
//
//		stateDB.RevertToSnapshot(snapshot)
//
//		rootNode := &SNode{key: state.GetStakeState(rootKey)}
//		rootNode.MiddleOrder(state)
//
//		root := stateDB.IntermediateRoot(true)
//		fmt.Println("root:", root.String())
//		fmt.Println()
//	}
//
//	hash := state.GetStakeState(rootKey)
//	rootNode := &SNode{key: hash}
//	rootNode.MiddleOrder(state)
//
//	root := stateDB.IntermediateRoot(true)
//	fmt.Println("root:", root.String())
//}

func TestDelByHash(t *testing.T) {
	state, _ := newState()
	tree, all := initTree(state)

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
			//snapshot := stateDB.Snapshot()
			//fmt.Println("snapshot : ", snapshot)
			tree.deleteNodeByHash(key, num)

			if num != 1 {
				start++
				tree.insert(initNode(start, 1, all))
			}

			//if index == 2 {
			//	root := stateDB.IntermediateRoot(true)
			//	fmt.Println("root : ", root.String())
			//} else {
			//	stateDB.RevertToSnapshot(snapshot)
			//}

			rootNode := &SNode{key: state.GetStakeState(rootKey)}
			rootNode.init(state)
			rootNode.println()
			//rootNode.MiddleOrder(state)
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
	root := &SNode{key: hash}
	fmt.Println("rootNode", common.Bytes2Hex(hash[:]))

	root.MiddleOrder(state)
	fmt.Println()
}
