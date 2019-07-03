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

type ItemType int

const (
	ITEMTYPE_CACHE = ItemType(0)
	ITEMTYPE_CONS  = ItemType(1)
	ITEMTYPE_DB    = ItemType(2)
)

type ConsItem struct {
	item     CItem
	index    int
	dirty    bool
	itemType ItemType
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

func (self *Cons) CreatePoint(objPre string, statePre string, inCons bool) (ret Cpoint) {
	ret.objPre = objPre
	ret.statePre = statePre
	ret.inCons = inCons
	ret.cons = self
	return
}

func (self *Cons) DeleteObj(name string, item CItem, it ItemType) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	old := self.getObj(name, item.CopyTo(), it)
	cl := changelog{
		name,
		old,
		self.ver,
	}
	self.content[name] = ConsItem{nil, len(self.cls), true, it}
	self.cls = append(self.cls, cl)
	return
}

func (self *Cons) addObj(name string, item CItem, it ItemType) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	old := self.getObj(name, item.CopyTo(), it)
	cl := changelog{
		name,
		old,
		self.ver,
	}
	self.content[name] = ConsItem{item, len(self.cls), true, it}
	self.cls = append(self.cls, cl)
	return
}

func (self *Cons) getData(k []byte, it ItemType) (ret []byte) {
	switch it {
	case ITEMTYPE_CACHE:
		return
	case ITEMTYPE_CONS:
		if v, err := self.db.CurrentTri().TryGet(k); err != nil {
			return
		} else {
			return v
		}
	case ITEMTYPE_DB:
		if v, err := self.db.GlobalGetter().Get(k); err != nil {
			return
		} else {
			return v
		}
	default:
		panic(errors.New("Unknow item type"))
	}
}

func (self *Cons) getObj(name string, item CItem, it ItemType) (ret CItem) {
	if i, ok := self.content[name]; !ok {
		if v := self.getData([]byte(name), it); v == nil {
			return nil
		} else {
			if e := rlp.DecodeBytes(v, item); e != nil {
				return nil
			}
			self.content[name] = ConsItem{item, -1, false, it}
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
