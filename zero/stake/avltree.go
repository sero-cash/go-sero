package stake

import (
	"bytes"
	"errors"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/utils"
)

var (
	errNotExist       = errors.New("index is not existed")
	errTreeNil        = errors.New("tree is null")
	errTreeIndexExist = errors.New("tree index is existed")
)

var (
	rootKey_new = common.BytesToHash([]byte("ROOT"))
)

type AVLTree struct {
	state State
}

func NewAVLTree(state State) *AVLTree {
	return &AVLTree{state}
}

func CopyFromOldV0(state State, old *STree) *AVLTree {
	list := make([]*Node, 0)
	old.midtraverse(old.newRootNode(), func(node *Node) {
		list = append(list, node)
	}, nil)

	tree := NewAVLTree(state)
	return tree
}

func InitAVLTree(state State) {
	oldTree := &STree{state}
	CopyFromOldV1(oldTree)
}

func CopyFromOldV1(old *STree) *AVLTree {
	old.Lasttraverse(old.newRootNode(), func(node *Node) {
		left := node.left(old.state)
		right := node.right(old.state)

		if left != nil && right != nil {
			height := max(int(left.factor), int(right.factor)) + 1
			old.state.SetStakeState(node.factorKey(),
				common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(height)), 32)))
		} else if left != nil {
			old.state.SetStakeState(node.factorKey(),
				common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(left.factor)+1), 32)))
		} else if right != nil {
			old.state.SetStakeState(node.factorKey(),
				common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(right.factor)+1), 32)))
		} else {
			old.state.SetStakeState(node.factorKey(),
				common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(1), 32)))
		}
	})

	tree := NewAVLTree(old.state)
	return tree
}

func Copy(state State, old *AVLTree) *AVLTree {
	tree := NewAVLTree(state)
	node := old.newRootNode()
	state.SetStakeState(rootKey_new, node.key)
	old.midtraverse(node, func(node *Node) {
		node.store(state)
		node.setLeftChild(state, node.left(old.state))
		node.setRightChild(state, node.right(old.state))

		// tree.Insert(&Node{key: node.key, total: node.num, num: node.num, height: 1})
	}, nil)
	return tree
}

func max(data1 int, data2 int) int {
	if data1 > data2 {
		return data1
	}
	return data2
}

func (tree *AVLTree) getHeight(node *Node) int {
	if node == nil {
		return 0
	}

	data := tree.state.GetStakeState(node.factorKey())
	return int(utils.DecodeNumber32(data[28:32]))
}

func (tree *AVLTree) llRotation(node *Node) *Node {
	prchild := node.right(tree.state)
	node.setRightChild(tree.state, prchild.left(tree.state))
	prchild.setLeftChild(tree.state, node)
	node.setFactor(tree.state, max(node.leftChildFactor(tree.state), node.rightChildFactor(tree.state))+1)
	prchild.setFactor(tree.state, max(prchild.leftChildFactor(tree.state), prchild.rightChildFactor(tree.state))+1)

	total := safeSub(node.total, prchild.total) + node.rightChildTotal(tree.state)
	prchild.setTotal(tree.state, node.total)
	node.setTotal(tree.state, total)

	return prchild
}

func (tree *AVLTree) rrRotation(node *Node) *Node {
	plchild := node.left(tree.state)
	node.setLeftChild(tree.state, plchild.right(tree.state))
	plchild.setRightChild(tree.state, node)

	node.setFactor(tree.state, max(node.leftChildFactor(tree.state), node.rightChildFactor(tree.state))+1)
	plchild.setFactor(tree.state, max(plchild.leftChildFactor(tree.state), plchild.rightChildFactor(tree.state))+1)

	total := safeSub(node.total, plchild.total) + node.leftChildTotal(tree.state)
	plchild.setTotal(tree.state, node.total)
	node.setTotal(tree.state, total)
	return plchild
}

func (tree *AVLTree) lrRotation(node *Node) *Node {
	plchild := tree.llRotation(node.left(tree.state))
	node.setLeftChild(tree.state, plchild)
	return tree.rrRotation(node)
}

func (tree *AVLTree) rlRotation(node *Node) *Node {
	prchild := tree.rrRotation(node.right(tree.state))
	node.setRightChild(tree.state, prchild)
	return tree.llRotation(node)
}

// 处理节点高度问题
func (tree *AVLTree) handleBF(node *Node) *Node {
	leftChild := node.left(tree.state)
	rightChild := node.right(tree.state)

	if tree.getHeight(leftChild)-tree.getHeight(rightChild) == 2 {
		if leftChild.leftChildFactor(tree.state)-leftChild.rightChildFactor(tree.state) > 0 { // RR
			node = tree.rrRotation(node)
		} else {
			node = tree.lrRotation(node)
		}
	} else if tree.getHeight(leftChild)-tree.getHeight(rightChild) == -2 {
		if rightChild.leftChildFactor(tree.state)-rightChild.rightChildFactor(tree.state) < 0 { // LL
			node = tree.llRotation(node)
		} else {
			node = tree.rlRotation(node)
		}
	}
	return node
}

// 插入节点 ---> 依次向上递归，调整树平衡
func (tree *AVLTree) Insert(node *Node) {
	hash := tree.state.GetStakeState(rootKey_new)
	rootNode := tree.insertNode(tree.newRootNode(), node)
	if rootNode.key != hash {
		tree.state.SetStakeState(rootKey_new, rootNode.key)
	}
}

func (tree *AVLTree) insertNode(parent *Node, node *Node) *Node {
	if parent == nil {
		node.store(tree.state)
		return node
	}

	parent.setTotal(tree.state, parent.total+node.num)
	if cmp(parent.key, node.key) > 0 {
		lchild := tree.insertNode(parent.left(tree.state), node)
		parent.setLeftChild(tree.state, lchild)
		parent = tree.handleBF(parent)
	} else {
		rchild := tree.insertNode(parent.right(tree.state), node)
		parent.setRightChild(tree.state, rchild)
		parent = tree.handleBF(parent)
	}

	parent.setFactor(tree.state, max(parent.leftChildFactor(tree.state), parent.rightChildFactor(tree.state))+1)
	return parent
}

func (tree *AVLTree) Midtraverse()  {
	tree.midtraverse(tree.newRootNode(), func(node *Node) {
		node.Print()
	}, nil)
}

// 中序遍历树，并根据钩子函数处理数据
func (tree *AVLTree) midtraverse(node *Node, handle func(*Node), check func(*Node)) error {
	if node == nil {
		return nil
	} else {
		if check != nil {
			check(node)
		}

		if err := tree.midtraverse(node.left(tree.state), handle, check); err != nil { // 处理左子树
			return err
		}
		handle(node)
		if err := tree.midtraverse(node.right(tree.state), handle, check); err != nil { // 处理右子树
			return err
		}
	}
	return nil
}

func (tree *AVLTree) GetNode(nodeHash common.Hash) *Node {
	// hash := tree.state.GetStakeState(rootKey_new)
	// if nodeHash == emptyHash {
	// 	return nil
	// }
	node := &Node{key: nodeHash, pkey: rootKey_new}
	node.load(tree.state)
	return node
}

func (tree *AVLTree) newRootNode() *Node {
	hash := tree.state.GetStakeState(rootKey_new)
	if hash == emptyHash {
		return nil
	}
	node := &Node{key: hash, pkey: rootKey_new}
	node.load(tree.state)
	return node
}

// 查找并返回节点
func (tree *AVLTree) FindByIndex(index uint32) (*Node, error) {
	node := tree.newRootNode()

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

func (tree *AVLTree) Delete(key common.Hash, num uint32) *Node {
	rootNode := tree.newRootNode()
	node := tree.delete(rootNode, key, num)
	if node == nil {
		tree.state.SetStakeState(rootKey_new, emptyHash)
	} else {
		if rootNode.key != node.key {
			tree.state.SetStakeState(rootKey_new, node.key)
		}
	}
	return node
}

func (tree *AVLTree) delete(node *Node, key common.Hash, num uint32) *Node {
	if node == nil {
		return nil
	}

	lchild := node.left(tree.state)
	rchild := node.right(tree.state)
	if node.key == key {
		if node.num == num {
			if lchild == nil && rchild == nil {

				return nil
			} else if lchild == nil || rchild == nil {
				if lchild != nil {
					return lchild
				} else {
					return rchild
				}
			} else {
				n := lchild
				for {
					nright := n.right(tree.state)
					if nright == nil {
						break
					}
					n = nright
				}

				if n.key != lchild.key {
					tn := lchild
					for {
						tn.setTotal(tree.state, safeSub(tn.total+node.num, n.num))
						tn = tn.right(tree.state)
						if tn == nil || tn.key == n.key {
							break
						}
					}
				}

				ncpoy := n.copy()
				nodeCopy := node.copy()

				ncpoylchild := ncpoy.left(tree.state)
				ncpoyrchild := ncpoy.right(tree.state)
				nodeCopylchild := nodeCopy.left(tree.state)
				nodeCopyrchild := nodeCopy.right(tree.state)

				if bytes.Compare(ncpoy.pkey[0:29], nodeCopy.key[0:29]) == 0 {
					node.setNode(tree.state, ncpoy, node.pkey, nodeCopy, nodeCopyrchild)
					n.setNode(tree.state, nodeCopy, ncpoy.key, ncpoylchild, ncpoyrchild)
				} else {
					node.setNode(tree.state, ncpoy, node.pkey, nodeCopylchild, nodeCopyrchild)
					n.setNode(tree.state, nodeCopy, ncpoy.pkey, ncpoylchild, ncpoyrchild)
				}

				lchild = tree.delete(node.left(tree.state), key, num)
				node.setLeftChild(tree.state, lchild)
			}
		} else {
			node.setNum(tree.state, safeSub(node.num, num))
		}
	} else if cmp(node.key, key) > 0 {
		lchild = tree.delete(lchild, key, num)
		node.setLeftChild(tree.state, lchild)
	} else {
		rchild = tree.delete(rchild, key, num)
		node.setRightChild(tree.state, rchild)
	}

	node.setFactor(tree.state, max(tree.getHeight(lchild), tree.getHeight(rchild))+1)
	node.setTotal(tree.state, safeSub(node.total, num))

	node = tree.handleBF(node)
	return node
}

func (tree *AVLTree) Size() uint32 {
	return tree.newRootNode().total
}
