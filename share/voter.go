package share

import (
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/types"
	"sync"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/params"
)


const (
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
)

type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash, number uint64) (*state.StateDB, error)

	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription


type Voter struct {
	chainconfig  *params.ChainConfig
	chain        blockChain
	voteFeed     event.Feed
	scope        event.SubscriptionScope
	chainHeadCh  chan core.ChainHeadEvent
	chainHeadSub event.Subscription
	//abi       types.Signer
	mu sync.RWMutex

	currentState  *state.StateDB      // Current state in the blockchain head
	pendingState  *state.ManagedState // Pending state tracking virtual nonces
	currentMaxGas uint64              // Current gas limit for transaction caps

	wg sync.WaitGroup // for shutdown sync
}

func NewVoter(chainconfig *params.ChainConfig, chain blockChain) *Voter {
	// Sanitize the input to ensure no vulnerable gas prices are set

	// Create the transaction pool with its initial settings
	voter:= &Voter{
		chainconfig: chainconfig,
		chain:       chain,
		chainHeadCh: make(chan core.ChainHeadEvent, chainHeadChanSize),
	}

	// Subscribe events from blockchain
	voter.chainHeadSub = voter.chain.SubscribeChainHeadEvent(voter.chainHeadCh)

	// Start the event loop and return

	return voter
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
func (pool *Voter) SubscribeNewVoteEvent(ch chan<- core.NewVoteEvent) event.Subscription {
	return pool.scope.Track(pool.voteFeed.Subscribe(ch))
}