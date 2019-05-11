package state1

import (
	"os"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/sero-cash/go-sero/zero/lstate"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/zero/txs/zstate"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

type SeroDB struct {
	db *leveldb.DB
}

func (self *SeroDB) Get(key []byte) ([]byte, error) {
	return self.db.Get(key, nil)
}
func (self *SeroDB) Has(key []byte) (bool, error) {
	return self.db.Has(key, nil)
}

func (self *SeroDB) Put(key []byte, value []byte) error {
	return self.db.Put(key, value, nil)
}

type State1 struct {
	db           SeroDB
	bc           lstate.BlockChain
	last_st      *State1_storage
	st           *State1_storage
	zst          *zstate.ZState
	tks          []keys.Uint512
	saveing_name string
	saveing_num  uint64
}

func NewState1(bc lstate.BlockChain) (ret State1) {
	ret.bc = bc
	if db, err := leveldb.OpenFile(zconfig.State1_db_dir(), nil); err != nil {
		panic(err)
	} else {
		ret.db = SeroDB{db}
	}
	return
}

func (self *State1) valid() bool {
	if self.last_st == nil {
		return false
	} else {
		return true
	}
}

func (self *State1) begin(last_file_name string, chose_hash *common.Hash, tks []keys.Uint512) {
	if chose_hash != nil {
		self.zst = self.bc.NewState(chose_hash)
	}
	self.tks = tks
	if self.last_st == nil {
		if len(last_file_name) != 0 {
			lst := loadState(self.zst, last_file_name)
			self.last_st = &lst
		}
	}
}

func (self *State1) needParse(num uint64, hash *common.Hash) (ret bool, e error) {
	current_name := state1_file_name(num, hash)
	current_file := zconfig.State1_file(current_name)
	if _, err := os.Stat(current_file); err != nil {
		if os.IsNotExist(err) {
			ret = true
			return
		} else {
			e = err
			return
		}
	} else {
		ret = false
		return
	}

}

func (self *State1) update(parent_hash *common.Hash, num uint64, hash *common.Hash, block *localdb.Block) {
	self.saveing_name = state1_file_name(num, hash)
	self.saveing_num = num
	var load_name string

	if num == 0 {
		load_name = ""
	} else {
		parent_num := num - 1
		load_name = state1_file_name(parent_num, parent_hash)
	}

	if self.st == nil {
		s1 := loadState(self.zst, load_name)
		self.st = &s1
	} else {
		self.st.State = self.zst
	}

	self.st.updateWitness(self.tks, num, block)
}

func (self *State1) save() {
	self.st.finalize(self.saveing_name, self.saveing_num)
	self.last_st = self.st
	self.st = nil
}

func (self *State1) ZState() *zstate.ZState {
	return self.zst
}

func (self *State1) GetOut(root *keys.Uint256) (src *lstate.OutState, e error) {
	if self.last_st != nil {
		return self.last_st.GetOut(root)
	} else {
		e = errors.New("GetOut but state1 is nil")
		return
	}
}

func (self *State1) GetPkgs(tk *keys.Uint512, is_from bool) (ret []*lstate.Pkg) {
	if self.last_st != nil {
		return self.last_st.GetPkgs(tk, is_from)
	} else {
		return nil
	}
}
func (self *State1) GetOuts(tk *keys.Uint512) (outs []*lstate.OutState, e error) {
	if self.last_st != nil {
		return self.last_st.GetOuts(tk)
	} else {
		e = errors.New("GetOuts but state1 is nil")
		return
	}
}
