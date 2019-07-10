package stake

import (
	"encoding/binary"
	"errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/rawdb"
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

type Status int8

const (
	STATUS_VALID     Status = 0
	STATUS_OUTOFDATE Status = 1
	STATUS_FINISHED  Status = 2
)

type Share struct {
	PKr             keys.PKr
	VotePKr         keys.PKr
	TransactionHash common.Hash
	PoolId          *common.Hash `rlp:"nil"`
	Value           *big.Int     `rlp:"nil"`
	BlockNumber     uint64
	InitNum         uint32
	Fee             uint16

	Num          uint32
	WillVoteNum  uint32
	Status       Status
	ReturnAmount *big.Int `rlp:"nil"`
	Profit       *big.Int `rlp:"nil"`
	LastPayTime  uint64
}

func (s *Share) Id() []byte {
	hw := sha3.NewKeccak256()
	hash := common.Hash{}
	rlp.Encode(hw, []interface{}{
		s.PKr,
		s.VotePKr,
		s.TransactionHash,
		s.PoolId,
		s.Value,
		s.BlockNumber,
		s.InitNum,
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
		s.WillVoteNum,
		s.Status,
		s.ReturnAmount,
		s.Profit,
		s.LastPayTime,
	})
	hw.Sum(hash[:0])
	return hash.Bytes()
}

func (s *Share) CopyTo() (ret consensus.CItem) {
	share := &Share{
		PKr:             s.PKr,
		VotePKr:         s.VotePKr,
		TransactionHash: s.TransactionHash,
		PoolId:          s.PoolId,
		Value:           new(big.Int).Set(s.Value),
		BlockNumber:     s.BlockNumber,
		InitNum:         s.InitNum,
		Fee:             s.Fee,
		Num:             s.Num,
		WillVoteNum:     s.WillVoteNum,
		Status:          s.Status,
		ReturnAmount:    new(big.Int).Set(s.ReturnAmount),
		Profit:          new(big.Int).Set(s.Profit),
		LastPayTime:     s.LastPayTime,
	}
	return share
}

func (s *Share) CopyFrom(ret consensus.CItem) {
	obj := ret.(*Share)
	s.PKr = obj.PKr
	s.VotePKr = obj.VotePKr
	s.TransactionHash = obj.TransactionHash
	s.PoolId = obj.PoolId
	s.BlockNumber = obj.BlockNumber
	s.Value = new(big.Int).Set(obj.Value)
	s.Fee = obj.Fee
	s.InitNum = obj.InitNum
	s.Num = obj.Num
	s.WillVoteNum = obj.WillVoteNum
	s.Status = obj.Status
	s.ReturnAmount = new(big.Int).Set(obj.ReturnAmount)
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

type StakePool struct {
	PKr             keys.PKr
	VotePKr         keys.PKr
	BlockNumber     uint64
	TransactionHash common.Hash
	Amount          *big.Int `rlp:"nil"`
	Fee             uint16

	CurrentShareNum uint32
	WishVoteNum     uint32
	ChoicedShareNum uint32
	MissedVoteNum   uint32
	ExpireNum       uint32

	Profit      *big.Int `rlp:"nil"`
	LastPayTime uint64
	Closed      bool
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
		s.BlockNumber,
		s.TransactionHash,
		s.Amount,
		s.Fee,
		s.CurrentShareNum,
		s.WishVoteNum,
		s.ChoicedShareNum,
		s.MissedVoteNum,
		s.ExpireNum,
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
		BlockNumber:     s.BlockNumber,
		TransactionHash: s.TransactionHash,
		Amount:          new(big.Int).Set(s.Amount),
		Fee:             s.Fee,

		CurrentShareNum: s.CurrentShareNum,
		WishVoteNum:     s.CurrentShareNum,
		ChoicedShareNum: s.ChoicedShareNum,
		MissedVoteNum:   s.MissedVoteNum,
		ExpireNum:       s.ExpireNum,

		Profit:      new(big.Int).Set(s.Profit),
		LastPayTime: s.LastPayTime,
		Closed:      s.Closed,
	}
}

func (s *StakePool) CopyFrom(ret consensus.CItem) {
	obj := ret.(*StakePool)
	s.PKr = obj.PKr
	s.VotePKr = obj.VotePKr
	s.BlockNumber = obj.BlockNumber
	s.TransactionHash = obj.TransactionHash
	s.Amount = new(big.Int).Set(obj.Amount)
	s.Fee = obj.Fee
	s.CurrentShareNum = obj.CurrentShareNum
	s.WishVoteNum = obj.WishVoteNum
	s.ChoicedShareNum = obj.ChoicedShareNum
	s.MissedVoteNum = obj.MissedVoteNum
	s.ExpireNum = obj.ExpireNum
	s.Profit = new(big.Int).Set(obj.Profit)
	s.LastPayTime = obj.LastPayTime
	s.Closed = obj.Closed
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
	GetBlock(hash common.Hash, number uint64) *types.Block
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
	missedNum    consensus.KVPoint
}

var (
	ShareDB      = consensus.DBObj{"STAKE$SHARE$"}
	StakePoolDB  = consensus.DBObj{"STAKE$POOL$"}
	missedNumKey = []byte("missednum")
	blockVoteDB  = consensus.DBObj{"STAKE$BLOCKVOTES$"}
)

func NewStakeState(statedb *state.StateDB) *StakeState {
	cons := statedb.GetStakeCons()

	stakeState := &StakeState{statedb: statedb}
	stakeState.missedNum = consensus.NewKVPt(cons, "STAKE$EMISSEDNNUM$", "")
	stakeState.sharePool = consensus.NewKVPt(cons, "STAKE$SHAREPOOL$CONS$", "")
	stakeState.shareObj = consensus.NewObjPt(cons, "STAKE$SHAREOBJ$CONS", ShareDB.Pre, "share")
	stakeState.stakePoolObj = consensus.NewObjPt(cons, "STAKE$POOL$CONS", StakePoolDB.Pre, "pool")

	return stakeState
}

func (self *StakeState) SetStakeState(key common.Hash, value common.Hash) {
	self.sharePool.SetValue(key[:], value[:])
}

func (self *StakeState) GetStakeState(key common.Hash) common.Hash {
	return common.BytesToHash(self.sharePool.GetValue(key.Bytes()))
}

func (self *StakeState) AddShare(share *Share) {
	//tree := NewTree(self)
	//tree.insert(&SNode{key: common.BytesToHash(share.Id()), num: share.InitNum})
	share.Status = STATUS_VALID
	share.Num = share.InitNum
	if share.Profit == nil {
		share.Profit = new(big.Int)
	}
	if share.ReturnAmount == nil {
		share.ReturnAmount = new(big.Int)
	}
	self.updateShare(share)
}

func (self *StakeState) updateShare(share *Share) {
	self.shareObj.AddObj(share)
}

func (self *StakeState) UpdateStakePool(pool *StakePool) {
	self.stakePoolObj.AddObj(pool)
}

func (self *StakeState) IsEffect(currentBlockNumber uint64) bool {
	tree := NewTree(self)
	if seroparam.Is_Dev() {
		return self.ShareSize() > 2
	} else {
		missedNum := decodeNumber32(self.missedNum.GetValue(missedNumKey))
		seletedNum := (currentBlockNumber - seroparam.SIP3()) * 3
		if seletedNum == 0 {
			return false
		}
		ratio := float64(missedNum) / float64(seletedNum)
		if ratio > 0.3 || tree.size() < 20000 {
			return false
		}
		return true
	}
}

func (self *StakeState) ShareSize() uint32 {
	tree := NewTree(self)
	return tree.size()
}

func (self *StakeState) SeleteShare(seed common.Hash) (ints []uint32, shares []*Share, err error) {
	tree := NewTree(self)
	tree.MiddleOrder()

	ints, err = FindShareIdxs(tree.size(), 3, NewHash256PRNG(seed[:]))
	if err != nil {
		return
	}
	for _, i := range ints {
		node := tree.findByIndex(uint32(i))
		share := self.GetShare(node.key)
		if share == nil {
			err = errors.New("not found share by index")
			return
		}
		shares = append(shares, share)
	}
	return
}

func (self *StakeState) GetShare(key common.Hash) *Share {
	item := self.shareObj.GetObj(key.Bytes(), &Share{})
	if item == nil {
		return nil
	}
	return item.(*Share)
}

func (self *StakeState) getShare(key common.Hash) *Share {
	share := self.GetShare(key)
	if share == nil {
		log.Crit("ProcessBeforeApply: Get share error", "key", common.BytesToHash(key[:]))
	}
	return share
}

func (self *StakeState) GetStakePool(poolId common.Hash) *StakePool {
	item := self.stakePoolObj.GetObj(poolId.Bytes(), &StakePool{})
	if item == nil {
		return nil
	}
	return item.(*StakePool)
}

func (self *StakeState) getStakePool(poolId common.Hash) *StakePool {
	pool := self.GetStakePool(poolId)
	if pool == nil {
		log.Crit("ProcessBeforeApply: Get stakePoolObj error", "poolId", common.BytesToHash(poolId[:]))
	}
	return pool
}

func GetBlockRecords(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share, pools []*StakePool) {
	if blockNumber == 52 {
		i := 0
		i++
	}
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

func (self *StakeState) getBlockRecords(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share, pools []*StakePool) {
	records := state.StakeDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				shares = append(shares, self.getShare(key))
			}
		}
		if record.Name == "pool" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				pools = append(pools, self.getStakePool(key))
			}
		}
	}
	return
}

func GetShare(getter serodb.Getter, hash common.Hash) *Share {
	ret := ShareDB.GetObject(getter, hash[:], &Share{})
	return ret.(*Share)
}

func GetSharesByBlock(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share) {
	records := state.StakeDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				ret := ShareDB.GetObject(getter, each.Hash, &Share{})
				shares = append(shares, ret.(*Share))
			}
		}

	}
	return
}

func (self *StakeState) getShares(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share) {
	records := state.StakeDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				shares = append(shares, self.getShare(key))
			}
		}

	}
	return
}

var (
	basePrice = big.NewInt(2000000000000000000)
	addition  = big.NewInt(368891382302157)

	baseReware = big.NewInt(2330000000000000000)
	rewareStep = big.NewInt(11022927689594)

	maxPrice = big.NewInt(5930000000000000000)

	outOfDateWindow      = uint64(544320)
	missVotedWindow      = uint64(725760)
	payWindow            = uint64(42336)
	statisticsMissWindow = uint64(6048)
)

func getStatisticsMissWindow() uint64 {
	if seroparam.Is_Dev() {
		return 10
	}
	return statisticsMissWindow
}

func getOutOfDateWindow() uint64 {
	if seroparam.Is_Dev() {
		return 100
	}
	return outOfDateWindow
}

func getMissVotedWindow() uint64 {
	if seroparam.Is_Dev() {
		return 105
	}
	return missVotedWindow
}

func getPayPeriod() uint64 {
	if seroparam.Is_Dev() {
		return 5
	}
	return payWindow
}

func (self *StakeState) CurrentPrice() *big.Int {
	tree := NewTree(self)
	return new(big.Int).Add(basePrice, new(big.Int).Mul(addition, big.NewInt(int64(tree.size()))))
}

func sum(basePrice, addition *big.Int, n int64) *big.Int {
	return new(big.Int).Add(new(big.Int).Mul(basePrice, big.NewInt(n)), new(big.Int).Div(new(big.Int).Mul(new(big.Int).Mul(big.NewInt(n), big.NewInt(n-1)), addition), big.NewInt(2)))
}

func (self *StakeState) CaleAvgPrice(amount *big.Int) (uint32, *big.Int, *big.Int) {
	basePrice := self.CurrentPrice()
	left := int64(1)
	right := new(big.Int).Div(amount, basePrice).Int64()
	if right <= 1 {
		return uint32(right), basePrice, basePrice
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
	return uint32(left), new(big.Int).Div(sumAmount, big.NewInt(left)), basePrice
}

func (self *StakeState) StakeCurrentReward() (*big.Int, *big.Int) {
	if seroparam.Is_Dev() {
		return big.NewInt(600000000000000000), big.NewInt(900000000000000000)
	}

	size := NewTree(self).size()
	soleAmount := new(big.Int).Add(baseReware, new(big.Int).Mul(rewareStep, big.NewInt(int64(size))))

	if soleAmount.Cmp(maxPrice) > 1 {
		soleAmount = new(big.Int).Set(maxPrice)
	}
	return soleAmount, new(big.Int).Div(new(big.Int).Mul(soleAmount, big.NewInt(3)), big.NewInt(2))
}

func (self *StakeState) CheckVotes(block *types.Block, bc blockChain) error {
	header := block.Header()
	if len(header.CurrentVotes) > 3 || len(header.ParentVotes) > 3 {
		return errors.New("header.CurrentVotes.len > 3 or header.ParentVotes > 3")
	}

	if self.IsEffect(block.NumberU64()) {
		if len(header.CurrentVotes) < 2 || len(header.CurrentVotes) > 3 {
			return errors.New("header.CurrentVotes.len < 2 or header.ParentVotes > 3")
		}
	}

	if len(header.CurrentVotes) > 0 {
		// check repeated vote
		voteMap := map[keys.Uint512]types.HeaderVote{}
		for _, vote := range header.CurrentVotes {
			voteMap[vote.Sign] = vote
		}
		if len(voteMap) != len(header.CurrentVotes) {
			return errors.New("vote sign repeated")
		}

		//check shareIds
		shareMapNum := map[common.Hash]uint8{}
		tree := NewTree(self)
		seed := header.HashPos()
		indexs, err := FindShareIdxs(tree.size(), 3, NewHash256PRNG(seed[:]))
		if err == nil {
			for _, index := range indexs {
				sndoe := tree.findByIndex(index)
				shareMapNum[sndoe.key] += 1
			}
		}

		for _, vote := range header.CurrentVotes {
			if shareMapNum[vote.Id] != 0 {
				shareMapNum[vote.Id] -= 1
			} else {
				return errors.New("vote error")
			}
		}
	}

	if len(header.ParentVotes) == 1 {
		block := bc.GetBlock(header.ParentHash, header.Number.Uint64()-1)
		preHeader := block.Header()
		shareMapNum := map[common.Hash]uint8{}
		voteHash := rawdb.ReadBlockVotes(bc.GetDB(), header.ParentHash)
		for _, key := range voteHash {
			shareMapNum[key] += 1
		}

		parentVoteMap := map[keys.Uint512]types.HeaderVote{}
		for _, vote := range preHeader.CurrentVotes {
			if shareMapNum[vote.Id] == 0 {
				return errors.New("vote error")
			}
			shareMapNum[vote.Id] -= 1
			parentVoteMap[vote.Sign] = vote
		}

		for _, vote := range header.ParentVotes {
			if _, ok := parentVoteMap[vote.Sign]; ok {
				return errors.New("exist in parent header votes")
			}
			if shareMapNum[vote.Id] == 0 {
				return errors.New("vote error")
			} else {
				shareMapNum[vote.Id] -= 1
			}
		}
	}
	return nil
}

func (self *StakeState) ProcessBeforeApply(bc blockChain, header *types.Header) {
	self.processVotedShare(header, bc)
	self.processOutDate(header, bc)
	self.processMissVoted(header, bc)
	self.processNowShares(header, bc)
	self.payProfit(bc, header)
	self.statisticsByWindow(header, bc)
}

func (self *StakeState) statisticsByWindow(header *types.Header, bc blockChain) {
	if header.Number.Uint64() < 2 || !self.IsEffect(header.Number.Uint64()) {
		return
	}
	value := self.missedNum.GetValue(missedNumKey)

	preHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	prepreHeader := bc.GetHeader(preHeader.ParentHash, preHeader.Number.Uint64()-1)
	missedNum := decodeNumber32(value) + uint32(3-len(prepreHeader.CurrentVotes)-len(preHeader.ParentVotes))

	statisticsMissWindow := getStatisticsMissWindow()
	if prepreHeader.Number.Uint64() > statisticsMissWindow {
		windiwHeader := bc.GetHeaderByNumber(preHeader.Number.Uint64() - statisticsMissWindow - 1)
		preWindiwHeader := bc.GetHeaderByNumber(prepreHeader.Number.Uint64() - statisticsMissWindow - 1)
		missedNum -= uint32(3 - len(preWindiwHeader.CurrentVotes) - len(windiwHeader.ParentVotes))
	}
	log.Info("ProcessBeforeApply: statisticsByWindow", "missedNum", missedNum)
	self.missedNum.SetValue(missedNumKey, encodeNumber32(missedNum))
}

func (self *StakeState) processVotedShare(header *types.Header, bc blockChain) {
	log.Info("ProcessBeforeApply: processVotedShare", "blockNumber", header.Number.Uint64())
	if header.Number.Uint64() == 1 {
		return
	}
	preHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)

	tree := NewTree(self)
	poshash := preHeader.HashPos()
	indexs, err := FindShareIdxs(tree.size(), 3, NewHash256PRNG(poshash[:]))
	log.Info("processVotedShare selete share", "poshash", poshash, "blockNumber", preHeader.Number.Uint64(), "indexs", indexs, "pool size", tree.size())
	if err == nil {
		ndoes := []*SNode{}
		for _, index := range indexs {
			sndoe := tree.findByIndex(index)
			ndoes = append(ndoes, sndoe)

			share := self.getShare(sndoe.key)
			share.WillVoteNum += 1
			if share.Num > 0 {
				log.Info("ProcessBeforeApply: share.Num-1", "share.Num", share.Num, "key", common.Bytes2Hex(share.Id()))
				share.Num -= 1
			} else {
				log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.Num=0")
			}

			if share.PoolId != nil {
				pool := self.getStakePool(*share.PoolId)
				pool.ChoicedShareNum += 1
				pool.MissedVoteNum += 1

				if pool.CurrentShareNum > 0 {
					log.Info("ProcessBeforeApply: pool.CurrentShareNum", "poolId", common.Bytes2Hex(pool.Id()), "pool.CurrentShareNum", pool.CurrentShareNum)
					pool.CurrentShareNum -= 1
				} else {
					log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.CurrentShareNum=0")
				}
				pool.WishVoteNum += 1
				self.UpdateStakePool(pool)
				log.Info("ProcessBeforeApply: pool", "poolId", common.Bytes2Hex(pool.Id()), "pool.CurrentShareNum", pool.CurrentShareNum, "pool.WillVoteNum", pool.WishVoteNum)
			}
			self.updateShare(share)
		}
		for _, node := range ndoes {
			tree.deleteNodeByHash(node.key, 1)
		}
	}

	soloReware, reward := self.StakeCurrentReward()
	if len(preHeader.CurrentVotes) > 0 {
		log.Info("ProcessBeforeApply: process vote CurrentVotes", "size", len(preHeader.CurrentVotes), "blockNumber", preHeader.Number.Uint64())
		log.Info("ProcessBeforeApply: currentReward", "soloReware", soloReware, "reward", reward)
		for _, vote := range preHeader.CurrentVotes {
			self.rewardVote(vote, soloReware, reward)
		}
	}

	if len(preHeader.ParentVotes) > 0 {
		log.Info("ProcessBeforeApply: process vote ParentVotes", "size", len(preHeader.ParentVotes), "blockNumber", preHeader.Number.Uint64())
		reward = new(big.Int).Sub(reward, new(big.Int).Div(reward, big.NewInt(3)))
		soloReware = new(big.Int).Sub(soloReware, new(big.Int).Div(soloReware, big.NewInt(3)))
		log.Info("ProcessBeforeApply: currentReward", "soloReware", soloReware, "reward", reward)
		for _, vote := range preHeader.ParentVotes {
			self.rewardVote(vote, soloReware, reward)
		}
	}

}

func (self *StakeState) rewardVote(vote types.HeaderVote, soloReware, reward *big.Int) {

	share := self.getShare(vote.Id)
	if share.WillVoteNum > 0 {
		share.WillVoteNum -= 1
		log.Info("processVotedShare rewardVote", "shareId", common.Bytes2Hex(share.Id()), "share.WillVoteNum", share.WillVoteNum)
	} else {
		log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.WillVoteNum=0")
	}

	if share.PoolId != nil {
		pool := self.getStakePool(*share.PoolId)
		if pool.WishVoteNum > 0 {
			pool.WishVoteNum -= 1
			if pool.Closed && pool.CurrentShareNum == 0 && pool.WishVoteNum == 0 {
				pool.AddProfit(pool.Amount)
			}
		} else {
			log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.WillVoteNum=0")
		}
		self.UpdateStakePool(pool)
	}

	if vote.IsPool {
		pool := self.getStakePool(*share.PoolId)
		if pool.MissedVoteNum > 0 {
			pool.MissedVoteNum -= 1
		} else {
			log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.MissedVoteNum=0")
		}
		poolProfit := new(big.Int).Div(new(big.Int).Mul(reward, big.NewInt(int64(share.Fee))), big.NewInt(10000))
		pool.AddProfit(poolProfit)
		share.AddProfit(new(big.Int).Add(share.Value, new(big.Int).Sub(reward, poolProfit)))
		self.UpdateStakePool(pool)
	} else {
		share.AddProfit(new(big.Int).Add(share.Value, soloReware))
		log.Info("processVotedShare rewardVote", "shareId", common.Bytes2Hex(share.Id()), "share.Value", share.Value, "soloReware", soloReware, "Profit", share.Profit)
	}
	self.updateShare(share)
}

func (self *StakeState) processOutDate(header *types.Header, bc blockChain) (shares []*Share) {
	outOfDatePeriod := getOutOfDateWindow()
	if header.Number.Uint64() < outOfDatePeriod {
		return
	}
	tree := NewTree(self)
	preHeader := bc.GetHeaderByNumber(header.Number.Uint64() - outOfDatePeriod)
	shares = self.getShares(bc.GetDB(), preHeader.Hash(), preHeader.Number.Uint64())
	if len(shares) > 0 {
		for _, share := range shares {
			if share.BlockNumber == preHeader.Number.Uint64() {
				if share.Status == STATUS_OUTOFDATE {
					continue
				}
				if share.Num == 0 {
					share.Status = STATUS_OUTOFDATE
					self.updateShare(share)
					continue
				}
				sndoe := tree.deleteNodeByHash(common.BytesToHash(share.Id()), share.Num)
				if sndoe == nil {
					log.Crit("ProcessBeforeApply: processOutDate share not found", "shareId", common.Bytes2Hex(share.Id()), "InitNum", share.InitNum, "Num", share.Num, "WillVoteNum", share.WillVoteNum)
				}
				if share.Num != sndoe.num {
					log.Crit("ProcessBeforeApply: processOutDate err", "share.num", share.Num, "snode.num", sndoe.num)
				}

				if share.PoolId != nil {
					pool := self.getStakePool(*share.PoolId)
					if pool.CurrentShareNum >= share.Num {
						pool.CurrentShareNum -= share.Num
						pool.ExpireNum += share.Num
					} else {
						log.Crit("ProcessBeforeApply: processOutDate err", "pool.CurrentShareNum", pool.CurrentShareNum, "share.num", share.Num)
					}
					if pool.Closed && pool.CurrentShareNum == 0 && pool.WishVoteNum == 0 {
						pool.AddProfit(pool.Amount)
						log.Info("ProcessBeforeApply: processOutDate close pool", "poolId", common.Bytes2Hex(share.Id()))
					}
					self.UpdateStakePool(pool)
				}

				log.Info("ProcessBeforeApply: processOutDate set share.Num = 0", "shareId", common.Bytes2Hex(share.Id()), "share.num", share.Num)
				share.AddProfit(new(big.Int).Mul(share.Value, big.NewInt(int64(share.Num))))
				share.Status = STATUS_OUTOFDATE
				self.updateShare(share)
			}
		}
	}
	return
}

func (self *StakeState) processMissVoted(header *types.Header, bc blockChain) {
	missVotedPeriod := getMissVotedWindow()
	if header.Number.Uint64() < missVotedPeriod {
		return
	}
	perHeader := bc.GetHeaderByNumber(header.Number.Uint64() - missVotedPeriod)
	shares := self.getShares(bc.GetDB(), perHeader.Hash(), perHeader.Number.Uint64())
	if len(shares) > 0 {
		for _, share := range shares {
			if share.BlockNumber == perHeader.Number.Uint64() {
				if share.Status == STATUS_FINISHED {
					continue
				}
				if share.WillVoteNum == 0 {
					share.Status = STATUS_FINISHED
					self.updateShare(share)
					continue
				}

				//if share.Num != 0 {
				//	log.Crit("ProcessBeforeApply: processMissVoted err, share.Num!=0", "shareId", common.Bytes2Hex(share.Id()), "share.Num", share.Num)
				//}

				if share.PoolId != nil {
					pool := self.getStakePool(*share.PoolId)
					if pool.WishVoteNum < share.WillVoteNum {
						log.Crit("ProcessBeforeApply: processMissVoted err", "poolId", common.Bytes2Hex(share.Id()), "error", "pool.WishVoteNum < share.WillVoteNum")
					}
					pool.WishVoteNum -= share.WillVoteNum
					if pool.Closed && pool.CurrentShareNum == 0 && pool.WishVoteNum == 0 {
						pool.AddProfit(pool.Amount)
						log.Info("ProcessBeforeApply: processMissVoted close pool", "poolId", common.Bytes2Hex(share.Id()))
					}
					self.UpdateStakePool(pool)
				}

				share.AddProfit(new(big.Int).Mul(share.Value, big.NewInt(int64(share.WillVoteNum))))
				log.Info("ProcessBeforeApply: processMissVoted set share.Statue = STATUS_FINISHED", "shareId", common.Bytes2Hex(share.Id()), "share.WillVoteNum", share.WillVoteNum)
				share.Status = STATUS_FINISHED
				self.updateShare(share)
			}
		}
	}
}

func (self *StakeState) processNowShares(header *types.Header, bc blockChain) {
	perHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	shares := GetSharesByBlock(bc.GetDB(), perHeader.Hash(), perHeader.Number.Uint64())
	//shares := self.getShares(bc.GetDB(), perHeader.Hash(), perHeader.Number.Uint64(), shareCacheMap)
	if len(shares) > 0 {
		tree := NewTree(self)
		for _, share := range shares {
			if share.BlockNumber != perHeader.Number.Uint64() {
				continue
			}
			tree.insert(&SNode{key: common.BytesToHash(share.Id()), num: share.Num, total: share.Num, nodeNum: 1})
			if share.PoolId != nil {
				pool := self.getStakePool(*share.PoolId)
				pool.CurrentShareNum += share.Num
				self.UpdateStakePool(pool)
			}
		}
	}
}

func (self *StakeState) payProfit(bc blockChain, header *types.Header) {
	payPeriod := getPayPeriod()
	if header.Number.Uint64() < payPeriod {
		return
	}
	preHeader := bc.GetHeaderByNumber(header.Number.Uint64() - payPeriod)
	shares, pools := self.getBlockRecords(bc.GetDB(), preHeader.Hash(), preHeader.Number.Uint64())

	for _, share := range shares {
		if share.Profit.Sign() > 0 && header.Number.Uint64()-share.LastPayTime >= payPeriod {
			addr := common.Address{}
			copy(addr[:], share.PKr[:])
			asset := assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256(),
				Value:    utils.U256(*share.Profit),
			},
			}

			share.ReturnAmount.Add(share.ReturnAmount, share.Profit)
			log.Info("ProcessBeforeApply: payProfit rewardVote", "shareId", common.Bytes2Hex(share.Id()), "Profit", share.Profit)
			share.LastPayTime = header.Number.Uint64()
			share.SetProfit(big.NewInt(0))
			self.statedb.GetZState().AddTxOutWithCheck(addr, asset)
			self.updateShare(share)
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
			self.statedb.GetZState().AddTxOutWithCheck(addr, asset)
			self.UpdateStakePool(pool)
		}
	}
}

func decodeNumber32(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return binary.BigEndian.Uint32(data)
}

func encodeNumber32(number uint32) []byte {
	enc := make([]byte, 4)
	binary.BigEndian.PutUint32(enc, number)
	return enc
}
