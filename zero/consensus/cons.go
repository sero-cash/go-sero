package consensus

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
)

type changelog struct {
	key string
	old CItem
	ver int
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
	dirty   bool
	inCons  bool
	inBlock bool
	inDB    bool
}

type Cons struct {
	db      DB
	content map[string]ConsItem
	cls     []changelog
	ver     int
}

func NewCons(db DB) (ret Cons) {
	ret.content = make(map[string]ConsItem)
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

func (self *Cons) DeleteObj(k *key, item CItem, inCons bool, inBlock bool, inDB bool) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	old := self.getObj(k, item.CopyTo(), inCons, inBlock, inDB)
	cl := changelog{
		k.k(),
		old,
		self.ver,
	}
	self.content[k.k()] = ConsItem{*k, nil, len(self.cls), true, inCons, inBlock, inDB}
	self.cls = append(self.cls, cl)
	return
}

func (self *Cons) addObj(k *key, item CItem, inCons bool, inBlock bool, inDB bool) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	old := self.getObj(k, item.CopyTo(), inCons, inBlock, inDB)
	cl := changelog{
		k.k(),
		old,
		self.ver,
	}
	self.content[k.k()] = ConsItem{*k, item, len(self.cls), true, inCons, inBlock, inDB}
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

func (self *Cons) getObj(k *key, item CItem, inCons bool, inBlock bool, inDB bool) (ret CItem) {
	if i, ok := self.content[k.k()]; !ok {
		if v := self.getData([]byte(k.k()), inCons, inDB); v == nil {
			return nil
		} else {
			if e := rlp.DecodeBytes(v, item); e != nil {
				return nil
			}
			self.content[k.k()] = ConsItem{*k, item, -1, false, inCons, inBlock, inDB}
			return item
		}
	} else {
		if i.item != nil {
			item.CopyFrom(i.item)
			return item
		} else {
			return nil
		}
	}
}

func (self *Cons) GetConsPairs() {
}

func (self *Cons) SaveCons() {
	//get cons
}

func (self *Cons) SaveDB(batch serodb.Batch) {
	//get blocks
	//get dbs
}

func GetObjectByState(getter serodb.Getter, state *keys.Uint256, item CItem) (ret CItem) {
	if v, err := getter.Get(state[:]); err != nil {
		return
	} else {
		if e := rlp.DecodeBytes(v, item); e != nil {
			return nil
		}
		return item
	}
}
