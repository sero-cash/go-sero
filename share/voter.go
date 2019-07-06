package share

import (
	"sync"
	"time"

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
	chainLotterySize = 10
	lifeTime         = 30 * time.Minute

	delayNum         = 2
	lotteryQueueSize = 12
)

type blockChain interface {
	CurrentBlock() *types.Block
	GetHeaderByHash(hash common.Hash) *types.Header
	StateAt(root common.Hash, number uint64) (*state.StateDB, error)
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
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
		lotteryQueue: &PriorityQueue{},
	}
	voter.lotteryQueue.Init(lotteryQueueSize)

	// Subscribe events from blockchain
	//voter.chainHeadSub = voter.chain.SubscribeChainHeadEvent(voter.chainHeadCh)

	// Start the event loop and return
	go voter.loop()
	go voter.lotteryTaskLoop()

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
			self.lotteryQueue.PushItem(lottery.PosHash, &lotteryItem{lottery: lottery, attempts: uint8(0)}, time.Now())

		}
	}
}

const lotteryLifeTime = 28 * time.Second

func (self *Voter) voteLoop() {
	for {
		for _, item := range self.lotteryQueue.GetQueueItems() {
			if time.Since(item.Time) > lotteryLifeTime {
				continue
			}
			lItem := item.Value.(lotteryItem)
			parentHeader := self.chain.GetHeaderByHash(lItem.lottery.ParentHash)
			if parentHeader == nil {
				continue
			}
			currentNum := self.chain.CurrentBlock().NumberU64()
			if currentNum > parentHeader.Number.Uint64()+uint64(delayNum) {
				continue
			}
			selfShares, err := self.SelfShares(lItem.lottery.PosHash, parentHeader.Hash(), parentHeader.Number.Uint64())
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

type voteInfo struct {
	poshash common.Hash
	parent  common.Hash
	votePKr keys.PKr
	isPool  bool
	seed    address.Seed
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

func (self *Voter) SelfShares(poshash common.Hash, parent common.Hash, parentNumber uint64) ([]voteInfo, error) {
	state, err := self.chain.StateAt(parent, parentNumber)
	if err != nil {
		log.Info("lotteryTaskLoop", "stateAt", poshash, "err", err)
		return nil, err
	} else {
		stakeState := stake.NewStakeState(state)
		shares, err := stakeState.SeleteShare(poshash)
		if err != nil {
			return nil, err
		}
		log.Info("lotteryTaskLoop", "SeleteShare", poshash, "err", err)
		var voteInfos []voteInfo
		for _, share := range shares {
			wallets := self.sero.AccountManager().Wallets()
			if share.PoolId != nil {
				pool := stakeState.GetStakePool(*share.PoolId)
				if pool == nil {
					log.Info("lotteryTaskLoop", "GetStakePool", share.PoolId, "note exist")
				} else {
					for _, w := range wallets {
						if w.IsMine(pkrToAddress(*share.VoteKr)) {
							seed, err := w.GetSeed()
							if err != nil {
								return nil, err
							}
							voteInfos = append(voteInfos, voteInfo{poshash, parent, *share.VoteKr, true, *seed})
						}
					}
				}
			}
			for _, w := range wallets {
				if w.IsMine(pkrToAddress(*share.VoteKr)) {
					seed, err := w.GetSeed()

					if err != nil {
						return nil, err
					}
					if cotainsSeed(voteInfos, *seed) {
						continue
					} else {
						voteInfos = append(voteInfos, voteInfo{poshash, parent, *share.VoteKr, false, *seed})
					}
				}
			}

		}
		return voteInfos, nil
	}

}

func (self *Voter) sign(info voteInfo) {
	vote := &types.Vote{common.Hash{}, info.poshash, info.isPool, keys.Uint512{}}
	go self.voteWorkFeed.Send(core.NewVoteEvent{vote})
	self.SendVoteEvent(vote)
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

	} else {
		self.lotterys[lottery.PosHash] = time.Now()
		self.lotteryCh <- lottery
		self.SendLotteryEvent(lottery)
	}

}

func (self *Voter) AddVote(vote *types.Vote) {
	log.Info("AddVote", "hashpos", vote.PosHash)
	self.voteMu.Lock()
	defer self.voteMu.Unlock()
	_, exits := self.votes[vote.Hash()]
	if exits {

	} else {
		go self.voteWorkFeed.Send(core.NewVoteEvent{vote})
		self.votes[vote.Hash()] = time.Now()

	}
}
