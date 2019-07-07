package stake

import (
	"encoding/binary"
	"errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/consensus"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
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
	WishVotNum      uint32

	Fee         uint16
	Profit      *big.Int `rlp:"nil"`
	LastPayTime uint64
}

type StakePool struct {
	PKr             keys.PKr
	VotePKr         keys.PKr
	TransactionHash common.Hash
	Amount          *big.Int `rlp:"nil"`
	Fee             uint16
	ShareNum        uint32
	ChoicedNum      uint32
	MissedNum       uint32
	WishVotNum      uint32
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
		s.WishVotNum,
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
		WishVotNum:      s.WishVotNum,
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
	s.WishVotNum = obj.WishVotNum
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
		s.ChoicedNum,
		s.MissedNum,
		s.WishVotNum,
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
		ChoicedNum:      s.ChoicedNum,
		MissedNum:       s.MissedNum,
		WishVotNum:      s.ShareNum,
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
	s.ChoicedNum = obj.ChoicedNum
	s.MissedNum = obj.MissedNum
	s.WishVotNum = obj.WishVotNum
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
	statedb *state.StateDB

	sharePool    consensus.KVPoint
	shareObj     consensus.ObjPoint
	stakePoolObj consensus.ObjPoint
	//ShareDB      consensus.DBObj
}

var (
	ShareDB     = consensus.DBObj{"STAKE$SHARE$"}
	StakePoolDB = consensus.DBObj{"STAKE$POOL$"}
)

func NewStakeState(statedb *state.StateDB) *StakeState {
	cons := statedb.GetStakeCons()

	stakeState := &StakeState{statedb: statedb}

	stakeState.sharePool = consensus.NewKVPt(cons, "STAKE$SHAREPOOL$CONS$", "")
	stakeState.shareObj = consensus.NewObjPt(cons, "STAKE$SHAREOBJ$CONS", ShareDB.Pre, "share")
	stakeState.stakePoolObj = consensus.NewObjPt(cons, "STAKE$POOL$CONS", StakePoolDB.Pre, "pool")

	return stakeState
}

func (state *StakeState) SetStakeState(key common.Hash, value common.Hash) {
	state.sharePool.SetValue(key[:], value[:])
}

func (state *StakeState) GetStakeState(key common.Hash) common.Hash {
	return common.BytesToHash(state.sharePool.GetValue(key.Bytes()))
}

func (state *StakeState) UpdateShare(share *Share) {
	tree := NewTree(state)
	tree.insert(&SNode{key: common.BytesToHash(share.Id()), num: share.InitNum})
	state.updateShare(share)
}

func (state *StakeState) updateShare(share *Share) {
	state.shareObj.AddObj(share)
}

func (state *StakeState) UpdateStakePool(pool *StakePool) {
	state.stakePoolObj.AddObj(pool)
}

func (state *StakeState) ShareSize() uint32 {
	tree := NewTree(state)
	return tree.size()
}

func (state *StakeState) SeleteShare(seed common.Hash) (shares []*Share, err error) {
	tree := NewTree(state)

	ints, err := FindShareIdxs(tree.size(), 3, NewHash256PRNG(seed[:]))
	if err != nil {
		return nil, err
	}
	for _, i := range ints {
		node := tree.findByIndex(uint32(i))
		share := state.GetShare(node.key)
		if share == nil {
			return nil, errors.New("not found share by index")
		}
		shares = append(shares, share)
	}
	return
}

func (state *StakeState) GetShare(key common.Hash) *Share {
	item := state.shareObj.GetObj(key.Bytes(), &Share{})
	if item == nil {
		return nil
	}
	return item.(*Share)
}

func (state *StakeState) getShare(key common.Hash, cacheMap map[common.Hash]*Share) *Share {
	if share, ok := cacheMap[key]; ok {
		return share
	} else {
		share := state.GetShare(key)
		if share == nil {
			log.Crit("ProcessBeforeApply: Get share error", "key", common.BytesToHash(key[:]))
		}
		cacheMap[key] = share
		return share
	}

}

func (state *StakeState) GetStakePool(poolId common.Hash) *StakePool {
	item := state.stakePoolObj.GetObj(poolId.Bytes(), &Share{})
	if item == nil {
		return nil
	}
	return item.(*StakePool)
}

func (state *StakeState) getStakePool(poolId common.Hash, cacheMap map[common.Hash]*StakePool) *StakePool {

	if pool, ok := cacheMap[poolId]; ok {
		return pool
	} else {
		pool := state.GetStakePool(poolId)
		if pool == nil {
			log.Crit("ProcessBeforeApply: Get stakePoolObj error", "poolId", common.BytesToHash(poolId[:]))
		}
		cacheMap[poolId] = pool
		return pool
	}
}

func GetBlockRecords(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share, pools []*StakePool) {
	records := state.StakeDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				ret := ShareDB.GetObject(getter, each.Hash, &Share{})
				shares = append(shares, ret.(*Share))
			}
		}
		if record.Name == "pool" {
			for _, each := range record.Pairs {
				ret := StakePoolDB.GetObject(getter, each.Hash, &StakePool{})
				pools = append(pools, ret.(*StakePool))
			}
		}
	}
	return
}

func (state *StakeState) getBlockRecords(getter serodb.Getter, blockHash common.Hash, blockNumber uint64, shareCacheMap map[common.Hash]*Share, poolCacheMap map[common.Hash]*StakePool) (shares []*Share, pools []*StakePool) {
	records := ShareDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				if share, ok := shareCacheMap[key]; ok {
					shares = append(shares, share)
				} else {
					shares = append(shares, state.getShare(key, shareCacheMap))
				}
			}
		}
		if record.Name == "pool" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				if pool, ok := poolCacheMap[key]; ok {
					pools = append(pools, pool)
				} else {
					pools = append(pools, state.getStakePool(key, poolCacheMap))
				}
			}
		}
	}
	return
}

func GetSharesByBlock(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) []*Share {
	records := ShareDB.GetBlockRecords(getter, blockNumber, &blockHash)
	list := []*Share{}
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				ret := ShareDB.GetObject(getter, each.Hash, &Share{})
				list = append(list, ret.(*Share))
			}
		}

	}
	return []*Share{}
}

func (state *StakeState) getShares(getter serodb.Getter, blockHash common.Hash, blockNumber uint64, shareCacheMap map[common.Hash]*Share) []*Share {
	records := ShareDB.GetBlockRecords(getter, blockNumber, &blockHash)
	list := []*Share{}
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				if share, ok := shareCacheMap[key]; ok {
					list = append(list, share)
				} else {
					list = append(list, state.getShare(key, shareCacheMap))
				}
			}
		}

	}
	return []*Share{}
}

var (
	basePrice = big.NewInt(2000000000000000000)
	addition  = big.NewInt(368891382302157)

	baseReware = big.NewInt(2330000000000000000)
	rewareStep = big.NewInt(11022927689594)

	maxPrice = big.NewInt(5930000000000000000)

	outOfDatePeriod = uint64(544320)
	missVotedPeriod = uint64(725760)
	payPeriod       = uint64(42336)
)

func getOutOfDatePeriod() uint64 {
	if seroparam.Is_Dev() {
		return 10
	}
	return outOfDatePeriod
}

func getMissVotedPeriod() uint64 {
	if seroparam.Is_Dev() {
		return 15
	}
	return missVotedPeriod
}

func getPayPeriod() uint64 {
	if seroparam.Is_Dev() {
		return 3
	}
	return missVotedPeriod
}

func (state *StakeState) CurrentPrice() *big.Int {
	tree := NewTree(state)
	return new(big.Int).Add(basePrice, new(big.Int).Mul(addition, big.NewInt(int64(tree.size()))))
}

func sum(basePrice, addition *big.Int, n int64) *big.Int {
	return new(big.Int).Add(new(big.Int).Mul(basePrice, big.NewInt(n)), new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(big.NewInt(n), big.NewInt(n-1)), addition), big.NewInt(2)))
}

func (state *StakeState) CaleAvgPrice(amount *big.Int) (uint32, *big.Int) {
	basePrice := state.CurrentPrice()
	left := int64(1)
	right := new(big.Int).Div(amount, basePrice).Int64()
	if right <= 1 {
		return uint32(right), basePrice
	}
	minx := new(big.Int).Set(amount)
	//n := int64(0)
	for {
		if right < left {
			break
		}
		mid := (left + right) / 2
		sumAmount := sum(basePrice, addition, mid)
		sub := new(big.Int).Sub(amount, sumAmount)
		abs := new(big.Int).Abs(sub)

		if minx.Cmp(new(big.Int).Abs(abs)) > 0 {
			//n = mid
			minx = abs
		}

		if sub.Sign() < 0 {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	sumAmount := sum(basePrice, addition, left)
	if sumAmount.Cmp(amount) > 0 {
		left -= 1
		sumAmount = sum(basePrice, addition, left)
	}
	return uint32(left), new(big.Int).Div(sumAmount, big.NewInt(left))
}

func (state *StakeState) StakeCurrentReward() (*big.Int, *big.Int) {
	size := NewTree(state).size()

	soleAmount := new(big.Int).Add(baseReware, new(big.Int).Mul(rewareStep, big.NewInt(int64(size))))

	if soleAmount.Cmp(maxPrice) > 1 {
		soleAmount = new(big.Int).Set(maxPrice)
	}
	return soleAmount, new(big.Int).Div(new(big.Int).Mul(soleAmount, big.NewInt(3)), big.NewInt(2))
}

func (state *StakeState) ProcessBeforeApply(bc blockChain, header *types.Header) {

	shareCacheMap := map[common.Hash]*Share{}
	poolCacheMap := map[common.Hash]*StakePool{}

	state.processVotedShare(header, bc, shareCacheMap, poolCacheMap)
	state.processOutDate(header, bc, shareCacheMap, poolCacheMap)
	state.processMissVoted(header, bc, shareCacheMap, poolCacheMap)
	//state.processNowShares(bc.GetHeader(header.ParentHash, header.Number.Uint64()-1), poolCacheMap, bc.GetDB())
	state.payProfit(bc, header, shareCacheMap, poolCacheMap)

	for _, share := range shareCacheMap {
		state.updateShare(share)
	}
	for _, pool := range poolCacheMap {
		state.UpdateStakePool(pool)
	}
}

func (state *StakeState) processVotedShare(header *types.Header, bc blockChain, shareCacheMap map[common.Hash]*Share, poolCacheMap map[common.Hash]*StakePool) {
	if header.Number.Uint64() == 0 {
		return
	}

	preHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	tree := NewTree(state)
	hash := preHeader.Hash()
	indexs, err := FindShareIdxs(tree.size(), 3, NewHash256PRNG(hash[:]))
	if err == nil {
		for _, index := range indexs {
			sndoe := tree.deletNodeByIndex(index)
			if sndoe == nil {
				log.Crit("ProcessBeforeApply: selete share error", "blockHash", hash.String(), "blockNumber", preHeader.Number, "index", index, "poolSize", tree.size())
			} else {
				share := state.getShare(sndoe.key, shareCacheMap)
				if share.Num != sndoe.num {
					log.Crit("ProcessBeforeApply: deletNodeByIndex err", "share.num", share.Num, "snode.num", sndoe.num)
				}
				share.WishVotNum += 1

				if share.Num > 0 {
					share.Num -= 1
				} else {
					log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.Num=0")
				}

				if share.PoolId != nil {
					pool := state.getStakePool(*share.PoolId, poolCacheMap)
					pool.ChoicedNum += 1
					pool.MissedNum += 1

					if pool.ShareNum > 0 {
						pool.ShareNum -= 1
					} else {
						log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.ShareNum=0")
					}
					pool.WishVotNum += 1
				}
			}
		}
	}

	soloReware, reward := state.StakeCurrentReward()
	for _, vote := range preHeader.CurrentVotes {
		share := state.getShare(vote.Hash, shareCacheMap)

		if share.WishVotNum > 0 {
			share.WishVotNum -= 1
		} else {
			log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.WishVotNum=0")
		}

		pool := state.getStakePool(*share.PoolId, poolCacheMap)
		if pool != nil {
			if pool.WishVotNum > 0 {
				pool.WishVotNum -= 1
			} else {
				log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.WishVotNum=0")
			}
		}

		if vote.IsPool {
			if pool.MissedNum > 0 {
				pool.MissedNum -= 1
			} else {
				log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.MissedNum=0")
			}
			poolProfit := new(big.Int).Div(new(big.Int).Mul(reward, big.NewInt(int64(share.Fee))), big.NewInt(10000))
			pool.AddProfit(poolProfit)
			share.AddProfit(new(big.Int).Add(share.Value, new(big.Int).Sub(reward, poolProfit)))

			if pool.Closed && pool.ShareNum == 0 && pool.WishVotNum == 0 {
				pool.AddProfit(pool.Amount)
			}
		} else {
			share.AddProfit(soloReware)
		}
	}

	if len(preHeader.ParentHash) > 0 {
		reward = new(big.Int).Sub(reward, new(big.Int).Div(reward, big.NewInt(3)))
		soloReware = new(big.Int).Sub(soloReware, new(big.Int).Div(soloReware, big.NewInt(3)))
		for _, vote := range preHeader.ParentVotes {
			share := state.getShare(vote.Hash, shareCacheMap)

			if share.WishVotNum > 0 {
				share.WishVotNum -= 1
			} else {
				log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.WishVotNum=0")
			}

			pool := state.getStakePool(*share.PoolId, poolCacheMap)
			if pool != nil {
				if pool.WishVotNum > 0 {
					pool.WishVotNum -= 1
				} else {
					log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.WishVotNum=0")
				}
			}

			if vote.IsPool {
				if pool.MissedNum > 0 {
					pool.MissedNum -= 1
				} else {
					log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.MissedNum=0")
				}
				poolProfit := new(big.Int).Div(new(big.Int).Mul(reward, big.NewInt(int64(share.Fee))), big.NewInt(10000))
				pool.AddProfit(poolProfit)
				share.AddProfit(new(big.Int).Add(share.Value, new(big.Int).Sub(reward, poolProfit)))

				if pool.Closed && pool.ShareNum == 0 && pool.WishVotNum == 0 {
					pool.AddProfit(pool.Amount)
				}
			} else {
				share.AddProfit(soloReware)
			}
		}
	}

}

func (state *StakeState) processOutDate(header *types.Header, bc blockChain, shareCacheMap map[common.Hash]*Share, poolCacheMap map[common.Hash]*StakePool) (shares []*Share) {
	outOfDatePeriod := getOutOfDatePeriod()
	if header.Number.Uint64() < outOfDatePeriod {
		return
	}
	tree := NewTree(state)
	preHeader := bc.GetHeaderByNumber(header.Number.Uint64() - outOfDatePeriod)
	shares = state.getShares(bc.GetDB(), header.Hash(), header.Number.Uint64(), shareCacheMap)
	if len(shares) > 0 {
		for _, share := range shares {
			if share.BlockNumber == preHeader.Number.Uint64() {
				if share.Num == 0 {
					continue
				}
				sndoe := tree.deleteNodeByHash(common.BytesToHash(share.Id()))
				if sndoe == nil {
					log.Crit("ProcessBeforeApply: processOutDate share not found", "hash", common.Bytes2Hex(share.Id()), "InitNum", share.InitNum, "Num", share.Num, "WishVotNum", share.WishVotNum)
				}
				if share.Num != sndoe.num {
					log.Crit("ProcessBeforeApply: processOutDate err", "share.num", share.Num, "snode.num", sndoe.num)
				}

				pool := state.getStakePool(*share.PoolId, poolCacheMap)
				if pool != nil {
					if pool.ShareNum >= share.Num {
						pool.ShareNum -= share.Num
					} else {
						log.Crit("ProcessBeforeApply: processOutDate err", "pool.ShareNum", pool.ShareNum, "share.num", share.Num)
					}
					if pool.Closed && pool.ShareNum == 0 && pool.WishVotNum == 0 {
						pool.AddProfit(pool.Amount)
					}
				}

				share.AddProfit(new(big.Int).Mul(share.Value, big.NewInt(int64(share.Num))))
				share.Num = 0
			}
		}
	}
	return
}

func (state *StakeState) processMissVoted(header *types.Header, bc blockChain, shareCacheMap map[common.Hash]*Share, poolCacheMap map[common.Hash]*StakePool) {
	missVotedPeriod := getMissVotedPeriod()
	if header.Number.Uint64() < missVotedPeriod {
		return
	}
	perHeader := bc.GetHeaderByNumber(header.Number.Uint64() - missVotedPeriod)
	shares := state.getShares(bc.GetDB(), perHeader.Hash(), perHeader.Number.Uint64(), shareCacheMap)
	if len(shares) > 0 {
		for _, share := range shares {
			if share.BlockNumber == perHeader.Number.Uint64() {
				if share.WishVotNum != 0 {

					if share.Num != 0 {
						log.Crit("ProcessBeforeApply: processOutDate err, snode.num ！= 0")
					}

					pool := state.getStakePool(*share.PoolId, poolCacheMap)
					if pool != nil {
						pool.WishVotNum -= share.WishVotNum
						if pool.Closed && pool.ShareNum == 0 && pool.WishVotNum == 0 {
							pool.AddProfit(pool.Amount)
						}
					}

					share.AddProfit(new(big.Int).Mul(share.Value, big.NewInt(int64(share.WishVotNum))))
					share.WishVotNum = 0
				} else {
					log.Error("ProcessBeforeApply: processMissVoted err", "share.WishVotNum", share.WishVotNum, "share.BlockNumber", share.BlockNumber, "currenBlockNumber", header.Number)
				}
			}
		}
	}
}

func (state *StakeState) processNowShares(header *types.Header, shareCashMap map[common.Hash]*Share, poolCashMap map[common.Hash]*StakePool, db serodb.Database) {
	number := header.Number.Uint64()
	shares := state.getShares(db, header.Hash(), number, shareCashMap)
	if len(shares) > 0 {
		tree := NewTree(state)
		for _, share := range shares {
			if share.BlockNumber != number {
				continue
			}
			tree.insert(&SNode{key: common.BytesToHash(share.State()), num: share.Num})
			if share.PoolId != nil {
				pool := state.getStakePool(*share.PoolId, poolCashMap)
				pool.ShareNum += share.Num
			}
		}
	}
}

func (state *StakeState) payProfit(bc blockChain, header *types.Header, shareCashMap map[common.Hash]*Share, poolCashMap map[common.Hash]*StakePool) {
	payPeriod := getPayPeriod()
	if header.Number.Uint64() < payPeriod {
		return
	}
	preHeader := bc.GetHeaderByNumber(header.Number.Uint64() - payPeriod)
	shares, pools := state.getBlockRecords(bc.GetDB(), preHeader.Hash(), preHeader.Number.Uint64(), shareCashMap, poolCashMap)

	for _, share := range shares {
		if share.Profit.Sign() > 0 && header.Number.Uint64()-share.LastPayTime >= payPeriod {
			addr := common.Address{}
			copy(addr[:], share.PKr[:])
			asset := assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
				Value:    utils.U256(*share.Profit),
			},
			}
			share.LastPayTime = header.Number.Uint64()
			share.SetProfit(big.NewInt(0))
			state.statedb.GetZState().AddTxOutWithCheck(addr, asset)
		}
	}
	for _, pool := range pools {
		if pool.Profit.Sign() > 0 && header.Number.Uint64()-pool.LastPayTime >= payPeriod {
			addr := common.Address{}
			copy(addr[:], pool.PKr[:])
			asset := assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
				Value:    utils.U256(*pool.Profit),
			},
			}
			pool.LastPayTime = header.Number.Uint64()
			pool.SetProfit(big.NewInt(0))
			state.statedb.GetZState().AddTxOutWithCheck(addr, asset)
		}
	}
}

func (state *StakeState) deleteShare(tree *STree, share *Share, poolCashMap map[common.Hash]*StakePool) {
	sndoe := tree.deleteNodeByHash(common.BytesToHash(share.Id()))
	if sndoe == nil {
		log.Crit("ProcessBeforeApply: deleteShare share not found", "hash", common.Bytes2Hex(share.Id()), "InitNum", share.InitNum, "Num", share.Num, "WishVotNum", share.WishVotNum)
	}
	if share.Num != sndoe.num {
		log.Crit("ProcessBeforeApply: deleteShare err", "share.num", share.Num, "snode.num", sndoe.num)
	}

	pool := state.getStakePool(*share.PoolId, poolCashMap)
	if pool != nil {
		pool.ShareNum -= share.Num
	}

	share.Num = 0
	share.AddProfit(new(big.Int).Mul(share.Value, big.NewInt(int64(share.Num))))

}

func decodeNumber32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func encodeNumber32(number uint32) []byte {
	enc := make([]byte, 4)
	binary.BigEndian.PutUint32(enc, number)
	return enc
}
