package share

import (
	"sync"
	"time"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/types"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/params"
)

const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
	chainLotterySize  = 10

	evictionInterval = time.Minute
	lifeTime         = 30 * time.Minute
	delayNum         = 2
)

type blockChain interface {
	CurrentBlock() *types.Block
	GetHeaderByHash(hash common.Hash) *types.Header

	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
}

type Backend interface {
	AccountManager() *accounts.Manager
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

	currentState  *state.StateDB      // Current state in the blockchain head
	pendingState  *state.ManagedState // Pending state tracking virtual nonces
	currentMaxGas uint64              // Current gas limit for transaction caps

	wg       sync.WaitGroup // for shutdown sync
	votes    map[common.Hash]time.Time
	lotterys map[common.Hash]time.Time
}

func NewVoter(chainconfig *params.ChainConfig, chain blockChain, sero Backend) *Voter {
	// Sanitize the input to ensure no vulnerable gas prices are set

	// Create the transaction pool with its initial settings
	voter := &Voter{
		chainconfig: chainconfig,
		sero:        sero,
		chain:       chain,
		lotteryCh:   make(chan *types.Lottery, chainLotterySize),
	}

	// Subscribe events from blockchain
	//voter.chainHeadSub = voter.chain.SubscribeChainHeadEvent(voter.chainHeadCh)

	// Start the event loop and return
	go voter.loop()
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

func (self *Voter) voteLoop() {
	for {
		select {
		case lottery := <-self.lotteryCh:
			if self.chain.CurrentBlock().NumberU64() > lottery.ParentNum+uint64(delayNum) {

			} else {
				parentHeader := self.chain.GetHeaderByHash(lottery.PosHash)
				if parentHeader != nil {
					go self.sign(lottery, parentHeader)
				} else {

				}
			}

		}
	}
}

func (self *Voter) sign(lottery *types.Lottery, parentHeader *types.Header) {
	vote := &types.Vote{1, lottery.PosHash, keys.Uint512{}}
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
	log.Info("AddLottery", "block", lottery.ParentNum)
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
