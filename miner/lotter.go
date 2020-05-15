package miner

import (
	"time"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/stake"
)

type Lotter struct {
	worker *worker
	state  *state.StateDB
	block  *types.Block
	stake  *stake.StakeState

	lottery            types.Lottery
	currentHeaderVotes []types.HeaderVote
	parentHeaderVotes  []types.HeaderVote
}

func newLotter(worker *worker, block *types.Block, db *state.StateDB) (ret Lotter) {
	ret.worker = worker
	ret.state = db.Copy()
	ret.block = block
	ret.stake = stake.NewStakeState(ret.state)
	return
}

func selKey(idx uint32, sid []byte) string {
	return common.BytesToString(utils.EncodeNumber32(idx)) + string(sid)
}

type shareFilter struct {
	vote  *types.Vote
	share *stake.Share
}

type votesFilter struct {
	stake       *stake.StakeState
	filters     map[string]*shareFilter
	block       *types.Block
	parentBlock *types.Block
	idxs        []uint32
	shares      []*stake.Share
}

func NewVotesFilter(state *stake.StakeState, idxs []uint32, shares []*stake.Share, block *types.Block, parentBlock *types.Block) (ret votesFilter) {
	ret.stake = state
	ret.filters = make(map[string]*shareFilter)
	for i, idx := range idxs {
		share := shares[i]
		ret.filters[selKey(idx, share.Id())] = &shareFilter{nil, share}
	}
	ret.block = block
	ret.parentBlock = parentBlock
	ret.idxs = idxs
	ret.shares = shares
	return
}

func (self *votesFilter) result() (ret []types.Vote) {
	for _, v := range self.filters {
		if v.vote != nil {
			ret = append(ret, *v.vote)
		}
	}
	return
}

func (self *votesFilter) RunFilter(votes voteSet) (dels []types.Vote) {
	voteNumMap := map[c_type.Uint512]bool{}
	for _, sets := range votes {
		for _, vote := range sets {
			if _, ok := voteNumMap[vote.Sign]; ok {
				continue
			} else {
				if vote.PosHash == self.block.HashPos() {
					k := selKey(vote.Idx, vote.ShareId[:])
					if s, ok := self.filters[k]; ok {
						if s.vote == nil {
							if self.verify(&vote, s.share) {
								voteNumMap[vote.Sign] = true
								copy_vote := vote
								self.filters[k].vote = &copy_vote
							}
						}
					}
					dels = append(dels, vote)
				}
			}

		}
	}
	return
}

func (self *votesFilter) verify(vote *types.Vote, share *stake.Share) bool {
	var votePkr *c_type.PKr
	if vote.IsPool {
		if share.PoolId != nil {
			pool := self.stake.GetStakePool(*share.PoolId)
			if pool != nil && pool.CanBeVote() {
				votePkr = &pool.VotePKr
			}
		}
	} else {
		votePkr = &share.VotePKr
	}
	if votePkr != nil {
		parentPosHash := self.parentBlock.HashPos()
		stakHash := types.StakeHash(&vote.PosHash, &parentPosHash, vote.IsPool)
		if superzk.VerifyPKr_ByHeight(self.block.NumberU64(), stakHash.HashToUint256(), &vote.Sign, votePkr) {
			return true
		}
	}
	return false
}

func (self *Lotter) wait() bool {
	needWait := self.stake.NeedTwoVote(self.block.NumberU64())
	if !needWait {
		log.Info("not need pos")
	}

	parentBlock := self.worker.chain.GetBlockByHash(self.block.ParentHash())

	startTime := time.Now()

	idx, shares, err := self.stake.SeleteShare(self.block.HashPos(), self.block.NumberU64())
	if err != nil {
		log.Error("Lotter wait ", "error", err)
		return false
	}
	filter := NewVotesFilter(self.stake, idx, shares, self.block, parentBlock)
	count := 0
	for {
		currentHeader := self.worker.chain.CurrentHeader()
		if currentHeader.Hash() != self.block.ParentHash() {
			return false
		}

		if time.Since(startTime) > 5*time.Minute {
			return false
		}

		if count == 0 {
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(100 * time.Millisecond)
		}
		count++

		key := voteKey{self.block.NumberU64(), self.block.HashPos()}
		votes := self.worker.pendingVote.getMyPending(key)

		dels := filter.RunFilter(votes)
		self.worker.pendingVote.deleteVotes(key, dels)

		sels := filter.result()
		if len(sels) > 2 || !needWait {
			break
		}
	}

	parentVoteKey := voteKey{parentBlock.NumberU64(), parentBlock.HashPos()}
	parentVoteSet := self.worker.pendingVote.getMyPending(parentVoteKey)

	pidx, pshares := stake.SeleteBlockShare(self.worker.chain.GetDB(), parentBlock.Hash())
	ppBlock := self.worker.chain.GetBlockByHash(parentBlock.ParentHash())
	parentfilter := NewVotesFilter(self.stake, pidx, pshares, parentBlock, ppBlock)
	parentfilter.RunFilter(parentVoteSet)

	voteNumMap := map[c_type.Uint512]bool{}
	for _, vote := range filter.result() {
		//log.Info("pos currentVotes", "posHash", vote.PosHash, "block", vote.ParentNum+1, "share", vote.ShareId, "idx", vote.Idx)
		if _, ok := voteNumMap[vote.Sign]; ok {
			continue
		} else {
			voteNumMap[vote.Sign] = true
			self.currentHeaderVotes = append(self.currentHeaderVotes, types.HeaderVote{vote.ShareId, vote.IsPool, vote.Sign})
			if len(self.currentHeaderVotes) == 3 {
				break
			}
		}
	}

	parentVotes := parentfilter.result()
	if len(parentVotes) > 0 {
		//log.Info("parentVotes", "block", self.block.NumberU64(), "voteIds", pidx)
		voteNumMap := map[common.Hash]int{}
		for _, share := range pshares {
			voteNumMap[common.BytesToHash(share.Id())] += 1
		}

		voteMap := map[c_type.Uint512]bool{}
		for _, vote := range parentBlock.Header().CurrentVotes {
			voteMap[vote.Sign] = true
			voteNumMap[vote.Id] -= 1
		}
		for _, vote := range parentVotes {
			if _, ok := voteMap[vote.Sign]; ok {
				continue
			}
			if voteNumMap[vote.ShareId] == 0 {
				continue
			} else {
				//log.Info("pos parentVotes", "posHash", vote.PosHash, "block", vote.ParentNum+1, "share", vote.ShareId, "idx", vote.Idx)
				self.parentHeaderVotes = append(self.parentHeaderVotes, types.HeaderVote{vote.ShareId, vote.IsPool, vote.Sign})
				voteMap[vote.Sign] = true
				voteNumMap[vote.ShareId] -= 1
			}
		}
	}
	return true
}
