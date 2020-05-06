package stake

import (
	"errors"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/big"
	"strings"

	"github.com/sero-cash/go-sero/common"
)

var (
	emptyHash = common.Hash{}
	rootKey   = common.BytesToHash([]byte("ROOT"))
)

type STree struct {
	state State
}

// func newOldTree(state State) *STree {
// 	return &STree{common.Hash{}, state}
// }

func (tree *STree) newRootNode() *Node {
	rootNode := &Node{key: tree.state.GetStakeState(rootKey)}
	rootNode.load(tree.state)
	return rootNode
}

func (tree *STree) Midtraverse() {
	tree.midtraverse(tree.newRootNode(), func(node *Node) {
		node.Print()
	}, nil)
}

func (tree *STree) midtraverse(node *Node, handle func(*Node), check func(*Node)) {
	if node == nil {
		return
	}
	tree.midtraverse(node.left(tree.state), handle, check)
	handle(node)
	tree.midtraverse(node.right(tree.state), handle, check)
}

func (tree *STree) Lasttraverse(node *Node, handle func(*Node)) {
	if node == nil {
		return
	}
	tree.Lasttraverse(node.left(tree.state), handle)
	tree.Lasttraverse(node.right(tree.state), handle)
	handle(node)
}

func (tree *STree) Size() uint32 {
	parentHash := tree.state.GetStakeState(rootKey)
	parent := &Node{key: parentHash}
	return parent.load(tree.state).total
}

func cmp(hash0, hash1 common.Hash) int {
	return new(big.Int).SetBytes(hash0[0:32]).Cmp(new(big.Int).SetBytes(hash1[0:32]))
}

func cmp1(hash0, hash1 string) int {
	return strings.Compare(hash0, hash1)
}

func (tree *STree) Insert(node *Node) {
	hash := tree.state.GetStakeState(rootKey)
	if hash == emptyHash {
		tree.state.SetStakeState(rootKey, node.key)
		tree.state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(node.total), 32)))
		tree.state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(node.num), 32)))
		tree.state.SetStakeState(node.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.factor)), 32)))
	} else {
		tree.insertNode(&Node{key: hash}, node)
	}

}

func (tree *STree) insertNode(parent *Node, children *Node) {
	for {
		value := tree.state.GetStakeState(parent.totalKey())
		totalNum := utils.DecodeNumber32(value[28:32])
		tree.state.SetStakeState(parent.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(totalNum+children.total), 32)))
		value = tree.state.GetStakeState(parent.factorKey())
		nodeNum := utils.DecodeNumber32(value[28:32])
		tree.state.SetStakeState(parent.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(nodeNum+uint32(children.factor)), 32)))

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
			tree.state.SetStakeState(children.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(children.factor)), 32)))
			return
		} else {
			parent = &Node{key: hash}
		}
	}
}

func (tree *STree) Delete(nodeHash common.Hash, num uint32) *Node {
	rootHash := tree.state.GetStakeState(rootKey)
	if rootHash == nodeHash {
		node := &Node{key: rootHash}
		node.load(tree.state)
		tree.deleteNode(rootKey, node, num)
		return node
	} else {
		paths := []*Node{}
		parent := &Node{key: rootHash}
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
				node := &Node{key: nodeHash}
				node.load(tree.state)

				for _, path := range paths {
					value := tree.state.GetStakeState(path.totalKey())
					number := utils.DecodeNumber32(value[28:32])
					tree.state.SetStakeState(path.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(safeSub(number, num)), 32)))
					if num == node.num {
						value = tree.state.GetStakeState(path.factorKey())
						nodeNum := utils.DecodeNumber32(value[28:32])
						tree.state.SetStakeState(path.factorKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(safeSub(nodeNum, 1)), 32)))
					}
				}
				tree.deleteNode(key, node, num)
				return node
			}

			if hash == emptyHash {
				return nil
			} else {
				parent = &Node{key: hash}
			}
		}
	}
}

func (tree *STree) deleteNode(key common.Hash, children *Node, num uint32) *Node {
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
				if right.factor > left.factor {
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

func (tree *STree) FindByIndex(index uint32) (*Node, error) {
	root := tree.state.GetStakeState(rootKey)
	node := &Node{key: root}
	node.load(tree.state)

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
			return node, nil
		}
		index -= node.num

		right := node.right(tree.state)
		if right != nil {
			node = right
			continue
		}
		return nil, errors.New("not found node by index")
	}
}
