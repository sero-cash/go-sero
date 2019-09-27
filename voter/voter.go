package voter

import (
	"math/big"
	"sync"
	"time"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/accounts"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero/zero/stake"

	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/types"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/params"
)

const (
	evictionInterval = time.Minute
	chainLotterySize = 300
	lifeTime         = 30 * time.Minute

	delayNum         = 1
	lotteryQueueSize = 12
)

type blockChain interface {
	CurrentBlock() *types.Block
	GetHeaderByHash(hash common.Hash) *types.Header
	StateAt(header *types.Header) (*state.StateDB, error)
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	GetHeader(hash common.Hash, number uint64) *types.Header
	GetHeaderByNumber(number uint64) *types.Header
	GetBlock(hash common.Hash, number uint64) *types.Block
	GetDB() serodb.Database
}

type Voter struct {
	chain        blockChain
	sero         Backend
	lotteryCh    chan *types.Lottery
	voteFeed     event.Feed
	voteWorkFeed event.Feed
	lotteryFeed  event.Feed
	scope        event.SubscriptionScope
	//abi       types.Signer
	voteMu    sync.RWMutex
	lotteryMu sync.RWMutex

	votes    map[common.Hash]time.Time
	lotterys map[common.Hash]time.Time

	lotteryQueue *PriorityQueue
}

func NewVoter(chainconfig *params.ChainConfig, chain blockChain, sero Backend) *Voter {
	// Sanitize the input to ensure no vulnerable gas prices are set

	// Create the transaction pool with its initial settings
	voter := &Voter{
		sero:         sero,
		chain:        chain,
		lotteryCh:    make(chan *types.Lottery, chainLotterySize),
		votes:        make(map[common.Hash]time.Time),
		lotterys:     make(map[common.Hash]time.Time),
		lotteryQueue: &PriorityQueue{},
	}
	voter.lotteryQueue.Init(lotteryQueueSize)

	// Subscribe events from blockchain
	//voter.chainHeadSub = voter.chain.SubscribeChainHeadEvent(voter.chainHeadCh)

	// Start the event loop and return
	go voter.loop()
	go voter.lotteryTaskLoop()
	go voter.voteLoop()

	return voter
}

func (self *Voter) loop() {
	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()
	for {
		select {
		case <-evict.C:
			self.lotteryMu.Lock()

			dropLotterys := []common.Hash{}
			for k, v := range self.lotterys {
				if time.Since(v) > lifeTime {
					dropLotterys = append(dropLotterys, k)
				}
			}
			for _, h := range dropLotterys {
				delete(self.lotterys, h)
			}
			self.lotteryMu.Unlock()
			self.voteMu.Lock()
			dropVotes := []common.Hash{}
			for k, v := range self.votes {
				if time.Since(v) > lifeTime {
					dropVotes = append(dropVotes, k)
				}
			}
			for _, h := range dropVotes {
				delete(self.votes, h)
			}
			self.voteMu.Unlock()
		}
	}
}

func (self *Voter) IsLotteryValid(lottery *types.Lottery) bool {
	current := self.chain.CurrentBlock().NumberU64()
	if (lottery.ParentNum + 1) < current-1 {
		return false
	}
	if (lottery.ParentNum + 1) > current+2 {
		return false
	}
	return true
}

func (self *Voter) lotteryTaskLoop() {
	for {
		select {
		case lottery := <-self.lotteryCh:
			//current := self.chain.CurrentBlock().NumberU64()
			if !self.IsLotteryValid(lottery) {
				continue
			}
			//log.Info(">>>>>>>lotteryTaskLoop new lottery", "poshash", lottery.PosHash, "block", lottery.ParentNum+1, "localBlock", current)
			parentBlock := self.chain.GetBlock(lottery.ParentHash, lottery.ParentNum)
			if parentBlock == nil {
				log.Trace(">>>>>lotteryTaskLoop can not find parentblock", "parent block", lottery.ParentNum)
				self.lotteryQueue.PushItem(lottery.PosHash, &lotteryItem{Lottery: lottery, Attempts: uint8(0)}, lottery.ParentNum+1)
			} else {
				selfShares, err := self.SelfShares(lottery.PosHash, lottery.ParentHash, parentBlock.Number())
				if err != nil {
					log.Error("lotteryTaskLoop", "selfShare error ", err)
				} else {
					for _, s := range selfShares {
						go self.sign(s)
					}
				}
			}
		}
	}
}

func (self *Voter) voteLoop() {
	evict := time.NewTicker(time.Second)
	defer evict.Stop()
	for {
		select {
		case <-evict.C:
			current := self.chain.CurrentBlock().NumberU64()
			for item := self.lotteryQueue.Pop(); item != nil; item = self.lotteryQueue.Pop() {
				lItem := item.Value.(*lotteryItem)
				//log.Info(">>>>>voteLoop get Vote Item", "poshash", lItem.Lottery.PosHash, "block", lItem.Lottery.ParentNum+1)
				if current > lItem.Lottery.ParentNum+delayNum {
					log.Trace(">>>>>>not need vote", "current", current, "vote block", lItem.Lottery.ParentNum+1)
					continue
				}
				parentBlock := self.chain.GetBlock(lItem.Lottery.ParentHash, lItem.Lottery.ParentNum)
				if parentBlock == nil {
					log.Trace(">>>>>voteLoop get parent is nil", "hash", lItem.Lottery.ParentHash, "block", lItem.Lottery.ParentNum)
					self.lotteryQueue.PushItem(lItem.Lottery.PosHash, lItem, item.Block)
					break
				}
				selfShares, err := self.SelfShares(lItem.Lottery.PosHash, parentBlock.Hash(), parentBlock.Number())
				if err != nil {
					log.Trace("lotteryTaskLoop", "selfShare error ", err)
				} else {
					for _, s := range selfShares {
						self.sign(s)
					}
				}

			}
		}
	}
}

type voteInfo struct {
	index      uint32
	parentNum  uint64
	shareHash  common.Hash
	poshash    common.Hash
	statkeHash common.Hash
	votePKr    c_type.PKr
	isPool     bool
	seed       address.Seed
}

func cotainsVoteInfo(voteInfos []voteInfo, item voteInfo, pool *stake.StakePool) bool {
	if pool != nil && pool.Closed {
		return false
	}
	for _, v := range voteInfos {
		if v.seed == item.seed && v.index == item.index &&
			v.shareHash == v.shareHash && v.poshash == item.poshash &&
			v.parentNum == item.parentNum {
			return true
		}
	}
	return false
}

func pkrToAddress(pkr c_type.PKr) common.Address {
	var addr common.Address
	copy(addr[:], pkr[:])
	return addr
}

func GetSeedByVotePkr(wallets []accounts.Wallet, pkr c_type.PKr) *address.Seed {
	for _, w := range wallets {
		if w.IsMine(pkr) {
			seed, err := w.GetSeed()
			if err != nil {
				log.Trace("GetSeedByVotePkr", "err", err)
				return nil
			} else {
				return seed
			}
		}
	}
	return nil
}

func (self *Voter) SelfShares(poshash common.Hash, parent common.Hash, parentNumber *big.Int) ([]voteInfo, error) {
	current := self.chain.CurrentBlock().NumberU64()
	if current > delayNum+parentNumber.Uint64() {
		return nil, nil
	}
	parentHeader := self.chain.GetHeaderByHash(parent)
	if parentHeader == nil {
		return nil, nil
	}
	state, err := self.chain.StateAt(parentHeader)
	if err != nil {
		log.Trace("lotteryTaskLoop", "stateAt", poshash, "err", err)
		return nil, err
	} else {
		stakeState := stake.NewStakeState(state)

		header := &types.Header{
			ParentHash: parent,
			Number:     new(big.Int).Add(parentNumber, common.Big1),
		}
		err := stakeState.ProcessBeforeApply(self.chain, header)
		if err != nil {
			log.Error("lotteryTaskLoop", "err", err)
			return nil, nil
		}
		if stakeState.ShareSize() == 0 {
			return nil, nil
		}
		ints, shares, err := stakeState.SeleteShare(poshash)
		log.Info("SelfShares selete share", "poshash", poshash, "blockNumber", header.Number.Uint64(), "indexs", ints, "pool size", stakeState.ShareSize())
		if err != nil {
			log.Info("lotteryTaskLoop", "SeleteShare", poshash, "err", err)
			return nil, err
		}

		var voteInfos []voteInfo
		if len(ints) > 0 {
			parentPos := parentHeader.HashPos()
			wallets := self.sero.AccountManager().Wallets()
			for i, share := range shares {
				var pool *stake.StakePool
				if share.PoolId != nil {
					pool = stakeState.GetStakePool(*share.PoolId)
					if pool == nil {
						log.Error("lotteryTaskLoop", "GetStakePool", share.PoolId, "note exist")
					}
				}
				if pool != nil {
					stakeHash := types.StakeHash(&poshash, &parentPos, true)
					seed := GetSeedByVotePkr(wallets, pool.VotePKr)
					if seed != nil {
						voteInfos = append(voteInfos, voteInfo{
							ints[i],
							parentNumber.Uint64(),
							common.BytesToHash(share.Id()),
							poshash,
							stakeHash,
							pool.VotePKr,
							true,
							*seed})
					}
				}
				shareVoteSeed := GetSeedByVotePkr(wallets, share.VotePKr)
				if shareVoteSeed != nil {
					stakeHash := types.StakeHash(&poshash, &parentPos, false)
					info := voteInfo{
						ints[i],
						parentNumber.Uint64(),
						common.BytesToHash(share.Id()),
						poshash,
						stakeHash,
						share.VotePKr,
						false,
						*shareVoteSeed}
					if cotainsVoteInfo(voteInfos, info, pool) {
						continue
					} else {
						voteInfos = append(voteInfos, info)
					}
				}

			}
		}
		return voteInfos, nil
	}

}

func (self *Voter) sign(info voteInfo) {
	data := c_type.Uint256{}
	copy(data[:], info.statkeHash[:])
	sk := superzk.Seed2Sk(info.seed.SeedToUint256())
	sign, err := superzk.SignPKr(&sk, &data, &info.votePKr)
	if err != nil {
		log.Error("voter sign", "sign err", err)
		return
	}
	log.Info(">>>>>>>>>>>>>sign vote", "poshas", info.poshash, "block", info.parentNum+1, "share", info.shareHash, "idx", info.index, "isPool", info.isPool)
	vote := &types.Vote{info.index, info.parentNum, info.shareHash, info.poshash, info.isPool, sign}
	//go self.voteWorkFeed.Send(core.NewVoteEvent{vote})
	self.AddVote(vote)
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
func (self *Voter) SubscribeNewVoteEvent(ch chan<- core.NewVoteEvent) event.Subscription {
	return self.scope.Track(self.voteFeed.Subscribe(ch))
}

func (self *Voter) SubscribeWorkerVoteEvent(ch chan<- core.NewVoteEvent) event.Subscription {
	return self.scope.Track(self.voteWorkFeed.Subscribe(ch))
}

func (self *Voter) SubscribeNewLotteryEvent(ch chan<- core.NewLotteryEvent) event.Subscription {
	return self.scope.Track(self.lotteryFeed.Subscribe(ch))
}

func (self *Voter) SendLotteryEvent(lottery *types.Lottery) {
	go self.lotteryFeed.Send(core.NewLotteryEvent{lottery})
}

func (self *Voter) SendVoteEvent(vote *types.Vote) {
	go self.voteFeed.Send(core.NewVoteEvent{vote})
}

func (self *Voter) AddLottery(lottery *types.Lottery) {
	self.lotteryMu.Lock()
	defer self.lotteryMu.Unlock()
	current := self.chain.CurrentBlock().Number().Uint64()
	if current > lottery.ParentNum+delayNum {
		log.Trace("AddLottery droped", "current", current, "voteBlock", lottery.ParentNum+1)
		return
	}
	_, exits := self.lotterys[lottery.PosHash]
	if !exits {
		log.Trace("AddLottery", "poshas", lottery.PosHash, "block", lottery.ParentNum+1)
		self.lotterys[lottery.PosHash] = time.Now()
		self.lotteryCh <- lottery
		self.SendLotteryEvent(lottery)
	}
}

func (self *Voter) getStateByNumber(num uint64) (*state.StateDB, error) {
	header := self.chain.GetHeaderByNumber(num)
	if header == nil {
		return nil, nil
	}
	return self.chain.StateAt(header)

}

func (self *Voter) AddVote(vote *types.Vote) {
	self.voteMu.Lock()
	defer self.voteMu.Unlock()

	current := self.chain.CurrentBlock().Number().Uint64()
	if current > vote.ParentNum+delayNum {
		log.Trace("AddVote droped", "current", current, "voteBlock", vote.ParentNum+1)
		return
	}
	_, exits := self.votes[vote.Hash()]
	if !exits {
		log.Trace("AddVote", "hashpos", vote.PosHash, "block", vote.ParentNum+1)
		go self.voteWorkFeed.Send(core.NewVoteEvent{vote})
		self.SendVoteEvent(vote)
		self.votes[vote.Hash()] = time.Now()
	}
}
