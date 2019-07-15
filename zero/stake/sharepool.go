package stake

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sero-cash/go-sero/consensus/ethash"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
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
	s.LastPayTime = obj.LastPayTime
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

func (self *StakePool) Valid() bool {
	if !self.Closed {
		return true
	}
	if self.CurrentShareNum > 0 {
		return true
	}
	return false
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
		WishVoteNum:     s.WishVoteNum,
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
	//GetHeaderByNumber(number uint64) *types.Header
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(header *types.Header) (*state.StateDB, error)
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
	blockHash    consensus.KVPoint
}

var (
	ShareDB          = consensus.DBObj{"STAKE$SHARE$"}
	StakePoolDB      = consensus.DBObj{"STAKE$POOL$"}
	missedNumKey     = []byte("missednum")
	blockVotesPrefix = []byte("STAKE$BLOCKVOTES$")
)

type selectShare struct {
	Idx    []uint32
	Shares []common.Hash
}

func blockVotesKey(hash common.Hash) []byte {
	return append(blockVotesPrefix, hash[:]...)
}

func (self *StakeState) RecordVotes(batch serodb.Batch, block *types.Block) (error) {
	idx, shares, err := self.SeleteShare(block.HashPos())
	if err != nil {
		return err
	}
	ss := selectShare{}
	ss.Idx = idx
	for _, share := range shares {
		ss.Shares = append(ss.Shares, common.BytesToHash(share.State()))
	}

	data, err := rlp.EncodeToBytes(&ss)
	if err != nil {
		log.Crit("Failed to RLP encode blockVotes", "err", err)
	}

	if err := batch.Put(blockVotesKey(block.Hash()), data); err != nil {
		log.Crit("Failed to store blockVotes to number mapping", "err", err)
	}
	return nil
}

func SeleteBlockShare(getter serodb.Getter, block common.Hash) (idx []uint32, shares []*Share) {
	data, _ := getter.Get(blockVotesKey(block))
	if len(data) == 0 {
		return
	}
	ss := selectShare{}
	if err := rlp.Decode(bytes.NewReader(data), &ss); err != nil {
		log.Error("Invalid block blockVotes RLP", "hash", block, "err", err)
		return
	}
	idx = ss.Idx
	for _, hash := range ss.Shares {
		share := GetShare(getter, hash)
		if share == nil {
			log.Crit("Select Block Share Error: can not get share by hash", "hash", hash)
		}
		shares = append(shares, share)
	}
	return
}

func NewStakeState(statedb *state.StateDB) *StakeState {
	cons := statedb.GetStakeCons()
	stakeState := &StakeState{statedb: statedb}
	stakeState.missedNum = consensus.NewKVPt(cons, "STAKE$EMISSEDNNUM$", "")
	stakeState.sharePool = consensus.NewKVPt(cons, "STAKE$SHAREPOOL$CONS$", "")
	stakeState.shareObj = consensus.NewObjPt(cons, "STAKE$SHAREOBJ$CONS", ShareDB.Pre, "share")
	stakeState.stakePoolObj = consensus.NewObjPt(cons, "STAKE$POOL$CONS", StakePoolDB.Pre, "pool")
	stakeState.blockHash = consensus.NewKVPt(cons, "BLOCK$BLOCKHASH$", "")
	return stakeState
}

func (self *StakeState) setBlockHash(blockNumber uint64, blockHash common.Hash) {
	self.blockHash.SetValue(utils.EncodeNumber(blockNumber), blockHash[:])
}

func (self *StakeState) getBlockHash(blockNumber uint64) common.Hash {
	ret := self.blockHash.GetValue(utils.EncodeNumber(blockNumber))
	return common.BytesToHash(ret[:])
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

func (self *StakeState) NeedTwoVote(num uint64) bool {
	if seroparam.Is_Dev() {
		return self.ShareSize() > ValidVoteCount
	} else {
		window_size := getStatisticsMissWindow()
		if num > seroparam.SIP4()+window_size {
			missedNum := utils.DecodeNumber32(self.missedNum.GetValue(missedNumKey))
			seletedNum := window_size * MaxVoteCount
			ratio := float64(missedNum) / float64(seletedNum)
			if ratio > minMissRate || self.ShareSize() < minSharePoolSize {
				return false
			}
			return true
		} else {
			return false
		}
	}
}

func (self *StakeState) ShareSize() uint32 {
	tree := NewTree(self)
	return tree.size()
}

func (self *StakeState) SeleteShare(seed common.Hash) (ints []uint32, shares []*Share, err error) {
	tree := NewTree(self)
	tree.MiddleOrder()

	if tree.size() < 3 {
		return
	}

	ints, err = FindShareIdxs(tree.size(), 3, NewHash256PRNG(seed[:]))
	if err != nil {
		return
	}
	for _, i := range ints {
		node, e := tree.findByIndex(uint32(i))
		if e != nil {
			err = e
			return
		}
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

/*func (self *StakeState) getShare(key common.Hash) *Share {
	share := self.GetShare(key)
	if share == nil {
		log.Crit("getShare error", "key", common.BytesToHash(key[:]))
	}
	return share
}*/

func (self *StakeState) GetStakePool(poolId common.Hash) *StakePool {
	item := self.stakePoolObj.GetObj(poolId.Bytes(), &StakePool{})
	if item == nil {
		return nil
	}
	return item.(*StakePool)
}

/*func (self *StakeState) getStakePool(poolId common.Hash) *StakePool {
	pool := self.GetStakePool(poolId)
	if pool == nil {
		log.Crit("getStakePool error", "poolId", common.BytesToHash(poolId[:]))
	}
	return pool
}*/

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
				pool := ret.(*StakePool)
				log.Info("get blocks", "id", hexutil.Encode(each.Ref), "hash", hexutil.Encode(each.Hash), "retHash", hexutil.Encode(pool.State()))
				pools = append(pools, ret.(*StakePool))
			}
		}
	}
	return
}

func (self *StakeState) getBlockRecords(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share, pools []*StakePool, err error) {
	records := state.StakeDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				share := self.GetShare(key)
				if share == nil {
					err = errors.New("not found share by shareId")
					return
				}
				shares = append(shares, share)
			}
		}
		if record.Name == "pool" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				pool := self.GetStakePool(key)
				if pool == nil {
					err = errors.New("not found pool by poolId")
				}
				pools = append(pools, pool)
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

func (self *StakeState) getShares(getter serodb.Getter, blockHash common.Hash, blockNumber uint64) (shares []*Share, err error) {
	records := state.StakeDB.GetBlockRecords(getter, blockNumber, &blockHash)
	for _, record := range records {
		if record.Name == "share" {
			for _, each := range record.Pairs {
				key := common.BytesToHash(each.Ref)
				share := self.GetShare(key)
				if share == nil {
					err = errors.New("not found share by shareId")
					return
				}
				shares = append(shares, share)
			}
		}

	}
	return
}

func (self *StakeState) CurrentPrice() *big.Int {
	tree := NewTree(self)
	return new(big.Int).Add(basePrice, new(big.Int).Mul(addition, big.NewInt(int64(tree.size()))))
}

func SumAmount(basePrice, addition *big.Int, n int64) *big.Int {
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
		sumAmount := SumAmount(basePrice, addition, mid)
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
	sumAmount := SumAmount(basePrice, addition, left)
	if sumAmount.Cmp(amount) > 0 {
		left -= 1
		sumAmount = SumAmount(basePrice, addition, left)
	}
	return uint32(left), new(big.Int).Div(sumAmount, big.NewInt(left)), basePrice
}

func (self *StakeState) StakeCurrentReward(blockNumber *big.Int) (soloRewards *big.Int, totalRewards *big.Int) {
	if seroparam.Is_Dev() {
		return big.NewInt(600000000000000000), big.NewInt(900000000000000000)
	}

	size := NewTree(self).size()
	totalReward := new(big.Int).Add(baseReware, new(big.Int).Mul(rewareStep, big.NewInt(int64(size))))

	if totalReward.Cmp(maxReware) > 0 {
		totalReward = new(big.Int).Set(maxReware)
	}
	halve := ethash.Halve(blockNumber)
	fmt.Println(halve)

	soleAmount := new(big.Int).Div(new(big.Int).Mul(totalReward, big.NewInt(SOLO_RATE)), big.NewInt(TOTAL_RATE))
	return new(big.Int).Div(soleAmount, halve), new(big.Int).Div(totalReward, halve)
}

func (self *StakeState) checkShareRepeated(header *types.Header) error {
	tree := NewTree(self)
	seed := header.HashPos()
	indexs, err := FindShareIdxs(tree.size(), 3, NewHash256PRNG(seed[:]))
	if err != nil {
		return err
	}

	voteNumMap := map[common.Hash]int{}
	for _, index := range indexs {
		sndoe, err := tree.findByIndex(index)
		if err != nil {
			return err
		}
		voteNumMap[sndoe.key] += 1
	}
	for _, vote := range header.CurrentVotes {
		if num, ok := voteNumMap[vote.Id]; ok && num > 0 {
			voteNumMap[vote.Id] -= 1
		} else {
			return errors.New("vote repeated")
		}
	}
	return nil
}

func (self *StakeState) CheckVotes(block *types.Block, bc blockChain) error {

	header := block.Header()
	if len(header.CurrentVotes) > 3 || len(header.ParentVotes) > 3 {
		return errors.New("header.CurrentVotes.len > 3 or header.ParentVotes > 3")
	}

	if self.NeedTwoVote(block.NumberU64()) {
		if len(header.CurrentVotes) < 2 {
			return errors.New("header.CurrentVotes.len < 2")
		}
	}
	parentblock := bc.GetBlock(header.ParentHash, header.Number.Uint64()-1)
	if len(header.CurrentVotes) > 0 {
		// check repeated vote
		parentPosHash := parentblock.HashPos()
		blockPosHash := block.HashPos()
		ret := types.StakeHash(&blockPosHash, &parentPosHash)
		voteNumMap := map[keys.Uint512]bool{}
		for _, vote := range header.CurrentVotes {
			if err := self.verifyVote(vote, ret); err != nil {
				return err
			}
			voteNumMap[vote.Sign] = true
		}

		if len(voteNumMap) != len(header.CurrentVotes) {
			return errors.New("vote sign repeated")
		}

		if err := self.checkShareRepeated(header); err != nil {
			return err
		}
	}

	if len(header.ParentVotes) > 0 {
		preHeader := parentblock.Header()
		shareMapNum := map[common.Hash]int{}
		_, shares := SeleteBlockShare(bc.GetDB(), header.ParentHash)
		for _, share := range shares {
			shareMapNum[common.BytesToHash(share.Id())] += 1
		}

		parentVoteMap := map[keys.Uint512]types.HeaderVote{}
		for _, vote := range preHeader.CurrentVotes {
			if shareMapNum[vote.Id] == 0 {
				return errors.New("vote error")
			}
			shareMapNum[vote.Id] -= 1
			parentVoteMap[vote.Sign] = vote
		}

		perperBlock := bc.GetBlock(parentblock.ParentHash(), parentblock.NumberU64()-1)
		parentPosHash := perperBlock.HashPos()
		blockPosHash := parentblock.HashPos()
		ret := types.StakeHash(&blockPosHash, &parentPosHash)
		for _, vote := range header.ParentVotes {
			if err := self.verifyVote(vote, ret); err != nil {
				return err
			}

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

func (self *StakeState) verifyVote(vote types.HeaderVote, stakeHash common.Hash) error {
	share := self.GetShare(vote.Id)
	if share == nil {
		return errors.New("not found share by shareId")
	}
	if vote.IsPool {
		pool := self.GetStakePool(*share.PoolId)
		if pool == nil {
			return errors.New("not found pool by poolId")
		}
		if !keys.VerifyPKr(stakeHash.HashToUint256(), &vote.Sign, &pool.VotePKr) {
			return errors.New("Verify header votes error")
		}
	} else {
		if !keys.VerifyPKr(stakeHash.HashToUint256(), &vote.Sign, &share.VotePKr) {
			return errors.New("Verify header votes error")
		}
	}
	return nil

}

func (self *StakeState) processRemedyRewards(bc blockChain, header *types.Header) {
	if header.Number.Uint64() > 0 {
		parentHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)
		if len(parentHeader.ParentVotes) > 0 {
			soloReware, totalReward := self.StakeCurrentReward(parentHeader.Number)
			log.Info("accumulateRewards: currentReward", "soloReware", soloReware, "totalReward", totalReward, "header.ParentVotes", len(parentHeader.ParentVotes))
			reward := new(big.Int)
			for _, vote := range parentHeader.ParentVotes {
				if vote.IsPool {
					reward.Add(reward, new(big.Int).Div(totalReward, big.NewInt(3)))
				} else {
					reward.Add(reward, new(big.Int).Div(soloReware, big.NewInt(3)))
				}
			}
			asset := assets.Asset{
				&assets.Token{
					utils.CurrencyToUint256("SERO"),
					utils.U256(*reward),
				},
				nil,
			}
			log.Info("pos to pow reward:", "coinbase", parentHeader.Coinbase.String(), "value", reward, "at", parentHeader.Number.Uint64())
			self.statedb.NextZState().AddTxOut(parentHeader.Coinbase, asset)
		}
	}
}

func (self *StakeState) ProcessBeforeApply(bc blockChain, header *types.Header) (err error) {
	self.setBlockHash(header.Number.Uint64()-1, header.ParentHash)
	// last round
	err = self.processVotedShare(header, bc)
	if err != nil {
		return err
	}
	err = self.processOutDate(header, bc)
	if err != nil {
		return err
	}
	err = self.processMissVoted(header, bc)
	if err != nil {
		return err
	}
	// last round buy share
	err = self.processNowShares(header, bc)
	if err != nil {
		return err
	}
	// to last round circyle payment
	err = self.payProfit(bc, header)
	if err != nil {
		return err
	}
	// last block pow remedy rewards
	self.processRemedyRewards(bc, header)
	// last block statistics
	self.statisticsByWindow(header, bc)
	return
}

func (self *StakeState) statisticsByWindow(header *types.Header, bc blockChain) {
	statisticsMissWindow := getStatisticsMissWindow()
	if header.Number.Uint64() < 1 || header.Number.Uint64()-1 < seroparam.SIP4() {
		return
	}
	value := self.missedNum.GetValue(missedNumKey)

	preHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	totalVote := len(preHeader.CurrentVotes) + len(preHeader.ParentVotes)
	missedNum := utils.DecodeNumber32(value)
	if totalVote < 3 {
		missedNum += uint32(3 - totalVote)
	}

	if preHeader.Number.Uint64() >= seroparam.SIP4()+statisticsMissWindow {
		preNumber := preHeader.Number.Uint64() - statisticsMissWindow
		windiwHeader := bc.GetHeader(self.getBlockHash(preNumber), preNumber)
		totalWVote := len(windiwHeader.CurrentVotes) + len(windiwHeader.ParentVotes)
		var delNum uint32
		if totalWVote < 3 {
			delNum = uint32(3 - totalWVote)
		}
		if missedNum < delNum {
			log.Crit("ProcessBeforeApply: statisticsByWindow err", "missedNum", missedNum)
		} else {
			missedNum -= delNum
		}
	}
	log.Info("ProcessBeforeApply: statisticsByWindow", "missedNum", missedNum)
	self.missedNum.SetValue(missedNumKey, utils.EncodeNumber32(missedNum))
}

func (self *StakeState) processVotedShare(header *types.Header, bc blockChain) (err error) {
	log.Info("ProcessBeforeApply: processVotedShare", "blockNumber", header.Number.Uint64())
	if header.Number.Uint64() == 1 || header.Number.Uint64()-1 < seroparam.SIP4() {
		return
	}
	preHeader := bc.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	tree := NewTree(self)
	poshash := preHeader.HashPos()
	indexs, e := FindShareIdxs(tree.size(), 3, NewHash256PRNG(poshash[:]))
	log.Info("processVotedShare selete share", "poshash", poshash, "blockNumber", preHeader.Number.Uint64(), "indexs", indexs, "pool size", tree.size())
	if e == nil {
		ndoes := []*SNode{}
		for _, index := range indexs {
			sndoe, e := tree.findByIndex(index)
			if e != nil {
				err = e
				return
			}
			ndoes = append(ndoes, sndoe)

			share := self.GetShare(sndoe.key)
			if share == nil {
				err = errors.New("not found share by shareId")
				return
			}
			share.WillVoteNum += 1
			if share.Num > 0 {
				log.Info("ProcessBeforeApply: share.Num-1", "share.Num", share.Num, "key", common.Bytes2Hex(share.Id()))
				share.Num -= 1
			} else {
				return errors.New(fmt.Sprint("ProcessBeforeApply: process vote err ", "shareId=", common.Bytes2Hex(share.Id()), " error=", "share.Num=0"))
				//log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.Num=0")
			}
			self.updateShare(share)

			if share.PoolId != nil {
				pool := self.GetStakePool(*share.PoolId)
				if pool == nil {
					err = errors.New("not found pool by poolId")
					return
				}
				pool.ChoicedShareNum += 1
				pool.MissedVoteNum += 1

				if pool.CurrentShareNum > 0 {
					log.Info("ProcessBeforeApply: pool.CurrentShareNum", "poolId", common.Bytes2Hex(pool.Id()), "pool.CurrentShareNum", pool.CurrentShareNum)
					pool.CurrentShareNum -= 1
				} else {
					err = errors.New(fmt.Sprint("ProcessBeforeApply: process vote err", " poolId=", share.PoolId, " error=", "pool.CurrentShareNum=0"))
					return
					//log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.CurrentShareNum=0")
				}
				pool.WishVoteNum += 1
				self.UpdateStakePool(pool)
				log.Info("ProcessBeforeApply: pool", "poolId", common.Bytes2Hex(pool.Id()), "pool.CurrentShareNum", pool.CurrentShareNum, "pool.WillVoteNum", pool.WishVoteNum)
			}
		}
		for _, node := range ndoes {
			tree.deleteNodeByHash(node.key, 1)
		}
	}

	soloReware, reward := self.StakeCurrentReward(preHeader.Number)
	if len(preHeader.CurrentVotes) > 0 {
		//log.Info("ProcessBeforeApply: process vote CurrentVotes", "size", len(preHeader.CurrentVotes), "blockNumber", preHeader.Number.Uint64())
		log.Info("ProcessBeforeApply: currentReward", "soloReware", soloReware, "reward", reward)
		for _, vote := range preHeader.CurrentVotes {
			err = self.rewardVote(vote, soloReware, reward, preHeader.Number.Uint64())
			if err != nil {
				return
			}
		}
	}

	if len(preHeader.ParentVotes) > 0 {
		//log.Info("ProcessBeforeApply: process vote ParentVotes", "size", len(preHeader.ParentVotes), "blockNumber", preHeader.Number.Uint64())
		reward = new(big.Int).Sub(reward, new(big.Int).Div(reward, big.NewInt(3)))
		soloReware = new(big.Int).Sub(soloReware, new(big.Int).Div(soloReware, big.NewInt(3)))
		log.Info("ProcessBeforeApply for ParentVotes currentReward", "soloReware", soloReware, "reward", reward)
		for _, vote := range preHeader.ParentVotes {
			err = self.rewardVote(vote, soloReware, reward, preHeader.Number.Uint64())
			if err != nil {
				return
			}
		}
	}
	return nil
}

func (self *StakeState) rewardVote(vote types.HeaderVote, soloReware, reward *big.Int, block uint64) error {

	share := self.GetShare(vote.Id)
	if share == nil {
		return errors.New("not found share by shareId")
	}
	if share.WillVoteNum > 0 {
		share.WillVoteNum -= 1
		log.Info("processVotedShare rewardVote", "shareId", common.Bytes2Hex(share.Id()), "share.WillVoteNum", share.WillVoteNum)
	} else {
		return errors.New(fmt.Sprint("ProcessBeforeApply: process vote err", " shareId=", common.Bytes2Hex(share.Id()), " error=", "share.WillVoteNum=0"))
		//log.Crit("ProcessBeforeApply: process vote err", "shareId", common.Bytes2Hex(share.Id()), "error", "share.WillVoteNum=0")
	}

	if share.PoolId != nil {
		pool := self.GetStakePool(*share.PoolId)
		if pool == nil {
			return errors.New("not found pool by poolId")
		}
		if pool.WishVoteNum > 0 {
			pool.WishVoteNum -= 1
			if pool.Closed && pool.CurrentShareNum == 0 && pool.WishVoteNum == 0 {
				pool.AddProfit(pool.Amount)
				pool.Amount = new(big.Int)
			}
		} else {
			return errors.New(fmt.Sprint("ProcessBeforeApply: process vote err", " poolId=", share.PoolId, " error=", "pool.WillVoteNum=0"))
			//log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.WillVoteNum=0")
		}
		self.UpdateStakePool(pool)
	}

	if vote.IsPool {
		pool := self.GetStakePool(*share.PoolId)
		if pool == nil {
			return errors.New("not found pool by poolId")
		}
		if pool.MissedVoteNum > 0 {
			pool.MissedVoteNum -= 1
		} else {
			return errors.New(fmt.Sprint("ProcessBeforeApply: process vote err", " poolId=", share.PoolId, " error=", "pool.MissedVoteNum=0"))
			//log.Crit("ProcessBeforeApply: process vote err", "poolId", share.PoolId, "error", "pool.MissedVoteNum=0")
		}
		poolProfit := new(big.Int).Div(new(big.Int).Mul(reward, big.NewInt(int64(share.Fee))), big.NewInt(10000))
		log.Info("pos to pool fee", "pool", common.BytesToHash(pool.Id()), "reward", poolProfit, "at", block)
		pool.AddProfit(poolProfit)
		share.AddProfit(new(big.Int).Add(share.Value, new(big.Int).Sub(reward, poolProfit)))
		log.Info("pos to share reward", "pool", common.BytesToHash(pool.Id()), "share", common.BytesToHash(share.Id()), "reward", new(big.Int).Sub(reward, poolProfit), "at", block)
		self.UpdateStakePool(pool)
	} else {
		share.AddProfit(new(big.Int).Add(share.Value, soloReware))
		log.Info("pos to share reward", "share", common.BytesToHash(share.Id()), "reward", soloReware, "at", block)
		//log.Info("processVotedShare rewardVote", "shareId", common.Bytes2Hex(share.Id()), "share.Value", share.Value, "soloReware", soloReware, "Profit", share.Profit)
	}
	self.updateShare(share)
	return nil
}

func (self *StakeState) processOutDate(header *types.Header, bc blockChain) (err error) {
	outOfDatePeriod := getOutOfDateWindow()
	if header.Number.Uint64() < outOfDatePeriod || header.Number.Uint64()-outOfDatePeriod < seroparam.SIP4() {
		return
	}

	preNumber := header.Number.Uint64() - outOfDatePeriod
	if preNumber < seroparam.SIP4() {
		return
	}

	preHeader := bc.GetHeader(self.getBlockHash(preNumber), preNumber)
	shares, err := self.getShares(bc.GetDB(), preHeader.Hash(), preHeader.Number.Uint64())
	if err != nil {
		return
	}
	if len(shares) > 0 {
		tree := NewTree(self)
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
					err = errors.New("processOutDate share not found")
					return
					//log.Crit("ProcessBeforeApply: processOutDate share not found", "shareId", common.Bytes2Hex(share.Id()), "InitNum", share.InitNum, "Num", share.Num, "WillVoteNum", share.WillVoteNum)
				}
				if share.Num != sndoe.num {
					err = errors.New(fmt.Sprint("ProcessBeforeApply: processOutDate err", " share.num=", share.Num, " snode.num=", sndoe.num))
					return
				}

				if share.PoolId != nil {
					pool := self.GetStakePool(*share.PoolId)
					if pool == nil {
						err = errors.New("not found pool by poolId")
						return
					}
					if pool.CurrentShareNum >= share.Num {
						pool.CurrentShareNum -= share.Num
						pool.ExpireNum += share.Num
					} else {
						err = errors.New(fmt.Sprint("ProcessBeforeApply: processOutDate err", " pool.CurrentShareNum ", pool.CurrentShareNum, " share.num=", share.Num))
						return
						//log.Crit("ProcessBeforeApply: processOutDate err", "pool.CurrentShareNum", pool.CurrentShareNum, "share.num", share.Num)
					}
					if pool.Closed && pool.CurrentShareNum == 0 && pool.WishVoteNum == 0 {
						pool.AddProfit(pool.Amount)
						pool.Amount = new(big.Int)
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

func (self *StakeState) processMissVoted(header *types.Header, bc blockChain) (err error) {
	missVotedPeriod := getMissVotedWindow()
	if header.Number.Uint64() < missVotedPeriod {
		return
	}
	preNumber := header.Number.Uint64() - missVotedPeriod
	if preNumber < seroparam.SIP4() {
		return
	}
	preHeader := bc.GetHeader(self.getBlockHash(preNumber), preNumber)
	shares, err := self.getShares(bc.GetDB(), preHeader.Hash(), preHeader.Number.Uint64())
	if err != nil {
		return
	}
	if len(shares) > 0 {
		for _, share := range shares {
			if share.BlockNumber == preHeader.Number.Uint64() {
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
					pool := self.GetStakePool(*share.PoolId)
					if pool == nil {
						err = errors.New("not found pool by poolId")
						return
					}
					if pool.WishVoteNum < share.WillVoteNum {
						err = errors.New(fmt.Sprint("ProcessBeforeApply: processMissVoted err", " poolId=", common.Bytes2Hex(share.Id()), " error=", "pool.WishVoteNum < share.WillVoteNum"))
						return
						//log.Crit("ProcessBeforeApply: processMissVoted err", "poolId", common.Bytes2Hex(share.Id()), "error", "pool.WishVoteNum < share.WillVoteNum")
					}
					pool.WishVoteNum -= share.WillVoteNum
					if pool.Closed && pool.CurrentShareNum == 0 && pool.WishVoteNum == 0 {
						pool.AddProfit(pool.Amount)
						pool.Amount = new(big.Int)
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
	return nil
}

func (self *StakeState) processNowShares(header *types.Header, bc blockChain) (err error) {
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
				pool := self.GetStakePool(*share.PoolId)
				if pool == nil {
					err = errors.New("not found pool by poolId")
					return
				}
				pool.CurrentShareNum += share.Num
				self.UpdateStakePool(pool)
			}
		}
	}
	return nil
}

func (self *StakeState) payProfit(bc blockChain, header *types.Header) (err error) {
	payPeriod := getPayPeriod()
	if header.Number.Uint64() < payPeriod {
		return
	}
	preNumber := header.Number.Uint64() - payPeriod
	if preNumber < seroparam.SIP4() {
		return
	}
	preHeader := bc.GetHeader(self.getBlockHash(preNumber), preNumber)
	shares, pools, err := self.getBlockRecords(bc.GetDB(), preHeader.Hash(), preHeader.Number.Uint64())
	if err != nil {
		return
	}

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
			self.statedb.NextZState().AddTxOut(addr, asset)
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
			self.statedb.NextZState().AddTxOut(addr, asset)
			self.UpdateStakePool(pool)
		}
	}
	return nil
}
