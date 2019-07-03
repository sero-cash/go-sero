package consensus

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
)

type changelog struct {
	key   string
	old   CItem
	index int
	ver   int
}

type key struct {
	pre string
	id  []byte
}

func (self *key) k() string {
	return self.pre + string(self.id)
}

type ConsItem struct {
	key     key
	item    CItem
	index   int
	inCons  bool
	inBlock string
	inDB    bool
}

type Cons struct {
	db      DB
	content map[string]*ConsItem
	cls     []changelog
	ver     int
}

func NewCons(db DB) (ret Cons) {
	ret.content = make(map[string]*ConsItem)
	ret.db = db
	ret.ver = -1
	return
}

func (self *Cons) CreateSnapshot(ver int) {
	self.ver = ver
}

func (self *Cons) RevertToSnapshot(ver int) {
	l := len(self.cls)
	for ; l > 0; l-- {
		cl := self.cls[l-1]
		if cl.ver >= ver {
			if cl.old != nil {
				if v, ok := self.content[cl.key]; ok {
					v.item = cl.old
					self.content[cl.key] = v
				} else {
					panic(errors.New("revert snapshot but can not find the keypair"))
				}
			} else {
				delete(self.content, cl.key)
			}
		} else {
			break
		}
	}
	if l != len(self.cls) {
		self.cls = self.cls[:l]
	}
}

func (self *Cons) DeleteObj(k *key, item CItem, inCons bool, inBlock string, inDB bool) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	old := self.getObj(k, item.CopyTo(), inCons, inBlock, inDB)
	cl := changelog{}
	cl.key = k.k()
	if old != nil {
		cl.old = old.item
		cl.index = old.index
	} else {
		cl.old = nil
		cl.index = -1
	}
	cl.ver = self.ver
	self.content[k.k()] = &ConsItem{*k, nil, len(self.cls), inCons, inBlock, inDB}
	self.cls = append(self.cls, cl)
	return
}

func (self *Cons) addObj(k *key, item CItem, inCons bool, inBlock string, inDB bool) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	old := self.getObj(k, item.CopyTo(), inCons, inBlock, inDB)
	cl := changelog{}
	cl.key = k.k()
	if old != nil {
		cl.old = old.item
		cl.index = old.index
	} else {
		cl.old = nil
		cl.index = -1
	}
	cl.ver = self.ver
	self.content[k.k()] = &ConsItem{*k, item, len(self.cls), inCons, inBlock, inDB}
	self.cls = append(self.cls, cl)
	return
}

func (self *Cons) getData(k []byte, inCons bool, inDB bool) (ret []byte) {
	if inDB {
		if v, err := self.db.GlobalGetter().Get(k); err != nil {
			return
		} else {
			return v
		}
	}
	if inCons {
		if v, err := self.db.CurrentTri().TryGet(k); err != nil {
			return
		} else {
			return v
		}
	}
	return nil
}

func (self *Cons) getObj(k *key, item CItem, inCons bool, inBlock string, inDB bool) (ret *ConsItem) {
	if i, ok := self.content[k.k()]; !ok {
		if v := self.getData([]byte(k.k()), inCons, inDB); v == nil {
			return nil
		} else {
			if e := rlp.DecodeBytes(v, item); e != nil {
				return nil
			}
			ret = &ConsItem{*k, item, -1, false, "", false}
			self.content[k.k()] = ret
			return ret
		}
	} else {
		item.CopyFrom(i.item)
		return i
	}
}

type ConsItems []*ConsItem

func (self ConsItems) Len() int {
	return len(self)
}
func (self ConsItems) Less(i, j int) bool {
	return self[i].index < self[j].index
}
func (self ConsItems) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *Cons) FetchConsPairs(onlyget bool) (ret ConsItems) {
	for _, v := range self.content {
		if v.inCons {
			ret = append(ret, v)
			if !onlyget {
				v.inCons = false
			}
		}
	}
	sort.Sort(ret)
	return
}

func (self *Cons) FetchDBPairs(onlyget bool) (ret ConsItems) {
	for _, v := range self.content {
		if v.inDB {
			ret = append(ret, v)
			if !onlyget {
				v.inDB = false
			}
		}
	}
	sort.Sort(ret)
	return
}

type Record struct {
	Name   string
	Hashes [][]byte
}

func (self *Cons) FetchBlockRecords(onlyget bool) (ret []*Record) {
	cis := ConsItems{}
	for _, v := range self.content {
		if len(v.inBlock) > 0 {
			cis = append(cis, v)
		}
	}
	sort.Sort(cis)

	m := make(map[string]*Record)

	for _, v := range cis {
		var r *Record
		if record, ok := m[v.inBlock]; ok {
			r = record
		} else {
			r = &Record{Name: v.inBlock}
			m[v.inBlock] = r
			ret = append(ret, r)
		}
		r.Hashes = append(r.Hashes, v.key.id)
	}

	for _, v := range cis {
		v.inBlock = ""
	}
	return
}

func (self *Cons) Update() {
	conslist := self.FetchConsPairs(false)
	for _, v := range conslist {
		if b, err := rlp.EncodeToBytes(v.item); err != nil {
			panic(err)
		} else {
			if err := self.db.CurrentTri().TryUpdate([]byte(v.key.k()), b); err != nil {
				panic(err)
			} else {
				return
			}
		}
	}
}

func (self *Cons) Record(hash *common.Hash, batch serodb.Putter) {
	recordlist := self.FetchBlockRecords(false)
	SetBlockRecords(batch, self.db.Num(), hash, recordlist)

	dblist := self.FetchDBPairs(false)
	for _, v := range dblist {
		if b, err := rlp.EncodeToBytes(v.item); err != nil {
			panic(err)
		} else {
			if err := batch.Put([]byte(v.key.k()), b); err != nil {
				panic(err)
			}
		}
	}
}

const (
	BLOCK_RECORDS_NAME = "ZERO$CONS$BLOCK$RECORDS$"
)

func makeBlockName(num uint64, hash *common.Hash) (ret []byte) {
	ret = []byte(BLOCK_RECORDS_NAME)
	ret = append(ret, big.NewInt(int64(num)).Bytes()...)
	ret = append(ret, hash[:]...)
	return
}

func SetBlockRecords(batch serodb.Putter, num uint64, hash *common.Hash, records []*Record) {
	if b, err := rlp.EncodeToBytes(&records); err != nil {
		panic(err)
	} else {
		name := makeBlockName(num, hash)
		if err := batch.Put(name, b); err != nil {
			panic(err)
		} else {
			return
		}
	}
}

func GetBlockRecords(getter serodb.Getter, num uint64, hash *common.Hash) (records []*Record) {
	if b, err := getter.Get(makeBlockName(num, hash)); err != nil {
		return
	} else {
		if err := rlp.DecodeBytes(b, &records); err != nil {
			panic(err)
		} else {
			return
		}
	}
}

type DBObj struct {
	Pre string
}

func (self *DBObj) GetObject(getter serodb.Getter, id []byte, item CItem) (ret CItem) {
	k := key{self.Pre, id}
	if v, err := getter.Get([]byte(k.k())); err != nil {
		return
	} else {
		if e := rlp.DecodeBytes(v, item); e != nil {
			return nil
		}
		return item
	}
}
