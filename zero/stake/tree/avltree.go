package tree

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/stake"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/big"
)

var (
	errNotExist       = errors.New("index is not existed")
	errTreeNil        = errors.New("tree is null")
	errTreeIndexExist = errors.New("tree index is existed")
)

var (
	emptyHash = common.Hash{}
	rootKey   = common.BytesToHash([]byte("ROOT_NEW"))
)

type Node struct {
	pkey   common.Hash
	key    common.Hash
	num    int
	total  int
	height int
}

func (node *Node) Print() {
	if node == nil || node.key == emptyHash {
		return
	}
	fmt.Printf("%v, %s, %v, %v\n", node.height, node.key.String(), node.total, node.num)
}

func (node *Node) copy() *Node {
	return &Node{node.pkey, node.key, node.num, node.total, node.height}
}
func (node *Node) load(state stake.State) *Node {
	total := state.GetStakeState(node.totalKey())
	num := state.GetStakeState(node.numKey())
	height := state.GetStakeState(node.heightKey())
	node.total = int(utils.DecodeNumber32(total[28:32]))
	node.num = int(utils.DecodeNumber32(num[28:32]))
	node.height = int(utils.DecodeNumber32(height[28:32]))
	return node
}

func (node *Node) store(state stake.State) {
	state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.total)), 32)))
	state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.num)), 32)))
	state.SetStakeState(node.heightKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.height)), 32)))
}

func (node *Node) setNode(state stake.State, valNode *Node, pkey common.Hash, leftChild, rightChild *Node) {
	node.key = valNode.key
	node.num = valNode.num

	node.pkey = pkey
	state.SetStakeState(pkey, node.key)
	state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.num)), 32)))
	state.SetStakeState(node.heightKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(node.height)), 32)))

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

func (node *Node) setLeftChild(state stake.State, left *Node) {
	if left != nil {
		state.SetStakeState(node.leftKey(), left.key)
	} else {
		state.SetStakeState(node.leftKey(), emptyHash)
	}
}

func (node *Node) setRightChild(state stake.State, right *Node) {
	if right != nil {
		state.SetStakeState(node.rightKey(), right.key)
	} else {
		state.SetStakeState(node.rightKey(), emptyHash)
	}
}

func (node *Node) setTotal(state stake.State, val int) {
	node.total = val
	state.SetStakeState(node.totalKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(val)), 32)))
}

func (node *Node) setHeight(state stake.State, val int) {
	if node.height != val {
		node.height = val
		state.SetStakeState(node.heightKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(val)), 32)))
	}
}

func (node *Node) setNum(state stake.State, val int) {
	node.num = val
	state.SetStakeState(node.numKey(), common.BytesToHash(common.LeftPadBytes(utils.EncodeNumber32(uint32(val)), 32)))
}

func (node *Node) left(state stake.State) *Node {
	path := node.leftKey()
	hash := state.GetStakeState(path)
	if hash == emptyHash {
		return nil
	} else {
		left := &Node{key: hash, pkey: path}
		return left.load(state)
	}
}

func (node *Node) leftChildKey(state stake.State) common.Hash {
	leftChild := node.left(state)
	if leftChild != nil {
		return leftChild.key
	} else {
		return emptyHash
	}
}

func (node *Node) leftChildTotal(state stake.State) int {
	leftChild := node.left(state)
	if leftChild != nil {
		return leftChild.total
	} else {
		return 0
	}
}

func (node *Node) leftChildHeight(state stake.State) int {
	leftChild := node.left(state)
	if leftChild != nil {
		return leftChild.height
	} else {
		return 0
	}
}

func (node *Node) right(state stake.State) *Node {
	path := node.rightKey()
	hash := state.GetStakeState(path)
	if hash == emptyHash {
		return nil
	} else {
		right := &Node{key: hash, pkey: path}
		return right.load(state)
	}
}

func (node *Node) rightChildKey(state stake.State) common.Hash {
	rightChild := node.right(state)
	if rightChild != nil {
		return rightChild.key
	} else {
		return emptyHash
	}
}

func (node *Node) rightChildHeight(state stake.State) int {
	rightChild := node.right(state)
	if rightChild != nil {
		return rightChild.height
	} else {
		return 0
	}
}

func (node *Node) rightChildTotal(state stake.State) int {
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

func (node *Node) heightKey() common.Hash {
	hash := common.BytesToHash(node.key[:])
	hash[29] = 1
	hash[30] = 0
	hash[31] = 0
	return hash
}

type AVLTree struct {
	state stake.State
}

func NewTree(state stake.State) *AVLTree {
	return &AVLTree{state}
}

func cmp(hash0, hash1 common.Hash) int {
	return new(big.Int).SetBytes(hash0[0:32]).Cmp(new(big.Int).SetBytes(hash1[0:32]))
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

	data := tree.state.GetStakeState(node.heightKey())
	return int(utils.DecodeNumber32(data[28:32]))
}

func (tree *AVLTree) llRotation(node *Node) *Node {
	prchild := node.right(tree.state)
	node.setRightChild(tree.state, prchild.left(tree.state))
	prchild.setLeftChild(tree.state, node)
	node.setHeight(tree.state, max(node.leftChildHeight(tree.state), node.rightChildHeight(tree.state))+1)
	prchild.setHeight(tree.state, max(prchild.leftChildHeight(tree.state), prchild.rightChildHeight(tree.state))+1)

	total := safeSub(node.total, prchild.total) + node.rightChildTotal(tree.state)
	prchild.setTotal(tree.state, node.total)
	node.setTotal(tree.state, total)

	return prchild
}

func (tree *AVLTree) rrRotation(node *Node) *Node {
	plchild := node.left(tree.state)
	node.setLeftChild(tree.state, plchild.right(tree.state))
	plchild.setRightChild(tree.state, node)

	node.setHeight(tree.state, max(node.leftChildHeight(tree.state), node.rightChildHeight(tree.state))+1)
	plchild.setHeight(tree.state, max(plchild.leftChildHeight(tree.state), plchild.rightChildHeight(tree.state))+1)

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
		if leftChild.leftChildHeight(tree.state)-leftChild.rightChildHeight(tree.state) > 0 { // RR
			node = tree.rrRotation(node)
		} else {
			node = tree.lrRotation(node)
		}
	} else if tree.getHeight(leftChild)-tree.getHeight(rightChild) == -2 {
		if rightChild.leftChildHeight(tree.state)-rightChild.rightChildHeight(tree.state) < 0 { // LL
			node = tree.llRotation(node)
		} else {
			node = tree.rlRotation(node)
		}
	}
	return node
}

// 插入节点 ---> 依次向上递归，调整树平衡
func (tree *AVLTree) Insert(node *Node) {
	hash := tree.state.GetStakeState(rootKey)
	var rootNode *Node
	if hash != emptyHash {
		rootNode = tree.newRootNode()
	}

	rootNode = tree.insert(rootNode, node)
	if rootNode.key != hash {
		tree.state.SetStakeState(rootKey, rootNode.key)
	}
}

func (tree *AVLTree) insert(parent *Node, node *Node) *Node {
	if parent == nil {
		node.store(tree.state)
		return node
	}

	parent.setTotal(tree.state, parent.total+node.num)
	if cmp(parent.key, node.key) > 0 {
		lchild := tree.insert(parent.left(tree.state), node)
		parent.setLeftChild(tree.state, lchild)
		parent = tree.handleBF(parent)
	} else {
		rchild := tree.insert(parent.right(tree.state), node)
		parent.setRightChild(tree.state, rchild)
		parent = tree.handleBF(parent)
	}
	parent.setHeight(tree.state, max(parent.leftChildHeight(tree.state), parent.rightChildHeight(tree.state))+1)
	return parent
}

// 中序遍历树，并根据钩子函数处理数据
func (tree *AVLTree) Midtraverse(node *Node) error {
	if node == nil {
		return nil
	} else {
		if err := tree.Midtraverse(node.left(tree.state)); err != nil { // 处理左子树
			return err
		}
		node.load(tree.state).Print()
		if err := tree.Midtraverse(node.right(tree.state)); err != nil { // 处理右子树
			return err
		}
	}
	return nil
}

func (tree *AVLTree) newRootNode() *Node {
	node := &Node{key: tree.state.GetStakeState(rootKey), pkey: rootKey}
	node.load(tree.state)
	return node
}

// 查找并返回节点
func (tree *AVLTree) Search(index int) (*Node, error) {
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

func (tree *AVLTree) Delete(key common.Hash, num int) (*Node, bool, error) {
	rootNode := tree.newRootNode()
	node, b, err := tree.delete(rootNode, key, num)
	if node == nil {
		tree.state.SetStakeState(rootKey, emptyHash)
	} else {
		if rootNode.key != node.key {
			tree.state.SetStakeState(rootKey, node.key)
		}
	}
	return node, b, err
}

func (tree *AVLTree) delete(node *Node, key common.Hash, num int) (*Node, bool, error) {
	if node == nil {
		return nil, true, nil
	}

	flag := true
	lchild := node.left(tree.state)
	rchild := node.right(tree.state)
	if node.key == key {
		if node.num == num {
			if lchild == nil && rchild == nil {
				// tree.state.SetStakeState(node.pkey, emptyHash)
				return nil, true, nil
			} else if lchild == nil || rchild == nil {
				if lchild != nil {
					// tree.state.SetStakeState(node.pkey, lchild.key)
					return lchild, true, nil
				} else {
					// tree.state.SetStakeState(node.pkey, rchild.key)
					return rchild, true, nil
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

				lchild, _, _ = tree.delete(node.left(tree.state), key, 1)
				node.setLeftChild(tree.state, lchild)
			}
		} else {
			node.setNum(tree.state, safeSub(node.num, num))
		}
	} else if cmp(node.key, key) > 0 {
		lchild, _, _ = tree.delete(lchild, key, num)
		node.setLeftChild(tree.state, lchild)
	} else {
		rchild, _, _ = tree.delete(rchild, key, num)
		node.setRightChild(tree.state, rchild)
	}

	node.setHeight(tree.state, max(tree.getHeight(lchild), tree.getHeight(rchild))+1)
	node.setTotal(tree.state, safeSub(node.total, num))

	node = tree.handleBF(node)
	return node, flag, nil
}

func safeSub(a, b int) int {
	if a < b {
		panic("")
	}
	return a - b
}

func (tree *AVLTree) size() int {
	return tree.newRootNode().total
}
