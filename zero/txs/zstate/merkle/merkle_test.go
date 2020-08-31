package merkle

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/consensus"
)

type TreeState struct {
	db *consensus.FakeTri
}

func (self *TreeState) TryGet(key []byte) ([]byte, error) {
	return nil, nil
}
func (self *TreeState) TryUpdate(key, value []byte) error {
	return nil
}

func (self *TreeState) SetState(obj *c_type.PKr, key *c_type.Uint256, value *c_type.Uint256) {
	self.db.TryUpdate(key[:], value[:])
}
func (self *TreeState) GetState(obj *c_type.PKr, key *c_type.Uint256) (ret c_type.Uint256) {
	r, e := self.db.TryGet(key[:])
	if e == nil {
		copy(ret[:], r)
	}
	return
}
func (self *TreeState) GlobalGetter() serodb.Getter {
	return nil
}

var Address = c_type.PKr{}
var MerkleParam = NewParam(&Address, c_superzk.Combine)

func TestOutTree(t *testing.T) {
	// Create an empty state database
	superzk.ZeroInit("", 0)

	ft := consensus.NewFakeTri()
	outState := MerkleParam.NewMerkleTree(&TreeState{db: &ft})

	for i := 1; i <= 100; i++ {
		value := crypto.Keccak256Hash(big.NewInt(int64(i)).Bytes()).HashToUint256()
		outState.AppendLeaf(*value)

		/*for i := 1; i <= 15; i++ {
			key := indexPathKey(uint64(i), uint64(0))
			value := outState.db.GetState(&key)
			fmt.Println(i, ":", common.Bytes2Hex(value[:]))
		}*/

		/*if i == 3 {
			current := crypto.Keccak256Hash(big.NewInt(int64(1)).Bytes()).HashToUint256()
			index, getPaths, anchor := outState.GetPaths(*current)
			ret := CalcRoot(current, index, &getPaths)
			if anchor != ret {
				fmt.Println(i, 1)
				t.FailNow()
			}
		}*/

		for j := 1; j <= i; j++ {
			current := crypto.Keccak256Hash(big.NewInt(int64(j)).Bytes()).HashToUint256()
			index, getPaths, anchor := outState.GetPaths(*current)
			ret := MerkleParam.CalcRoot(current, index, &getPaths)
			if anchor != ret {
				fmt.Println(i, j)
				t.FailNow()
			}
		}
	}
}
