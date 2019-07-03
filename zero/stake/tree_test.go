package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/consensus"

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

func newState() State {
	//db := serodb.NewMemDatabase()
	//state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
	//return state
	db := consensus.NewFakeDB()
	return NewStakeState(&db)
}
func TestTree(t *testing.T) {
	tree, _ := initTree(newState())
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

func TestDelByIndex2(t *testing.T) {
	state := newState()
	tree, _ := initTree(state)
	tree.deletNodeByIndex(7)

	fmt.Println()
	root := &SNode{key: state.GetStakeState(rootKey)}
	root.MiddleOrder(state)

}
func TestDelByIndex(t *testing.T) {

	state := newState()
	tree, _ := initTree(state)

	fmt.Println()
	for {
		size := tree.size()

		if size == 0 {
			break
		}

		fmt.Println("size : ", size)
		index := rand.Uint32() % size
		node := tree.deletNodeByIndex(index)
		node.println()

		root := &SNode{key: state.GetStakeState(rootKey)}

		root.MiddleOrder(state)
		fmt.Println()
	}

	hash := state.GetStakeState(rootKey)
	root := &SNode{key: hash}
	fmt.Println("root", common.Bytes2Hex(hash[:]))

	root.MiddleOrder(state)
	fmt.Println()
}

func TestDelByHash(t *testing.T) {
	state := newState()
	tree, all := initTree(state)

	fmt.Println()
	for hash, _ := range all {
		fmt.Println("delet ", common.Bytes2Hex(hash[:]))
		tree.deleteNodeByHash(hash)

		root := &SNode{key: state.GetStakeState(rootKey)}
		root.MiddleOrder(state)
		fmt.Println()
	}
	hash := state.GetStakeState(rootKey)
	root := &SNode{key: hash}
	fmt.Println("root", common.Bytes2Hex(hash[:]))

	root.MiddleOrder(state)
	fmt.Println()
}
