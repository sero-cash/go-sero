package state

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/pkgstate"
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

func (self *StateDB) GetZState() *zstate.ZState {
	if self.zstate == nil {
		st := StateTri{
			self,
			self.trie,
			self.db.TrieDB().DiskDB(),
			self.db.TrieDB().WDiskDB(),
		}
		self.zstate = zstate.NewState(&st, self.number)
	}
	return self.zstate
}

func (self *StateDB) GetPkgState() *pkgstate.PkgState {
	return &self.GetZState().Pkgs
}

type ZeroDB struct {
	db *StateDB
}

func (self *ZeroDB) Num() uint64 {
	return self.db.number
}

func (self *ZeroDB) CurrentTri() Tri {
	return self.db.trie
}

func (self *ZeroDB) GlobalGetter() serodb.Getter {
	return self.db.db.TrieDB().DiskDB()
}

func (self *StateDB) GetZeroDB() *ZeroDB {
	return &ZeroDB{self}
}
