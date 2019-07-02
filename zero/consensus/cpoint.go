package consensus

import (
	"errors"

	"github.com/sero-cash/go-czero-import/keys"
)

type cpoint struct {
	objPre   string
	statePre string
	inCons   bool
	cons     *cons
}

func (self *cpoint) ObjName(name *keys.Uint256) (ret string) {
	return self.objPre + string(name[:])
}

func (self *cpoint) StateName(name *keys.Uint256) (ret string) {
	return self.statePre + string(name[:])
}

func (self *cpoint) AddObj(key *keys.Uint256, item CItem) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	name := self.ObjName(key)
	if self.statePre != "" {
		if state := item.State(); state != nil {
			v := Bytes(state[:])
			if self.inCons {
				self.cons.addObj(name, &v, ITEMTYPE_CONS)
			} else {
				self.cons.addObj(name, &v, ITEMTYPE_CACHE)
			}
			stateName := self.StateName(state)
			if self.inCons {
				self.cons.addObj(stateName, item, ITEMTYPE_DB)
			} else {
				self.cons.addObj(stateName, item, ITEMTYPE_CACHE)
			}
		} else {
			v := Bytes([]byte{1})
			if self.inCons {
				self.cons.addObj(name, &v, ITEMTYPE_CONS)
			} else {
				self.cons.addObj(name, &v, ITEMTYPE_CACHE)
			}
			stateName := self.StateName(key)
			if self.inCons {
				self.cons.addObj(stateName, item, ITEMTYPE_DB)
			} else {
				self.cons.addObj(stateName, item, ITEMTYPE_CACHE)
			}
		}
	} else {
		if self.inCons {
			self.cons.addObj(name, item, ITEMTYPE_CONS)
		} else {
			self.cons.addObj(name, item, ITEMTYPE_CACHE)
		}
	}
	return
}

func (self *cpoint) GetObj(key *keys.Uint256, item CItem) (ret CItem) {
	name := self.ObjName(key)
	if self.statePre != "" {
		if state := item.State(); state != nil {
			it := ITEMTYPE_CACHE
			if self.inCons {
				it = ITEMTYPE_CONS
			}
			if v := self.cons.getObj(name, &Bytes{}, it); v != nil {
				copy(state[:], (*v.(*Bytes)))
				stateName := self.StateName(state)
				if self.inCons {
					return self.cons.getObj(stateName, item, ITEMTYPE_DB)
				} else {
					return self.cons.getObj(stateName, item, ITEMTYPE_CACHE)
				}
			} else {
				return nil
			}
		} else {
			stateName := self.StateName(state)
			if self.inCons {
				return self.cons.getObj(stateName, item, ITEMTYPE_DB)
			} else {
				return self.cons.getObj(stateName, item, ITEMTYPE_CACHE)
			}
		}
	} else {
		if self.inCons {
			return self.cons.getObj(name, item, ITEMTYPE_CONS)
		} else {
			return self.cons.getObj(name, item, ITEMTYPE_CACHE)
		}
	}
}

func (self *cpoint) SetValue(key *keys.Uint256, value *keys.Uint256) {
	v := Bytes(value[:])
	self.AddObj(key, &v)
}

func (self *cpoint) GetValue(key *keys.Uint256) (ret *keys.Uint256) {
	if v := self.GetObj(key, &Bytes{}); v != nil {
		ret = &keys.Uint256{}
		copy(ret[:], (*v.(*Bytes))[:])
	}
	return
}
