package state

import (
	"fmt"
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/rlp"
	math2 "math"
	"math/big"
)

const DEPTH = cpt.DEPTH + 1

var (
	indexLeafKey = crypto.Keccak256Hash([]byte("outTreeIndex_Leaf"))
	indexTreeKey = crypto.Keccak256Hash([]byte("outTreeIndex_Self"))
	cap          = math.BigPow(2, DEPTH)
	startIndex   = math.BigPow(2, DEPTH-1).Uint64()
)

type Leaf struct {
	leafIndex uint64
	treeIndex uint64
	value     keys.Uint256
}

type OutState struct {
	db *StateDB
}

func NewOutStete(db *StateDB) OutState {
	return OutState{db}
}

func (self *OutState) AppendLeaf(value keys.Uint256) keys.Uint256 {
	leafIndex := self.nextLeafIndex()
	treeIndex := self.geCurrentTreeIndex()
	self.db.SetState(EmptyAddress, leafKey(value), common.BigToHash(new(big.Int).SetUint64(leafIndex)))
	self.db.SetState(EmptyAddress, treeKey(value), common.BigToHash(new(big.Int).SetUint64(treeIndex)))
	self.db.SetState(EmptyAddress, indexPathKey(leafIndex, treeIndex), common.Hash(value))

	current_value := value
	for ; leafIndex != 1; {
		brotherValue := self.db.GetState(EmptyAddress, indexPathKey(brother(leafIndex), treeIndex))
		if brotherValue == (common.Hash{}) {
			brotherValue = common.Hash{uint8(math2.Log2(float64(leafIndex)))}
		}

		if leafIndex%2 == 0 {
			current_value = cpt.Combine(&current_value, brotherValue.HashToUint256())
		} else {
			current_value = cpt.Combine(brotherValue.HashToUint256(), &current_value)
		}

		leafIndex = parent(leafIndex)
		self.db.SetState(EmptyAddress, indexPathKey(leafIndex, treeIndex), common.Hash(current_value))
	}
	return current_value
}

func (self *OutState) GetPaths(value keys.Uint256) (uint64, []*keys.Uint256) {
	leafIndex := new(big.Int).SetBytes(self.db.GetState(EmptyAddress, leafKey(value)).Bytes()).Uint64()
	if leafIndex == 0 {
		return 0, []*keys.Uint256{}
	}
	treeIndex := new(big.Int).SetBytes(self.db.GetState(EmptyAddress, treeKey(value)).Bytes()).Uint64()

	uint256Paths := []*keys.Uint256{}
	uint64s := paths(leafIndex)
	fmt.Println("path:", uint64s)
	for _, index := range uint64s {
		uint256Paths = append(uint256Paths, self.db.GetState(EmptyAddress, indexPathKey(index, treeIndex)).HashToUint256())
	}
	return leafIndex - startIndex, uint256Paths
}

func (self *OutState) nextLeafIndex() uint64 {
	leafIndex := self.getCurrentLeafIndex()
	if leafIndex == cap.Uint64() {
		leafIndex = startIndex
		treeIndex := self.geCurrentTreeIndex()
		treeIndex = treeIndex + 1
		self.db.SetState(EmptyAddress, indexTreeKey, common.BigToHash(new(big.Int).SetUint64(treeIndex)))
	}
	self.db.SetState(EmptyAddress, indexLeafKey, common.BigToHash(new(big.Int).SetUint64(leafIndex+1)))
	return leafIndex
}

func (self *OutState) getCurrentLeafIndex() uint64 {
	value := self.db.GetState(EmptyAddress, indexLeafKey)
	index := new(big.Int).SetBytes(value[:])
	if index.Sign() == 0 {
		return startIndex
	}
	return index.Uint64()
}

func (self *OutState) geCurrentTreeIndex() uint64 {
	value := self.db.GetState(EmptyAddress, indexTreeKey)
	index := new(big.Int).SetBytes(value[:])
	return index.Uint64()
}

func paths(index uint64) []uint64 {
	paths := []uint64{}
	if index == 0 {
		return paths;
	}

	for ; index != 1; {
		paths = append(paths, brother(index))
		index = parent(index)
	}
	return paths
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

func leafKey(value keys.Uint256) common.Hash {
	return crypto.Keccak256Hash(append([]byte("LEAF_"), value[:]...))
}

func treeKey(value keys.Uint256) common.Hash {
	return crypto.Keccak256Hash(append([]byte("TREE_"), value[:]...))
}

func indexPathKey(leafIndex, treeIndex uint64) common.Hash {
	bytes, _ := rlp.EncodeToBytes([]interface{}{leafIndex, treeIndex, []byte("PATH")})
	return crypto.Keccak256Hash(bytes)
}
