package state

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/consensus"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
)

type StateDbPut interface {
	Put(key, value []byte) error
}
type StateTri struct {
	db    *StateDB
	Tri   Trie
	Dbget serodb.Getter
	Dbput serodb.Putter
}

func (self *StateTri) TryGet(key []byte) ([]byte, error) {
	return self.Tri.TryGet(key)
}

func (self *StateTri) TryUpdate(key, value []byte) error {
	return self.Tri.TryUpdate(key, value)
}

func (self *StateTri) SetState(key *keys.Uint256, value *keys.Uint256) {
	self.db.SetState(EmptyAddress, common.Hash(*key), common.Hash(*value))
}
func (self *StateTri) GetState(key *keys.Uint256) (ret keys.Uint256) {
	v := self.db.GetState(EmptyAddress, common.Hash(*key))
	ret = keys.Uint256(v)
	return
}

func (self *StateTri) TryGlobalGet(key []byte) ([]byte, error) {
	return self.Dbget.Get(key)
}

func (self *StateTri) TryGlobalPut(key, value []byte) error {
	return self.Dbput.Put(key, value)
}

func (self *StateTri) GlobalGetter() serodb.Getter {
	return self.Dbget
}

func (self *StateDB) CurrentZState() *zstate.ZState {
	if self.number < 0 {
		panic(errors.New("Current ZState height num must >= 0"))
	}

	if self.zstate == nil {
		st := StateTri{
			self,
			self.trie,
			self.db.TrieDB().DiskDB(),
			self.db.TrieDB().WDiskDB(),
		}
		self.zstate = zstate.CurrentState(&st, uint64(self.number))
	}
	return self.zstate
}

func (self *StateDB) NextZState() *zstate.ZState {
	if self.number < -1 {
		panic(errors.New("Next ZState height num must >= -1"))
	}
	if self.zstate == nil {
		st := StateTri{
			self,
			self.trie,
			self.db.TrieDB().DiskDB(),
			self.db.TrieDB().WDiskDB(),
		}
		self.zstate = zstate.NextState(&st, self.number)
	}
	return self.zstate
}

type zeroDB struct {
	db *StateDB
}

func (self *zeroDB) CurrentTri() serodb.Tri {
	return self.db.trie
}

func (self *zeroDB) GlobalGetter() serodb.Getter {
	return self.db.db.TrieDB().DiskDB()
}


var StakeDB = consensus.DBObj{"STAKE$BLOCK$INDEX"}

func (self *StateDB) GetStakeCons() *consensus.Cons {
	if self.stakeState == nil {
		cons := consensus.NewCons(&zeroDB{self}, StakeDB.Pre)
		self.stakeState = &cons
	}
	return self.stakeState
}
