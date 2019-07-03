package consensus

import (
	"errors"
)

type Cpoint struct {
	objPre   string
	statePre string
	inCons   bool
	cons     *Cons
}

func (self *Cpoint) objName(name []byte) (ret string) {
	return self.objPre + string(name)
}

func (self *Cpoint) stateName(name []byte) (ret string) {
	return self.statePre + string(name)
}

func (self *Cpoint) AddObj(item PItem) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	name := self.objName(item.Id())
	if self.statePre != "" {
		if state := item.State(); state != nil {
			v := BytePair{item.Id(), state}
			if self.inCons {
				self.cons.addObj(name, &v, ITEMTYPE_CONS)
			} else {
				self.cons.addObj(name, &v, ITEMTYPE_CACHE)
			}
			stateName := self.stateName(state)
			if self.inCons {
				self.cons.addObj(stateName, item, ITEMTYPE_DB)
			} else {
				self.cons.addObj(stateName, item, ITEMTYPE_CACHE)
			}
		} else {
			v := BytePair{item.Id(), []byte{1}}
			if self.inCons {
				self.cons.addObj(name, &v, ITEMTYPE_CONS)
			} else {
				self.cons.addObj(name, &v, ITEMTYPE_CACHE)
			}
			stateName := self.stateName(item.Id())
			if self.inCons {
				self.cons.addObj(stateName, item, ITEMTYPE_DB)
			} else {
				self.cons.addObj(stateName, item, ITEMTYPE_CACHE)
			}
		}
	} else {
		if state := item.State(); state != nil {
			v := Bytes(state)
			if self.inCons {
				self.cons.addObj(name, &v, ITEMTYPE_CONS)
			} else {
				self.cons.addObj(name, &v, ITEMTYPE_CACHE)
			}
		} else {
			if self.inCons {
				self.cons.addObj(name, item, ITEMTYPE_CONS)
			} else {
				self.cons.addObj(name, item, ITEMTYPE_CACHE)
			}
		}
	}
	return
}

func (self *Cpoint) GetObj(key []byte, item PItem) (ret CItem) {
	name := self.objName(key)
	if self.statePre != "" {
		if state := item.State(); state != nil {
			it := ITEMTYPE_CACHE
			if self.inCons {
				it = ITEMTYPE_CONS
			}
			if v := self.cons.getObj(name, &BytePair{}, it); v != nil {
				stateName := self.stateName(v.(*BytePair).Value)
				if self.inCons {
					return self.cons.getObj(stateName, item, ITEMTYPE_DB)
				} else {
					return self.cons.getObj(stateName, item, ITEMTYPE_CACHE)
				}
			} else {
				return nil
			}
		} else {
			stateName := self.stateName(key)
			if self.inCons {
				return self.cons.getObj(stateName, item, ITEMTYPE_DB)
			} else {
				return self.cons.getObj(stateName, item, ITEMTYPE_CACHE)
			}
		}
	} else {
		if state := item.State(); state != nil {
			var v CItem
			if self.inCons {
				v = self.cons.getObj(name, &Bytes{}, ITEMTYPE_CONS)
			} else {
				v = self.cons.getObj(name, &Bytes{}, ITEMTYPE_CACHE)
			}
			if v != nil {
				return &BytePair{key, []byte(*v.(*Bytes))}
			} else {
				return nil
			}
		} else {
			if self.inCons {
				return self.cons.getObj(name, item, ITEMTYPE_CONS)
			} else {
				return self.cons.getObj(name, item, ITEMTYPE_CACHE)
			}
		}
	}
}

func (self *Cpoint) SetValue(key []byte, value []byte) {
	v := BytePair{key, value}
	self.AddObj(&v)
}

func (self *Cpoint) GetValue(key []byte) (ret []byte) {
	if v := self.GetObj(key, &BytePair{}); v != nil {
		ret = v.(*BytePair).Value
		return
	}
	return
}
