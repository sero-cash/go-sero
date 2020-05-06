package stake

import (
	"fmt"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/utils"
)

type ITree interface {
	Insert(*Node)
	Delete(nodeHash common.Hash, num uint32) *Node
	FindByIndex(index uint32) (*Node, error)

	// GetNode(nodeHash common.Hash) *Node
	Size() uint32
}

func NewTree(state State, blockNumber uint64) ITree {
	if blockNumber >= seroparam.SIP8() {
		return &AVLTree{state}
	} else {
		return &STree{state}
	}
}

type Node struct {
	pkey   common.Hash
	key    common.Hash
	num    uint32
	total  uint32
	factor int
}

func (node *Node) Print() {
	padn := "|"
	// for i := 2; i <= node.height; i++ {
	// 	padn += "   "
	// }
	if node == nil || node.key == emptyHash {
		return
	}
	fmt.Printf("%02v, %s%s, %v, %v\n", node.factor, padn, node.key.String(), node.total, node.num)
}

func (node *Node) copy() *Node {
	return &Node{node.pkey, node.key, node.num, node.total, node.factor}
}

func (node *Node) del(state State) {
	state.GetStakeState(node.totalKey())
}

func (node *Node) load(state State) *Node {
	total := state.GetStakeState(node.totalKey())
	num := state.GetStakeState(node.numKey())
	factor := state.GetStakeState(node.factorKey())
	node.total = utils.DecodeNumber32(total[28:32])
	node.num = utils.DecodeNumber32(num[28:32])
	node.factor = int(utils.DecodeNumber32(factor[28:32]))
	return node
}

func (node *Node) store(state State) {
	state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.total)), 32)))
	state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.num)), 32)))
	state.SetStakeState(node.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.factor)), 32)))
	state.SetStakeState(node.leftKey(), emptyHash)
	state.SetStakeState(node.rightKey(), emptyHash)
}

func (node *Node) setNode(state State, valNode *Node, pkey common.Hash, leftChild, rightChild *Node) {
	node.key = valNode.key
	node.num = valNode.num

	node.pkey = pkey
	state.SetStakeState(pkey, node.key)
	state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.num)), 32)))
	state.SetStakeState(node.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.factor)), 32)))
	state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.total)), 32)))

	if leftChild != nil {
		state.SetStakeState(node.leftKey(), leftChild.key)
	} else {
		state.SetStakeState(node.leftKey(), emptyHash)
	}

	if rightChild != nil {
		state.SetStakeState(node.rightKey(), rightChild.key)
	} else {
		state.SetStakeState(node.rightKey(), emptyHash)
	}
}

func (node *Node) setLeftChild(state State, left *Node) {
	if left != nil {
		state.SetStakeState(node.leftKey(), left.key)
	} else {
		state.SetStakeState(node.leftKey(), emptyHash)
	}
}

func (node *Node) setRightChild(state State, right *Node) {
	if right != nil {
		state.SetStakeState(node.rightKey(), right.key)
	} else {
		state.SetStakeState(node.rightKey(), emptyHash)
	}
}

func (node *Node) setTotal(state State, val uint32) {
	node.total = val
	state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(val)), 32)))
}

func (node *Node) setFactor(state State, val int) {
	if node.factor != val {
		node.factor = val
		state.SetStakeState(node.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(val)), 32)))
	}
}

func (node *Node) setNum(state State, val uint32) {
	node.num = val
	state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(val)), 32)))
}

func (node *Node) left(state State) *Node {
	path := node.leftKey()
	hash := state.GetStakeState(path)
	if hash == emptyHash {
		return nil
	} else {
		left := &Node{key: hash, pkey: path}
		return left.load(state)
	}
}

func (node *Node) leftChildKey(state State) common.Hash {
	leftChild := node.left(state)
	if leftChild != nil {
		return leftChild.key
	} else {
		return emptyHash
	}
}

func (node *Node) leftChildTotal(state State) uint32 {
	leftChild := node.left(state)
	if leftChild != nil {
		return leftChild.total
	} else {
		return 0
	}
}

func (node *Node) leftChildFactor(state State) int {
	leftChild := node.left(state)
	if leftChild != nil {
		return leftChild.factor
	} else {
		return 0
	}
}

func (node *Node) right(state State) *Node {
	path := node.rightKey()
	hash := state.GetStakeState(path)
	if hash == emptyHash {
		return nil
	} else {
		right := &Node{key: hash, pkey: path}
		return right.load(state)
	}
}

func (node *Node) rightChildKey(state State) common.Hash {
	rightChild := node.right(state)
	if rightChild != nil {
		return rightChild.key
	} else {
		return emptyHash
	}
}

func (node *Node) rightChildFactor(state State) int {
	rightChild := node.right(state)
	if rightChild != nil {
		return rightChild.factor
	} else {
		return 0
	}
}

func (node *Node) rightChildTotal(state State) uint32 {
	rightChild := node.right(state)
	if rightChild != nil {
		return rightChild.total
	} else {
		return 0
	}
}

func (node *Node) leftKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 0
	hash[31] = 0
	return hash
}

func (node *Node) rightKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 0
	hash[31] = 1
	return hash
}

func (node *Node) numKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 1
	hash[31] = 0
	return hash
}

func (node *Node) totalKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 1
	hash[31] = 1
	return hash
}

func (node *Node) factorKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 1
	hash[30] = 0
	hash[31] = 0
	return hash
}

func safeSub(a, b uint32) uint32 {
	if a < b {
		panic("")
	}
	return a - b
}
