package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/big"

	"github.com/sero-cash/go-sero/common"
)

var (
	emptyHash = common.Hash{}
	rootKey   = common.BytesToHash([]byte("ROOT"))
)

type SNode struct {
	key     common.Hash
	num     uint32
	total   uint32
	nodeNum uint32
}

func (node *SNode) init(state State) *SNode {
	total := state.GetStakeState(node.totalKey())
	num := state.GetStakeState(node.numKey())
	nodeNum := state.GetStakeState(node.nodeNumKey())
	node.total = utils.DecodeNumber32(total[28:32])
	node.num = utils.DecodeNumber32(num[28:32])
	node.nodeNum = utils.DecodeNumber32(nodeNum[28:32])
	return node
}

func (node *SNode) left(state State) *SNode {
	hash := state.GetStakeState(node.leftKey())
	if hash == emptyHash {
		return nil
	} else {
		left := &SNode{key: hash}
		return left.init(state)
	}
}

func (node *SNode) right(state State) *SNode {
	hash := state.GetStakeState(node.rightKey())
	if hash == emptyHash {
		return nil
	} else {
		right := &SNode{key: hash}
		return right.init(state)
	}
}

func (node *SNode) leftKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 0
	hash[31] = 0
	return hash
}

func (node *SNode) rightKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 0
	hash[31] = 1
	return hash
}

func (node *SNode) numKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 1
	hash[31] = 0
	return hash
}

func (node *SNode) totalKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 0
	hash[30] = 1
	hash[31] = 1
	return hash
}

func (node *SNode) nodeNumKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 1
	hash[30] = 0
	hash[31] = 0
	return hash
}

type STree struct {
	root  common.Hash
	state State
}

func NewTree(state State) *STree {
	return &STree{common.Hash{}, state}
}
func (node *SNode) Print(state State) {
	if node.key == emptyHash {
		return
	}
	total := state.GetStakeState(node.totalKey())
	num := state.GetStakeState(node.numKey())
	nodeNum := state.GetStakeState(node.nodeNumKey())
	fmt.Printf("%s, %v, %v, %v\n", common.Bytes2Hex(node.key[:]), utils.DecodeNumber32(total[28:32]), utils.DecodeNumber32(num[28:32]), utils.DecodeNumber32(nodeNum[28:32]))
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

func (tree *STree) MiddleOrder() {
	hash := tree.state.GetStakeState(rootKey)
	rootNode := &SNode{key: hash}
	rootNode.MiddleOrder(tree.state)
}
func (tree *STree) size() uint32 {
	parentHash := tree.state.GetStakeState(rootKey)
	parent := &SNode{key: parentHash}
	return parent.init(tree.state).total
}

func cmp(hash0, hash1 common.Hash) int {
	return new(big.Int).SetBytes(hash0[0:32]).Cmp(new(big.Int).SetBytes(hash1[0:32]))
}

func (tree *STree) insert(node *SNode) {
	hash := tree.state.GetStakeState(rootKey)
	if hash == emptyHash {
		tree.state.SetStakeState(rootKey, node.key)
		tree.state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(node.total), 32)))
		tree.state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(node.num), 32)))
		tree.state.SetStakeState(node.nodeNumKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(node.nodeNum), 32)))
	} else {
		tree.insertNode(&SNode{key: hash}, node)
	}

}

func (tree *STree) insertNode(parent *SNode, children *SNode) {
	for {
		value := tree.state.GetStakeState(parent.totalKey())
		totalNum := utils.DecodeNumber32(value[28:32])
		tree.state.SetStakeState(parent.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(totalNum+children.total), 32)))
		value = tree.state.GetStakeState(parent.nodeNumKey())
		nodeNum := utils.DecodeNumber32(value[28:32])
		tree.state.SetStakeState(parent.nodeNumKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(nodeNum+children.nodeNum), 32)))

		var hash, key common.Hash
		if parent.key == children.key {
			return
		}
		if cmp(children.key, parent.key) < 0 {
			key = parent.leftKey()
			hash = tree.state.GetStakeState(key)
		} else {
			key = parent.rightKey()
			hash = tree.state.GetStakeState(key)
		}
		if hash == emptyHash {
			tree.state.SetStakeState(key, children.key)
			tree.state.SetStakeState(children.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(children.num), 32)))
			tree.state.SetStakeState(children.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(children.total), 32)))
			tree.state.SetStakeState(children.nodeNumKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(children.nodeNum), 32)))
			return
		} else {
			parent = &SNode{key: hash}
		}
	}
}

func (tree *STree) deleteNodeByHash(nodeHash common.Hash, num uint32) *SNode {
	rootHash := tree.state.GetStakeState(rootKey)
	if rootHash == nodeHash {
		node := &SNode{key: rootHash}
		node.init(tree.state)
		tree.deleteNode(rootKey, node, num)
		return node
	} else {
		paths := []*SNode{}
		parent := &SNode{key: rootHash}
		for {
			var hash, key common.Hash
			if cmp(nodeHash, parent.key) < 0 {
				key = parent.leftKey()
				hash = tree.state.GetStakeState(key)
			} else {
				key = parent.rightKey()
				hash = tree.state.GetStakeState(key)
			}

			paths = append(paths, parent)

			if hash == nodeHash {
				node := &SNode{key: nodeHash}
				node.init(tree.state)

				for _, path := range paths {
					value := tree.state.GetStakeState(path.totalKey())
					number := utils.DecodeNumber32(value[28:32])
					tree.state.SetStakeState(path.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(number-num), 32)))
					if num == node.num {
						value = tree.state.GetStakeState(path.nodeNumKey())
						nodeNum := utils.DecodeNumber32(value[28:32])
						tree.state.SetStakeState(path.nodeNumKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(nodeNum-1), 32)))
					}
				}
				tree.deleteNode(key, node, num)
				return node
			}

			if hash == emptyHash {
				return nil
			} else {
				parent = &SNode{key: hash}
			}
		}
	}
}

func (tree *STree) deleteNode(key common.Hash, children *SNode, num uint32) *SNode {
	number := children.num - num
	tree.state.SetStakeState(children.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(number), 32)))
	tree.state.SetStakeState(children.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(children.total-num), 32)))

	if number == 0 {
		right := children.right(tree.state)
		if right != nil {
			left := children.left(tree.state)
			if left == nil {
				tree.state.SetStakeState(key, right.key)
			} else {
				if right.nodeNum > left.nodeNum {
					tree.state.SetStakeState(key, right.key)
					tree.insertNode(right, left)
				} else {
					tree.state.SetStakeState(key, left.key)
					tree.insertNode(left, right)
				}
			}
		} else {
			left := children.left(tree.state)
			if left != nil {
				tree.state.SetStakeState(key, left.key)
			} else {
				tree.state.SetStakeState(key, emptyHash)
				return children
			}
		}
	}
	return children
}

func (tree *STree) findByIndex(index uint32) *SNode {
	root := tree.state.GetStakeState(rootKey)
	node := &SNode{key: root}
	node.init(tree.state)

	for {
		left := node.left(tree.state)
		if left != nil {
			if index < left.total {
				node = left
				continue
			}
			index -= left.total
		}

		if index < node.num {
			return node
		}
		index -= node.num

		right := node.right(tree.state)
		if right != nil {
			node = right
			continue
		}
		panic("not found node by index")
	}
}
