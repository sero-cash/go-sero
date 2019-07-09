package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/crypto"
	"math/rand"
	"testing"
)

func (node *SNode) Print(state State) {
	if node.key == emptyHash {
		return
	}
	total := state.GetStakeState(node.totalKey())
	num := state.GetStakeState(node.numKey())
	fmt.Printf("%s, %v, %v\n", common.Bytes2Hex(node.key[:]), decodeNumber32(total[28:32]), decodeNumber32(num[28:32]))
}

func (node *SNode) println() {
	fmt.Printf("%s, %v, %v\n", common.Bytes2Hex(node.key[:]), node.total, node.num)
}

func (node *SNode) MiddleOrder(state State) {
	leftHash := state.GetStakeState(node.leftKey())
	if leftHash != emptyHash {
		left := &SNode{key: leftHash}
		left.MiddleOrder(state)
	}
	node.Print(state)
	rightHash := state.GetStakeState(node.rightKey())
	if rightHash != emptyHash {
		right := &SNode{key: rightHash}
		right.MiddleOrder(state)
	}
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

func initTree(state State) (*STree, map[common.Hash]uint32) {
	all := map[common.Hash]uint32{}
	tree := NewTree(state)
	for i := 1; i <= 10; i++ {
		u := uint8(rand.Intn(100))
		hash := crypto.Keccak256Hash([]byte{u, uint8(i)})
		for all[hash] > 0 || u == 0 {
			u = uint8(rand.Intn(100))
			hash = crypto.Keccak256Hash([]byte{u})
		}
		all[hash] = uint32(u)
		num := uint32(u)
		//num := uint32(1)
		tree.insert(&SNode{key: hash, total: num, num: num})
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

	fmt.Println()
	for {
		var key common.Hash
		for hash, _ := range all {
			key = hash
			fmt.Println("delete ", common.Bytes2Hex(hash[:]))

			//snapshot := stateDB.Snapshot()
			//fmt.Println("snapshot : ", snapshot)
			tree.deleteNodeByHash(hash, 1)
			//if index == 2 {
			//	root := stateDB.IntermediateRoot(true)
			//	fmt.Println("root : ", root.String())
			//} else {
			//	stateDB.RevertToSnapshot(snapshot)
			//}

			rootNode := &SNode{key: state.GetStakeState(rootKey)}
			rootNode.MiddleOrder(state)
			fmt.Println()
			break
		}
		all[key] -= 1
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
