package stake

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/serodb"
)

func newState() (*StakeState, *state.StateDB) {
	db := serodb.NewMemDatabase()
	state, _ := state.New(state.NewDatabase(db), nil)
	//state.GetZState()
	//return state
	//db := consensus.NewFakeDB()
	return NewStakeState(state), state
}

func TestAddShare(t *testing.T) {
	state, stateDB := newState()
	var pkr keys.PKr
	copy(pkr[:], crypto.Keccak512([]byte("123")))
	share1 := &Share{PKr: pkr, Value: big.NewInt(10000), InitNum: 10, Num: 10}
	state.AddPendingShare(share1)
	share2 := &Share{PKr: pkr, Value: big.NewInt(10001), InitNum: 10, Num: 10}
	state.AddPendingShare(share2)
	fmt.Println(common.Bytes2Hex(share1.Id()), common.Bytes2Hex(share2.Id()))

	root := stateDB.IntermediateRoot(true)
	fmt.Println("root:", root.String())

	fmt.Println(state.GetShare(common.BytesToHash(share1.Id())))
	fmt.Println(state.GetShare(common.BytesToHash(share2.Id())))

	//fmt.Println(state.ShareSize())
}

func TestCaleAvePrice(t *testing.T) {
	state, _ := newState()
	//var pkr keys.PKr
	//copy(pkr[:], crypto.Keccak512([]byte("123")))
	//share := &Share{PKr: keys.PKr{}, Value: big.NewInt(10000), InitNum: 10, Num: 10}
	//state.AddPendingShare(share)
	//root := stateDB.IntermediateRoot(true)
	//fmt.Println("root:", root.String())
	//fmt.Println(state.ShareSize())

	amount, _ := big.NewInt(0).SetString("0", 10)
	n, price, _ := state.CaleAvgPrice(amount)
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
		node, _ := tree.findByIndex(uint32(i))
		fmt.Println(common.Bytes2Hex(node.key[:]), node.num)
	}

}

func TestPosRewad(t *testing.T) {
	state, _ := newState()

	var pkr keys.PKr
	copy(pkr[:], crypto.Keccak512([]byte("123")))
	share := &Share{PKr: keys.PKr{}, Value: big.NewInt(10000), InitNum: 326592+10, Num: 326592+10}
	//state.AddPendingShare(share)
	//fmt.Println("root:", root.String())

	tree := NewTree(state)
	tree.insert(&SNode{key: common.BytesToHash(share.Id()), num: share.Num, total: share.Num})
	fmt.Println(state.ShareSize())
	fmt.Println(maxReware)
	fmt.Println(state.StakeCurrentReward(big.NewInt(3057599)))
	fmt.Println(state.StakeCurrentReward(big.NewInt(3057600)))
	fmt.Println(state.StakeCurrentReward(big.NewInt(3057600 + 8294400)))
}

func TestPosDif(t *testing.T) {
	state, _ := newState()

	var pkr keys.PKr
	copy(pkr[:], crypto.Keccak512([]byte("123")))
	share := &Share{PKr: keys.PKr{}, Value: big.NewInt(10000), InitNum: 10000, Num: 10000}
	//state.AddPendingShare(share)
	//fmt.Println("root:", root.String())

	tree := NewTree(state)
	tree.insert(&SNode{key: common.BytesToHash(share.Id()), num: share.Num, total: share.Num})
	fmt.Println(state.ShareSize())
	price := state.CurrentPrice()
	fmt.Println(price)
	//basePrice = big.NewInt(2000000000000000000)

	amount := sum(price, addition, 10000)
	fmt.Println(amount)
	fmt.Println(state.CaleAvgPrice(amount))
}
