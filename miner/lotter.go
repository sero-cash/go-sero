package miner

import (
	"time"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/stake"
)

type Lotter struct {
	worker  *worker
	state   *state.StateDB
	header  *types.Header
	hashPos common.Hash
	stake   *stake.StakeState
	key     voteKey
	lottery types.Lottery

	parentBlock *types.Block

	currentHeaderVotes []types.HeaderVote
	parentHeaderVotes  []types.HeaderVote
}

func newLotter(worker *worker, header *types.Header, db *state.StateDB) (ret Lotter) {
	ret.worker = worker
	ret.header = header
	ret.state = db.Copy()
	ret.stake = stake.NewStakeState(ret.state)
	ret.hashPos = header.HashPos()
	ret.key = voteKey{header.Number.Uint64(), ret.hashPos}
	ret.lottery = types.Lottery{header.ParentHash, header.Number.Uint64() - 1, ret.hashPos}
	return
}

func selKey(idx uint32, sid []byte) string {
	return common.BytesToString(utils.EncodeNumber32(idx)) + string(sid)
}

func (self *Lotter) verify(vote *types.Vote, share *stake.Share) bool {

	var votePkr *keys.PKr
	if vote.IsPool {
		if share.PoolId != nil {
			pool := self.stake.GetStakePool(*share.PoolId)
			if pool != nil && pool.Valid() {
				votePkr = &pool.VotePKr
			}
		}
	} else {
		votePkr = &share.VotePKr
	}

	if votePkr != nil {
		parentPosHash := self.parentBlock.HashPos()
		stakHash := types.StakeHash(&vote.PosHash, &parentPosHash)
		if keys.VerifyPKr(stakHash.HashToUint256(), &vote.Sign, votePkr) {
			return true
		}
	}
	return false
}

type shareFilter struct {
	vote  *types.Vote
	share *stake.Share
}

type votesFilter map[string]*shareFilter

func (self *votesFilter) result() (ret []types.Vote) {
	for _, v := range *self {
		if v.vote != nil {
			ret = append(ret, *v.vote)
		}
	}
	return
}

func (self *Lotter) NewFilter(idxs []uint32, shares []*stake.Share) (filter votesFilter) {
	filter = make(votesFilter)
	for i, idx := range idxs {
		share := shares[i]
		filter[selKey(idx, share.Id())] = &shareFilter{nil, share}
	}
	return
}

func (self *Lotter) RunFilter(filter map[string]*shareFilter, votes voteSet) (dels []types.Vote) {

	for _, v := range votes {
		for _, vote := range v {
			if vote.PosHash == self.hashPos {
				k := selKey(vote.Idx, vote.ShareId[:])
				if s, ok := filter[k]; ok {
					if s.vote == nil {
						if self.verify(&vote, s.share) {
							copy_vote:=vote
							filter[k].vote = &copy_vote
						}
					}
				}
				dels = append(dels, vote)
			}
		}
	}

	return
}
func (self *Lotter) wait() bool {
	needWait := self.stake.NeedTwoVote(self.header.Number.Uint64())
	if !needWait{
		log.Info("not need pos")
	}

	self.parentBlock = self.worker.chain.GetBlockByHash(self.header.ParentHash)
	startTime := time.Now()

	idx, shares, err := self.stake.SeleteShare(self.hashPos)
	if err != nil {
		log.Error("Lotter wait ", "error", err)
		return false
	}
	filter := self.NewFilter(idx, shares)
	count:=0
	for {
		currentHeader := self.worker.chain.CurrentHeader()
		if currentHeader.Hash() != self.header.ParentHash {
			return false
		}

		if time.Since(startTime) > 5*time.Minute {
			return false
		}

		if count ==0{
			time.Sleep(1 * time.Second)
		}else{
			time.Sleep(100 * time.Millisecond)
		}
        count++

		votes := self.worker.pendingVote.getMyPending(self.key)

		dels := self.RunFilter(filter, votes)
		self.worker.pendingVote.deleteVotes(self.key, dels)

		sels := filter.result()
		if len(sels) > 2 || !needWait {
			break
		}
	}

	parentVoteKey := voteKey{self.parentBlock.NumberU64(), self.parentBlock.HashPos()}
	parentVoteSet := self.worker.pendingVote.getMyPending(parentVoteKey)

	pidx, pshares := stake.SeleteBlockShare(self.worker.chain.GetDB(), self.parentBlock.Hash())
	parentfilter := self.NewFilter(pidx, pshares)
	self.RunFilter(parentfilter, parentVoteSet)

	for _, vote := range filter.result() {
		log.Info("pos currentVotes", "posHash", vote.PosHash, "block", vote.ParentNum+1, "share", vote.ShareId, "idx", vote.Idx)
		self.currentHeaderVotes = append(self.currentHeaderVotes, types.HeaderVote{vote.ShareId, vote.IsPool, vote.Sign})
		if len(self.currentHeaderVotes) == 3 {
			break
		}
	}

	parentVotes := parentfilter.result()
	if len(parentVotes) > 0 {
		log.Info("parentVotes", "block", self.header.Number.Uint64(), "voteIds", pidx)
		voteNumMap := map[common.Hash]int{}
		for _, share := range pshares {
			voteNumMap[common.BytesToHash(share.Id())] += 1
		}

		voteMap := map[keys.Uint512]bool{}
		for _, vote := range self.parentBlock.Header().CurrentVotes {
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
				log.Info("pos parentVotes", "posHash", vote.PosHash, "block", vote.ParentNum+1, "share", vote.ShareId, "idx", vote.Idx)
				self.parentHeaderVotes = append(self.parentHeaderVotes, types.HeaderVote{vote.ShareId, vote.IsPool, vote.Sign})
				voteMap[vote.Sign] = true
				voteNumMap[vote.ShareId] -= 1
			}
		}
	}
	return true
}
