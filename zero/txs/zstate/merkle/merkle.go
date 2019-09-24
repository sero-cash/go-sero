package merkle

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"

	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/rlp"
)

type CombinFunc func(*c_type.Uint256, *c_type.Uint256) (out c_type.Uint256)

type Param struct {
	emptyRoots   [c_type.DEPTH + 1]c_type.Uint256
	obj          c_type.PKr
	is_load      bool
	indexLeafKey c_type.Uint256
	indexTreeKey c_type.Uint256
	cap          uint64
	leafcap      uint64
	startIndex   uint64
	combine      CombinFunc
}

func NewParam(obj *c_type.PKr, combine CombinFunc) (ret Param) {
	ret.obj = *obj
	ret.indexLeafKey = c_type.Uint256(crypto.Keccak256Hash([]byte("outTreeIndex_Leaf")))
	ret.indexTreeKey = c_type.Uint256(crypto.Keccak256Hash([]byte("outTreeIndex_Self")))
	ret.cap = toPow2(DEPTH + 1)
	ret.leafcap = toPow2(DEPTH)
	ret.startIndex = toPow2(DEPTH)
	ret.combine = combine
	return
}

func (self *Param) EmptyRoots() []c_type.Uint256 {
	if !self.is_load {
		self.is_load = true
		self.emptyRoots = createEmpty()
	}
	return self.emptyRoots[:]
}

type Leaf struct {
	leafIndex uint64
	treeIndex uint64
	value     c_type.Uint256
}

type MerkleTree struct {
	param *Param
	db    tri.Tri
}

func (self *Param) NewMerkleTree(db tri.Tri) (ret MerkleTree) {
	ret.db = db
	ret.param = self
	return
}

func (self *MerkleTree) AppendLeaf(value c_type.Uint256) c_type.Uint256 {
	leafIndex := self.nextLeafIndex()
	treeIndex := self.geCurrentTreeIndex()
	self.db.SetState(&self.param.obj, leafKey(value).NewRef(), c_type.Uint64_To_Uint256(leafIndex).NewRef())
	self.db.SetState(&self.param.obj, treeKey(value).NewRef(), c_type.Uint64_To_Uint256(treeIndex).NewRef())
	self.db.SetState(&self.param.obj, indexPathKey(leafIndex, treeIndex).NewRef(), &value)

	current_value := value
	depth := toDepth(leafIndex)
	for leafIndex != 1 {
		brotherIndex := brother(leafIndex)
		var brotherValue c_type.Uint256
		if brotherIndex > leafIndex {
			brotherValue = self.param.EmptyRoots()[depth]
		} else {
			brotherValue = self.db.GetState(&self.param.obj, indexPathKey(brotherIndex, treeIndex).NewRef())
			if brotherValue == c_type.Empty_Uint256 {
				panic(fmt.Sprintf("brother value is empty"))
			}
		}

		if leafIndex%2 == 0 {
			current_value = self.param.combine(&current_value, &brotherValue)
		} else {
			current_value = self.param.combine(&brotherValue, &current_value)
		}

		leafIndex = parent(leafIndex)
		depth++
		self.db.SetState(&self.param.obj, indexPathKey(leafIndex, treeIndex).NewRef(), &current_value)
	}
	return current_value
}

func (self *MerkleTree) GetLeafSize() (ret uint64) {
	leafIndex := self.getCurrentLeafIndex() - self.param.startIndex
	tree_count := self.geCurrentTreeIndex()
	return tree_count*self.param.leafcap + leafIndex
}

func (self *MerkleTree) GetPaths(value c_type.Uint256) (pos uint64, paths [DEPTH]c_type.Uint256, anchor c_type.Uint256) {
	leafIndex := c_type.Uint256_To_Uint64(self.db.GetState(&self.param.obj, leafKey(value).NewRef()).NewRef())
	if leafIndex == 0 {
		panic(fmt.Errorf("leaf index can not be 0"))
	}

	treeIndex := c_type.Uint256_To_Uint64(self.db.GetState(&self.param.obj, treeKey(value).NewRef()).NewRef())
	currentTreeIndex := self.geCurrentTreeIndex()

	var cur_leafIndex uint64
	if currentTreeIndex == treeIndex {
		cur_leafIndex = self.getCurrentLeafIndex() - 1
	} else {
		cur_leafIndex = self.param.cap - 1
	}

	anchor = self.db.GetState(&self.param.obj, indexPathKey(1, treeIndex).NewRef())
	pos = leafIndex - self.param.startIndex

	depth := toDepth(leafIndex)
	for leafIndex != 1 {
		brotherIndex := brother(leafIndex)
		var brotherValue c_type.Uint256
		if brotherIndex > cur_leafIndex {
			brotherValue = self.param.EmptyRoots()[depth]
		} else {
			brotherValue = self.db.GetState(&self.param.obj, indexPathKey(brotherIndex, treeIndex).NewRef())
			if brotherValue == c_type.Empty_Uint256 {
				panic(fmt.Sprintf("brother value is empty"))
			}
		}
		paths[depth] = brotherValue
		leafIndex = parent(leafIndex)
		cur_leafIndex = parent(cur_leafIndex)
		depth++
	}
	return
}

func (self *MerkleTree) nextLeafIndex() uint64 {
	leafIndex := self.getCurrentLeafIndex()
	if leafIndex == self.param.cap {
		leafIndex = self.param.startIndex
		treeIndex := self.geCurrentTreeIndex()
		treeIndex = treeIndex + 1
		self.db.SetState(&self.param.obj, &self.param.indexTreeKey, c_type.Uint64_To_Uint256(treeIndex).NewRef())
	}
	self.db.SetState(&self.param.obj, &self.param.indexLeafKey, c_type.Uint64_To_Uint256(leafIndex+1).NewRef())
	return leafIndex
}

func (self *MerkleTree) getCurrentLeafIndex() uint64 {
	value := self.db.GetState(&self.param.obj, &self.param.indexLeafKey)
	index := c_type.Uint256_To_Uint64(&value)
	if index == 0 {
		index = self.param.startIndex
	}
	return index
}

func (self *MerkleTree) geCurrentTreeIndex() uint64 {
	value := self.db.GetState(&self.param.obj, &self.param.indexTreeKey)
	index := c_type.Uint256_To_Uint64(&value)
	return index
}

func parent(index uint64) uint64 {
	return index / 2
}

func brother(index uint64) uint64 {
	if index%2 == 0 {
		return index + 1
	} else {
		return index - 1
	}
}

func leafKey(value c_type.Uint256) c_type.Uint256 {
	return c_type.Uint256(crypto.Keccak256Hash(append([]byte("LEAF_"), value[:]...)))
}

func treeKey(value c_type.Uint256) c_type.Uint256 {
	return c_type.Uint256(crypto.Keccak256Hash(append([]byte("TREE_"), value[:]...)))
}

func indexPathKey(leafIndex, treeIndex uint64) c_type.Uint256 {
	bytes, _ := rlp.EncodeToBytes([]interface{}{leafIndex, treeIndex, []byte("PATH")})
	return c_type.Uint256(crypto.Keccak256Hash(bytes))
}
