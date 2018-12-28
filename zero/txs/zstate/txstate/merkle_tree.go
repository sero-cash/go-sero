package txstate

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/rlp"
)

const DEPTH = cpt.DEPTH

func Combine(l *keys.Uint256, r *keys.Uint256) (out keys.Uint256) {
	return cpt.Combine(l, r)
}

func toDepth(index uint64) (ret uint8) {
	ret = 0
	for index != 0 {
		index >>= 1
		ret++
	}
	return DEPTH + 1 - ret
}
func toPow2(index int) (ret uint64) {
	ret = uint64(1) << uint64(index)
	return
}

var (
	indexLeafKey = keys.Uint256(crypto.Keccak256Hash([]byte("outTreeIndex_Leaf")))
	indexTreeKey = keys.Uint256(crypto.Keccak256Hash([]byte("outTreeIndex_Self")))
	cap          = toPow2(DEPTH + 1)
	startIndex   = toPow2(DEPTH)
)

type Leaf struct {
	leafIndex uint64
	treeIndex uint64
	value     keys.Uint256
}

type MerkleTree struct {
	db tri.Tri
}

func NewMerkleTree(db tri.Tri) (ret MerkleTree) {
	ret.db = db
	return
}

func (self *MerkleTree) AppendLeaf(value keys.Uint256) keys.Uint256 {
	leafIndex := self.nextLeafIndex()
	treeIndex := self.geCurrentTreeIndex()
	self.db.SetState(leafKey(value).NewRef(), keys.Uint64_To_Uint256(leafIndex).NewRef())
	self.db.SetState(treeKey(value).NewRef(), keys.Uint64_To_Uint256(treeIndex).NewRef())
	self.db.SetState(indexPathKey(leafIndex, treeIndex).NewRef(), &value)

	current_value := value
	depth := toDepth(leafIndex)
	for leafIndex != 1 {
		brotherIndex := brother(leafIndex)
		var brotherValue keys.Uint256
		if brotherIndex > leafIndex {
			brotherValue = cpt.EmptyRoots()[depth]
		} else {
			brotherValue = self.db.GetState(indexPathKey(brotherIndex, treeIndex).NewRef())
			if brotherValue == keys.Empty_Uint256 {
				panic(fmt.Sprintf("brother value is empty"))
			}
		}

		if leafIndex%2 == 0 {
			current_value = Combine(&current_value, &brotherValue)
		} else {
			current_value = Combine(&brotherValue, &current_value)
		}

		leafIndex = parent(leafIndex)
		depth++
		self.db.SetState(indexPathKey(leafIndex, treeIndex).NewRef(), &current_value)
	}
	return current_value
}

func CalcRoot(value *keys.Uint256, pos uint64, paths *[DEPTH]keys.Uint256) (ret keys.Uint256) {
	ret = *value
	for _, path := range paths {
		if pos%2 == 0 {
			ret = Combine(&ret, &path)
		} else {
			ret = Combine(&path, &ret)
		}
		pos >>= 1
	}
	return
}

func (self *MerkleTree) GetPaths(value keys.Uint256) (pos uint64, paths [DEPTH]keys.Uint256, anchor keys.Uint256) {
	leafIndex := keys.Uint256_To_Uint64(self.db.GetState(leafKey(value).NewRef()).NewRef())
	if leafIndex == 0 {
		panic(fmt.Errorf("leaf index can not be 0"))
	}
	cur_leafIndex := self.getCurrentLeafIndex() - 1
	treeIndex := keys.Uint256_To_Uint64(self.db.GetState(treeKey(value).NewRef()).NewRef())

	anchor = self.db.GetState(indexPathKey(1, treeIndex).NewRef())
	pos = leafIndex - startIndex

	depth := toDepth(leafIndex)
	for leafIndex != 1 {
		brotherIndex := brother(leafIndex)
		var brotherValue keys.Uint256
		if brotherIndex > cur_leafIndex {
			brotherValue = cpt.EmptyRoots()[depth]
		} else {
			brotherValue = self.db.GetState(indexPathKey(brotherIndex, treeIndex).NewRef())
			if brotherValue == keys.Empty_Uint256 {
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
	if leafIndex == cap {
		leafIndex = startIndex
		treeIndex := self.geCurrentTreeIndex()
		treeIndex = treeIndex + 1
		self.db.SetState(&indexTreeKey, keys.Uint64_To_Uint256(treeIndex).NewRef())
	}
	self.db.SetState(&indexLeafKey, keys.Uint64_To_Uint256(leafIndex+1).NewRef())
	return leafIndex
}

func (self *MerkleTree) getCurrentLeafIndex() uint64 {
	value := self.db.GetState(&indexLeafKey)
	index := keys.Uint256_To_Uint64(&value)
	if index == 0 {
		index = startIndex
	}
	return index
}

func (self *MerkleTree) geCurrentTreeIndex() uint64 {
	value := self.db.GetState(&indexTreeKey)
	index := keys.Uint256_To_Uint64(&value)
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

func leafKey(value keys.Uint256) keys.Uint256 {
	return keys.Uint256(crypto.Keccak256Hash(append([]byte("LEAF_"), value[:]...)))
}

func treeKey(value keys.Uint256) keys.Uint256 {
	return keys.Uint256(crypto.Keccak256Hash(append([]byte("TREE_"), value[:]...)))
}

func indexPathKey(leafIndex, treeIndex uint64) keys.Uint256 {
	bytes, _ := rlp.EncodeToBytes([]interface{}{leafIndex, treeIndex, []byte("PATH")})
	return keys.Uint256(crypto.Keccak256Hash(bytes))
}
