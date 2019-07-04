package consensus

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/common"
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

type consItem struct {
	key     key
	item    CItem
	index   int
	inCons  bool
	inBlock string
	inDB    bool
}

type Cons struct {
	db        DB
	pre       string
	content   map[string]*consItem
	cls       []changelog
	ver       int
	updateVer int
}

func NewCons(db DB, pre string) (ret Cons) {
	ret.content = make(map[string]*consItem)
	ret.pre = pre
	ret.db = db
	ret.ver = -1
	ret.updateVer = -1
	return
}

func (self *Cons) CreateSnapshot(ver int) {
	self.ver = ver
}

func (self *Cons) RevertToSnapshot(ver int) {
	if ver <= self.updateVer {
		panic(fmt.Errorf("revert snapshot version(%v) < update version(%v)", ver, self.updateVer))
	}
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

func (self *Cons) deleteObj(k *key, item CItem, inCons bool, inBlock string, inDB bool) {
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
	self.content[k.k()] = &consItem{*k, nil, len(self.cls), inCons, inBlock, inDB}
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
	self.content[k.k()] = &consItem{*k, item, len(self.cls), inCons, inBlock, inDB}
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

func (self *Cons) getObj(k *key, item CItem, inCons bool, inBlock string, inDB bool) (ret *consItem) {
	if i, ok := self.content[k.k()]; !ok {
		if v := self.getData([]byte(k.k()), inCons, inDB); v == nil {
			return nil
		} else {
			if e := rlp.DecodeBytes(v, item); e != nil {
				return nil
			}
			ret = &consItem{*k, item, -1, false, "", false}
			self.content[k.k()] = ret
			return ret
		}
	} else {
		item.CopyFrom(i.item)
		return i
	}
}

type consItems []*consItem

func (self consItems) Len() int {
	return len(self)
}
func (self consItems) Less(i, j int) bool {
	return self[i].index < self[j].index
}
func (self consItems) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *Cons) fetchConsPairs(onlyget bool) (ret consItems) {
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

func (self *Cons) fetchDBPairs(onlyget bool) (ret consItems) {
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

func (self *Cons) fetchBlockRecords(onlyget bool) (ret []*Record) {
	cis := consItems{}
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
	conslist := self.fetchConsPairs(false)
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
	recordlist := self.fetchBlockRecords(false)

	if len(recordlist) > 0 {
		DBObj{self.pre}.setBlockRecords(batch, self.db.Num(), hash, recordlist)
	}

	dblist := self.fetchDBPairs(false)
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
