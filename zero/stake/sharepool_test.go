package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/serodb"
	"math/big"
	"testing"
)

func newState() (*StakeState, *state.StateDB) {
	db := serodb.NewMemDatabase()
	state, _ := state.New(common.Hash{}, state.NewDatabase(db), 0)
	//state.GetZState()
	//return state
	//db := consensus.NewFakeDB()
	return NewStakeState(state), state
}

func TestCaleAvePrice(t *testing.T) {
	state, _ := newState()
	//var pkr keys.PKr
	//copy(pkr[:], crypto.Keccak512([]byte("123")))
	//share := &Share{PKr: keys.PKr{}, Value: big.NewInt(10000), InitNum: 10, Num: 10}
	//state.UpdateShare(share)
	//root := stateDB.IntermediateRoot(true)
	//fmt.Println("root:", root.String())
	//fmt.Println(state.ShareSize())

	amount, _ := big.NewInt(0).SetString("98000000000000000000", 10)
	n, price := state.CaleAvgPrice(amount)
	sum := sum(basePrice, addition, int64(n))
	fmt.Println(n, price, sum)
	fmt.Println(new(big.Int).Mul(big.NewInt(int64(n)), price))
}

func TestSeleteShare(t *testing.T) {
	state, stateDB := newState()
	tree, _ := initTree(state)
	fmt.Println()
	stateDB.IntermediateRoot(true)

	seed := crypto.Keccak256Hash([]byte("abc"))
	prng := NewHash256PRNG(seed[:])

	ints, err := FindShareIdxs(tree.size(), 3, prng)
	fmt.Println(ints, err)

	for _, i := range ints {
		node := tree.findByIndex(uint32(i))
		fmt.Println(common.Bytes2Hex(node.key[:]), node.num)
	}

}
