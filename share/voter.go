package share

import (
	"math/big"
	"sync"
	"time"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero/zero/stake"

	"github.com/sero-cash/go-czero-import/keys"

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

	delayNum         = 6
	lotteryQueueSize = 12
)

type blockChain interface {
	CurrentBlock() *types.Block
	GetHeaderByHash(hash common.Hash) *types.Header
	StateAt(root common.Hash, number uint64) (*state.StateDB, error)
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	GetHeader(hash common.Hash, number uint64) *types.Header
	GetHeaderByNumber(number uint64) *types.Header
	GetBlock(hash common.Hash, number uint64) *types.Block
	GetDB() serodb.Database
}

type Voter struct {
	chainconfig  *params.ChainConfig
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
		chainconfig:  chainconfig,
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
	//go voter.voteLoop()

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

func (self *Voter) lotteryTaskLoop() {
	for {
		select {
		case lottery := <-self.lotteryCh:
			current := self.chain.CurrentBlock().NumberU64()
			log.Info("lotteryTaskLoop new lottery", "poshash", lottery.PosHash, "block", lottery.ParentNum+1, "localBlock", current)
			if current+delayNum >= lottery.ParentNum {
				parentHeader := self.chain.GetHeaderByHash(lottery.ParentHash)
				if parentHeader == nil {
					log.Info("lotteryTaskLoop can not find parentblokc", "parent block", lottery.ParentNum)
					self.lotteryQueue.PushItem(lottery.PosHash, &lotteryItem{Lottery: lottery, Attempts: uint8(0)}, lottery.ParentNum+1)
				} else {
					selfShares, err := self.SelfShares(lottery.PosHash, lottery.ParentHash, parentHeader.Number)
					if err != nil {
						log.Info("lotteryTaskLoop", "selfShare error ", err)
					} else {
						for _, s := range selfShares {
							go self.sign(s)
						}
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
				if lItem.Lottery.ParentNum+delayNum < current {
					log.Info("not need vote", "current", current, "vote block", lItem.Lottery.ParentNum+1)
					continue
				}
				parentHeader := self.chain.GetHeaderByHash(lItem.Lottery.ParentHash)
				if parentHeader == nil {
					lItem.Attempts += 1
					if lItem.Attempts < 2 {
						self.lotteryQueue.PushItem(lItem.Lottery.PosHash, lItem, item.Block)
					}
					continue
				}
				selfShares, err := self.SelfShares(lItem.Lottery.PosHash, parentHeader.Hash(), parentHeader.Number)
				if err != nil {
					log.Info("lotteryTaskLoop", "selfShare error ", err)
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
	votePKr    keys.PKr
	isPool     bool
	seed       address.Seed
}

func cotainsSeed(voteInfos []voteInfo, seed address.Seed) bool {
	for _, v := range voteInfos {
		if v.seed == seed {
			return true
		}
	}
	return false
}

func pkrToAddress(pkr keys.PKr) common.Address {
	var addr common.Address
	copy(addr[:], pkr[:])
	return addr
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
	state, err := self.chain.StateAt(parentHeader.Root, parentNumber.Uint64())
	if err != nil {
		log.Info("lotteryTaskLoop", "stateAt", poshash, "err", err)
		return nil, err
	} else {
		stakeState := stake.NewStakeState(state)

		header := &types.Header{
			ParentHash: parent,
			Number:     new(big.Int).Add(parentNumber, common.Big1),
		}
		stakeState.ProcessBeforeApply(self.chain, header)
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
		for i, share := range shares {
			wallets := self.sero.AccountManager().Wallets()
			if share.PoolId != nil {
				pool := stakeState.GetStakePool(*share.PoolId)
				if pool == nil {
					log.Info("lotteryTaskLoop", "GetStakePool", share.PoolId, "note exist")
				} else {
					for _, w := range wallets {
						if w.IsMine(pkrToAddress(share.VotePKr)) {
							seed, err := w.GetSeed()
							if err != nil {
								return nil, err
							}
							parentPos := parentHeader.HashPos()
							stakeHash := types.StakeHash(&poshash, &parentPos)
							voteInfos = append(voteInfos, voteInfo{
								ints[i],
								parentNumber.Uint64(),
								common.BytesToHash(share.Id()),
								poshash,
								stakeHash,
								share.VotePKr,
								true,
								*seed})
						}
					}
				}
			}
			for _, w := range wallets {
				if w.IsMine(pkrToAddress(share.VotePKr)) {
					seed, err := w.GetSeed()

					if err != nil {
						return nil, err
					}
					if cotainsSeed(voteInfos, *seed) {
						//continue
						parentPos := parentHeader.HashPos()
						stakeHash := types.StakeHash(&poshash, &parentPos)
						voteInfos = append(voteInfos, voteInfo{
							ints[i],
							parentNumber.Uint64(),
							common.BytesToHash(share.Id()),
							poshash,
							stakeHash,
							share.VotePKr,
							false,
							*seed})
					} else {
						parentPos := parentHeader.HashPos()
						stakeHash := types.StakeHash(&poshash, &parentPos)
						voteInfos = append(voteInfos, voteInfo{
							ints[i],
							parentNumber.Uint64(),
							common.BytesToHash(share.Id()),
							poshash,
							stakeHash,
							share.VotePKr,
							false,
							*seed})
					}
				}
			}

		}
		return voteInfos, nil
	}

}

func (self *Voter) sign(info voteInfo) {
	data := keys.Uint256{}
	copy(data[:], info.statkeHash[:])
	sign, err := keys.SignPKr(info.seed.SeedToUint256(), &data, &info.votePKr)
	if err != nil {
		log.Info("voter sign", "sign err", err)
		return
	}
	log.Info("sign vote", "poshas", info.poshash, "block", info.parentNum+1, "share", info.shareHash, "idx", info.index)
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
	log.Info("AddLottery", "block", lottery.ParentHash)
	self.lotteryMu.Lock()
	defer self.lotteryMu.Unlock()
	_, exits := self.lotterys[lottery.PosHash]
	if exits {
		self.SendLotteryEvent(lottery)
	} else {
		self.lotterys[lottery.PosHash] = time.Now()
		self.lotteryCh <- lottery
		self.SendLotteryEvent(lottery)
	}

}

func (self *Voter) getStakeHashByParentHash(posHash common.Hash, parentHash common.Hash) common.Hash {
	parentHeader := self.chain.GetHeaderByHash(parentHash)
	if parentHeader == nil {
		return common.Hash{}
	}
	parentPosHash := parentHeader.HashPos()
	stakeHash := types.StakeHash(&posHash, &parentPosHash)
	return stakeHash

}

func (self *Voter) getStakeHashByParentNumb(posHash common.Hash, parentNumber uint64) common.Hash {
	parentHeader := self.chain.GetHeaderByNumber(parentNumber)
	if parentHeader == nil {
		return common.Hash{}
	}
	parentPosHash := parentHeader.HashPos()
	stakeHash := types.StakeHash(&posHash, &parentPosHash)
	return stakeHash

}
func (self *Voter) getStateByNumber(num uint64) (*state.StateDB, error) {
	header := self.chain.GetHeaderByNumber(num)
	if header == nil {
		return nil, nil
	}
	return self.chain.StateAt(header.Root, header.Number.Uint64())

}

func (self *Voter) AddVote(vote *types.Vote) {
	log.Info("AddVote", "hashpos", vote.PosHash)
	self.voteMu.Lock()
	defer self.voteMu.Unlock()

	current := self.chain.CurrentBlock().Number().Uint64()
	if current > vote.ParentNum+delayNum {
		return
	}
	_, exits := self.votes[vote.Hash()]
	if !exits {
		go self.voteWorkFeed.Send(core.NewVoteEvent{vote})
		self.SendVoteEvent(vote)
		self.votes[vote.Hash()] = time.Now()
	}
}
