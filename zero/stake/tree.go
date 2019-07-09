package stake

import (
	"github.com/sero-cash/go-sero/common"
	"math/big"
)

var (
	emptyHash = common.Hash{}
	rootKey   = common.BytesToHash([]byte("ROOT"))
)

type SNode struct {
	key       common.Hash
	num       uint32
	total     uint32
}

func (node *SNode) init(state State) *SNode {
	total := state.GetStakeState(node.totalKey())
	num := state.GetStakeState(node.numKey())
	node.total = decodeNumber32(total[28:32])
	node.num = decodeNumber32(num[28:32])
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
	hash[30] = 0
	hash[31] = 0
	return hash
}

func (node *SNode) rightKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[30] = 0
	hash[31] = 1
	return hash
}

func (node *SNode) numKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[30] = 1
	hash[31] = 0
	return hash
}

func (node *SNode) totalKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[30] = 1
	hash[31] = 1
	return hash
}

func (node *SNode) stateHashKey() common.Hash {
	return node.key
}

type STree struct {
	root  common.Hash
	state State
}

func NewTree(state State) *STree {
	return &STree{common.Hash{}, state}
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
		tree.state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(node.total), 32)))
		tree.state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(node.num), 32)))
	} else {
		tree.insertNode(&SNode{key: hash}, node)
	}

}

func (tree *STree) insertNode(parent *SNode, children *SNode) {
	for {
		value := tree.state.GetStakeState(parent.totalKey())
		number := decodeNumber32(value[28:32])
		tree.state.SetStakeState(parent.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(number+children.total), 32)))

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
			tree.state.SetStakeState(children.numKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(children.num), 32)))
			tree.state.SetStakeState(children.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(children.total), 32)))
			return
		} else {
			parent = &SNode{key: hash}
			//tree.insertNode(&SNode{key: key}, children)
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
					number := decodeNumber32(value[28:32])
					tree.state.SetStakeState(path.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(number-num), 32)))
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
//
//func (tree *STree) deletNodeByIndex(index uint32) *SNode {
//	if index >= tree.size() {
//		panic("index > size")
//	}
//	key := rootKey
//	node := &SNode{key: tree.state.GetStakeState(key)}
//	node.init(tree.state)
//
//	for {
//		left := node.left(tree.state)
//		if left != nil {
//			if index < left.total {
//				value := tree.state.GetStakeState(node.totalKey())
//				number := decodeNumber32(value[28:32])
//				tree.state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(number-1), 32)))
//
//				key = node.leftKey()
//				node = left
//				continue
//			}
//			index -= left.total
//		}
//
//		if index < node.num {
//			return tree.deleteNode(key, node, 1)
//		}
//		index -= node.num
//
//		right := node.right(tree.state)
//		if right != nil {
//			value := tree.state.GetStakeState(node.totalKey())
//			number := decodeNumber32(value[28:32])
//			tree.state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(number-1), 32)))
//
//			key = node.rightKey()
//			node = right
//			continue
//		}
//		panic("not found by index")
//	}
//
//}

func (tree *STree) deleteNode(key common.Hash, children *SNode, num uint32) *SNode {
	number := children.num - num
	tree.state.SetStakeState(children.numKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(number), 32)))
	tree.state.SetStakeState(children.totalKey(), common.BytesToHash(common.LeftPadBytes(encodeNumber32(children.total-num), 32)))

	if number == 0 {
		right := children.right(tree.state)
		if right != nil {
			tree.state.SetStakeState(key, right.key)
			left := children.left(tree.state)
			if left != nil {
				tree.insertNode(right, left)
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

//func (tree *STree) findNodeByHash(nodeHash common.Hash) *SNode {
//	rootHash := tree.state.GetStakeState(rootKey)
//	if rootHash == nodeHash {
//		node := &SNode{key: rootHash}
//		return node.init(tree.state)
//	} else {
//		parent := &SNode{key: rootHash}
//		for {
//			var key, key common.Hash
//
//			if cmp(nodeHash, parent.key) < 0 {
//				key = parent.leftKey()
//				key = tree.state.GetStakeState(key)
//			} else {
//				key = parent.rightKey()
//				key = tree.state.GetStakeState(key)
//			}
//
//			if key == nodeHash {
//				node := &SNode{key: key}
//				node.init(tree.state)
//				return node
//			}
//
//			if key == emptyHash {
//				panic("not found by key")
//			} else {
//				parent = &SNode{key: key}
//			}
//		}
//	}
//}
