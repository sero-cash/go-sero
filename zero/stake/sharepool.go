package stake

import (
	"encoding/binary"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/consensus"
	"math/big"
)

type Share struct {
	PKr             keys.PKr
	VoteKr          *keys.PKr `rlp:"nil"`
	TransactionHash common.Hash
	PoolId          *common.Hash `rlp:"nil"`
	Value           *big.Int     `rlp:"nil"`
	BlockNumber     uint64
	InitNum         uint32
	Num             uint32
	VotNum          uint32
	Fee             uint16
	Profit          *big.Int `rlp:"nil"`
	LastPayTime     uint64
}

type StakePool struct {
	PKr             keys.PKr
	VotePKr         keys.PKr
	TransactionHash common.Hash
	Amount          *big.Int `rlp:"nil"`
	Fee             uint16
	ShareNum        uint32
	Profit          *big.Int `rlp:"nil"`
	LastPayTime     uint64
	Closed          bool
}

func (s *Share) Id() []byte {
	hw := sha3.NewKeccak256()
	hash := common.Hash{}
	rlp.Encode(hw, []interface{}{
		s.PKr,
		s.VoteKr,
		s.TransactionHash,
		s.PoolId,
		s.Value,
		s.InitNum,
		s.BlockNumber,
		s.Fee,
	})
	hw.Sum(hash[:0])
	return hash.Bytes()
}

func (s *Share) State() []byte {
	hw := sha3.NewKeccak256()
	hash := common.Hash{}
	rlp.Encode(hw, []interface{}{
		s.Id(),
		s.Num,
		s.VotNum,
		s.Profit,
		s.LastPayTime,
	})
	hw.Sum(hash[:0])
	return hash.Bytes()
}

func (s *Share) CopyTo() (ret consensus.CItem) {
	return &Share{
		PKr:             s.PKr,
		VoteKr:          s.VoteKr,
		TransactionHash: s.TransactionHash,
		PoolId:          s.PoolId,
		Value:           new(big.Int).Set(s.Value),
		BlockNumber:     s.BlockNumber,
		InitNum:         s.InitNum,
		Num:             s.Num,
		VotNum:          s.VotNum,
		Fee:             s.Fee,
		Profit:          s.Profit,
		LastPayTime:     s.LastPayTime,
	}
}

func (s *Share) CopyFrom(ret consensus.CItem) {
	obj := ret.(*Share)
	s.PKr = obj.PKr
	s.VoteKr = obj.VoteKr
	s.TransactionHash = obj.TransactionHash
	s.PoolId = obj.PoolId
	s.BlockNumber = obj.BlockNumber
	s.Value = new(big.Int).Set(obj.Value)
	s.Fee = obj.Fee
	s.InitNum = obj.InitNum
	s.Num = obj.Num
	s.VotNum = obj.VotNum
	s.Profit = new(big.Int).Set(obj.Profit)
	s.LastPayTime = s.LastPayTime
}

func (s *Share) AddProfit(profit *big.Int) {
	if s.Profit == nil {
		s.Profit = new(big.Int)
	}
	s.Profit.Add(s.Profit, profit)
}

func (s *Share) SetProfit(profit *big.Int) {
	s.Profit = profit
}

func (s *StakePool) Id() []byte {
	return crypto.Keccak256Hash(s.PKr[:]).Bytes()
}

func (s *StakePool) State() []byte {
	hw := sha3.NewKeccak256()
	hash := common.Hash{}
	rlp.Encode(hw, []interface{}{
		s.Id(),
		s.VotePKr,
		s.TransactionHash,
		s.Fee,
		s.Amount,
		s.ShareNum,
		s.Profit,
		s.LastPayTime,
		s.Closed,
	})
	hw.Sum(hash[:0])
	return hash.Bytes()
}

func (s *StakePool) CopyTo() (ret consensus.CItem) {
	return &StakePool{
		PKr:             s.PKr,
		VotePKr:         s.VotePKr,
		TransactionHash: s.TransactionHash,
		Amount:          new(big.Int).Set(s.Amount),
		Fee:             s.Fee,
		ShareNum:        s.ShareNum,
		Profit:          s.Profit,
		LastPayTime:     s.LastPayTime,
		Closed:          s.Closed,
	}
}

func (s *StakePool) CopyFrom(ret consensus.CItem) {
	obj := ret.(*StakePool)
	s.PKr = obj.PKr
	s.VotePKr = obj.VotePKr
	s.TransactionHash = obj.TransactionHash
	s.Amount = new(big.Int).Set(obj.Amount)
	s.Fee = obj.Fee
	s.ShareNum = obj.ShareNum
	s.Closed = obj.Closed
	s.Profit = new(big.Int).Set(obj.Profit)
	s.LastPayTime = s.LastPayTime
}

func (s *StakePool) AddProfit(profit *big.Int) {
	if s.Profit == nil {
		s.Profit = new(big.Int)
	}
	s.Profit.Add(s.Profit, profit)
}

func (s *StakePool) SetProfit(profit *big.Int) {
	s.Profit = profit
}

type blockChain interface {
	GetHeader(hash common.Hash, number uint64) *types.Header
	GetHeaderByNumber(number uint64) *types.Header
	StateAt(root common.Hash, number uint64) (*state.StateDB, error)
	GetDB() serodb.Database
}

type State interface {
	SetStakeState(key common.Hash, value common.Hash)
	GetStakeState(key common.Hash) common.Hash
}

type StakeState struct {
	sharePool consensus.KVPoint
	shareObj  consensus.ObjPoint
	stakePool consensus.ObjPoint
	shareDB   consensus.DBObj
}

func NewStakeState(db consensus.DB) *StakeState {

	stakeState := &StakeState{}

	//zeroDB := statedb.GetZeroDB()
	//db := consensus.NewFakeDB()
	stakeState.shareDB = consensus.DBObj{"SHARE$DB$"}
	stakeDB := consensus.DBObj{"STAKE$POOL$"}
	blockDB := consensus.DBObj{"STAKE$BLOCK$INDEX$"}

	cons := consensus.NewCons(db, blockDB.Pre)

	stakeState.sharePool = consensus.NewKVPt(&cons, "SHARE$POOL$CONS$", "")
	stakeState.shareObj = consensus.NewObjPt(&cons, "SHARE$OBJ$CONS", stakeState.shareDB.Pre, "share")
	stakeState.stakePool = consensus.NewObjPt(&cons, "STAKE$POOL$CONS", stakeDB.Pre, "")

	return stakeState
}

func (state *StakeState) SetStakeState(key common.Hash, value common.Hash) {
	state.sharePool.SetValue(key[:], value[:])
}

func (state *StakeState) GetStakeState(key common.Hash) common.Hash {
	return common.BytesToHash(state.sharePool.GetValue(key.Bytes()))
}

func (state *StakeState) UpdateShare(share *Share) {
	state.shareObj.AddObj(share)
}

func (state *StakeState) UpdateStakePool(pool *StakePool) {
	state.stakePool.AddObj(pool)
}

func (state *StakeState) ShareSize() uint32 {
	tree := NewTree(state)
	return tree.size()
}

func (state *StakeState) GetShare(key common.Hash) *Share {
	item := state.shareObj.GetObj(key.Bytes(), &Share{})
	if item == nil {
		return nil
	}
	return item.(*Share)
}

func (state *StakeState) getShare(key common.Hash, cashMap map[common.Hash]*Share) *Share {
	if share, ok := cashMap[key]; ok {
		return share
	} else {
		share := state.GetShare(key)
		if share == nil {
			log.Crit("ProcessBeforeApply: Get share error", "key", common.BytesToHash(key[:]))
		}
		cashMap[key] = share
		return share
	}

}
func (state *StakeState) GetStakePool(poolId common.Hash) *StakePool {
	item := state.stakePool.GetObj(poolId.Bytes(), &Share{})
	if item == nil {
		return nil
	}
	return item.(*StakePool)
}
func (state *StakeState) getStakePool(poolId common.Hash, cashMap map[common.Hash]*StakePool) *StakePool {

	if pool, ok := cashMap[poolId]; ok {
		return pool
	} else {
		pool := state.GetStakePool(poolId)
		if pool == nil {
			log.Crit("ProcessBeforeApply: Get stakePool error", "poolId", common.BytesToHash(poolId[:]))
		}
		cashMap[poolId] = pool
		return pool
	}
}

func (state *StakeState) GetShares(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) []*Share {
	records := state.shareDB.GetBlockRecords(getter, blockNumber, &blockHash)
	list := []*Share{}
	for _, record := range records {
		for _, hash := range record.Hashes {

			ret := state.shareDB.GetObject(getter, hash, &Share{})
			list = append(list, ret.(*Share))
		}
	}
	return []*Share{}
}

var (
	basePrice   = big.NewInt(2000000000000000000)
	addition, _ = new(big.Int).SetString("368891382302157", 10)

	baseReware    = big.NewInt(2330000000000000000)
	rewareStep, _ = new(big.Int).SetString("11022927689594", 10)

	maxPrice = big.NewInt(5930000000000000000)

	outOfDateStep = uint64(544320)
	missVotedStep = uint64(725760)
	payStep       = uint64(42000)
)

func (state *StakeState) SharePrice(n uint32) *big.Int {
	return new(big.Int).Add(basePrice, new(big.Int).Mul(addition, big.NewInt(int64(n))))
}

func (state *StakeState) stakeReward() (*big.Int, *big.Int) {
	size := NewTree(state).size()

	soleAmount := new(big.Int).Add(baseReware, new(big.Int).Mul(rewareStep, big.NewInt(int64(size))))

	if soleAmount.Cmp(maxPrice) > 1 {
		soleAmount = new(big.Int).Set(maxPrice)
	}
	return soleAmount, new(big.Int).Div(new(big.Int).Mul(soleAmount, big.NewInt(3)), big.NewInt(2))
}
func (state *StakeState) pay() {

}

func (state *StakeState) processVotedShare(header *types.Header, shareCashMap map[common.Hash]*Share, poolCashMap map[common.Hash]*StakePool) {
	tree := NewTree(state)

	indexs := randomIndexs(header.Hash(), tree.size(), 3)
	for _, index := range indexs {
		sndoe := tree.deletNodeByIndex(index)
		if sndoe == nil {
			log.Crit("ProcessBeforeApply: selete share error", "blockHash", header.Hash().String(), "blockNumber", header.Number, "index", index, "poolSize", tree.size())
		} else {
			share := state.getShare(sndoe.key, shareCashMap)
			if share.Num != sndoe.num {
				log.Crit("ProcessBeforeApply: deletNodeByIndex err", "share.num", share.Num, "snode.num", sndoe.num)
			}
			share.VotNum += 1
			share.Num -= 1
			if share.PoolId != nil {
				pool := state.getStakePool(*share.PoolId, poolCashMap)
				pool.ShareNum -= share.Num
			}
		}
	}

	soloReware, reward := state.stakeReward()
	for _, vote := range append(header.CurrentVotes, header.ParentVotes...) {
		share := state.getShare(vote.Hash, shareCashMap)
		if share.VotNum > 0 {
			share.VotNum -= 1
		} else {
			log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "stakePool not exist")
		}

		if vote.Sign[0] != 1 {
			share.AddProfit(soloReware)
		} else {
			pool := state.getStakePool(*share.PoolId, poolCashMap)
			shareProfit := new(big.Int).Div(new(big.Int).Mul(reward, big.NewInt(int64(share.Fee))), big.NewInt(10000))
			share.AddProfit(shareProfit)
			pool.AddProfit(new(big.Int).Sub(reward, shareProfit))
		}
	}
}
func (state *StakeState) ProcessBeforeApply(bc blockChain, block *types.Block) {

	shareCashMap := map[common.Hash]*Share{}
	poolCashMap := map[common.Hash]*StakePool{}

	state.processVotedShare(bc.GetHeader(block.ParentHash(), block.NumberU64()-1), shareCashMap, poolCashMap)
	state.processOutDate(bc.GetHeaderByNumber(block.NumberU64()-outOfDateStep), bc, poolCashMap, bc.GetDB())

	state.processNowShares(bc.GetHeader(block.ParentHash(), block.NumberU64()-1), shareCashMap, poolCashMap, bc.GetDB())

	for _, share := range shareCashMap {
		state.UpdateShare(share)
	}
	for _, pool := range poolCashMap {
		state.UpdateStakePool(pool)
	}
}

func (state *StakeState) processNowShares(header *types.Header, shareCashMap map[common.Hash]*Share, poolCashMap map[common.Hash]*StakePool, db serodb.Database) {
	shares := state.GetShares(db, header.Hash(), header.Number.Uint64())
	if len(shares) > 0 {
		tree := NewTree(state)
		for _, share := range shares {
			tree.insert(&SNode{key: common.BytesToHash(share.State()), num: share.Num})
			if share.PoolId != nil {
				pool := state.getStakePool(*share.PoolId, poolCashMap)
				pool.ShareNum += share.Num
			}
		}
	}
}

func (state *StakeState) processOutDate(header *types.Header, bc blockChain, poolCashMap map[common.Hash]*StakePool, db serodb.Database) []*Share {

	tree := NewTree(state)
	shares := state.GetShares(db, header.Hash(), header.Number.Uint64())
	if len(shares) > 0 {
		for _, share := range shares {
			if share.BlockNumber > header.Number.Uint64() {
				continue
			}

			if share.Num == 0 || share.VotNum != 0 {
				continue
			}

			sndoe := tree.deleteNodeByHash(common.BytesToHash(share.Id()))
			if sndoe == nil {
				log.Crit("ProcessBeforeApply: deleteNodeByHash share not found", "hash", common.Bytes2Hex(share.Id()), "InitNum", share.InitNum, "Num", share.Num, "VotNum", share.VotNum)
			}
			if share.Num != sndoe.num {
				log.Crit("ProcessBeforeApply: deleteNodeByHash err", "share.num", share.Num, "snode.num", sndoe.num)
			}

			share.AddProfit(new(big.Int).Mul(share.Value, big.NewInt(int64(share.Num))))

			if share.Num != sndoe.num {
				log.Crit("ProcessBeforeApply: deleteNodeByHash err", "share.num", share.Num, "snode.num", sndoe.num)
			} else {
				pool := state.getStakePool(*share.PoolId, poolCashMap)
				share.Num = 0
				pool.ShareNum -= share.Num
			}
		}
	}
	return shares
}

func randomIndexs(hash common.Hash, seiz uint32, n int) (indexs []uint32) {
	return
}

func ReadCanonicalHash(db serodb.Database, number uint64) common.Hash {
	data, _ := db.Get(headerHashKey(number))
	if len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

func SharePrice(state StakeState) *big.Int {
	//2+(0.000011022927689594*n)
	return big.NewInt(0)
}

var (
	headerPrefix     = []byte("h")
	headerHashSuffix = []byte("n") // headerPrefix + num (uint64 big endian) + headerHashSuffix -> key
)
// headerHashKey = headerPrefix + num (uint64 big endian) + headerHashSuffix
func headerHashKey(number uint64) []byte {
	return append(append(headerPrefix, encodeNumber64(number)...), headerHashSuffix...)
}

func shareKey(hash common.Hash) []byte {
	return append([]byte("ps"), hash.Bytes()...)
}

func sharesBlockKey(number uint64, hash common.Hash) []byte {
	return append(append([]byte("pb"), encodeNumber64(number)...), hash.Bytes()...)
}

func stakePoolKey(hash common.Hash) []byte {
	return append([]byte("pl"), hash.Bytes()...)
}

func StakeRewardKey(pkr keys.PKr) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, []interface{}{
		pkr,
		"STAKEREWARD",
	})
	hw.Sum(h[:0])
	return h
}

func decodeNumber64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

func encodeNumber64(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func decodeNumber32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func encodeNumber32(number uint32) []byte {
	enc := make([]byte, 4)
	binary.BigEndian.PutUint32(enc, number)
	return enc
}
