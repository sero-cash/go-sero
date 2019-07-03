package consensus

import "errors"

type ObjPoint struct {
	objPre   string
	statePre string
	cons     *Cons
}

func NewObjPt(cons *Cons, objPre string, statePre string) (ret ObjPoint) {
	ret.objPre = objPre
	ret.statePre = statePre
	ret.cons = cons
	return
}

func (self *ObjPoint) AddObj(item PItem) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	stateHash := Bytes(item.State())
	self.cons.addObj(&key{self.objPre, item.Id()}, &stateHash, true, false, false)
	self.cons.addObj(&key{self.statePre, stateHash}, item, false, true, true)
	return
}

func (self *ObjPoint) GetObj(id []byte, item PItem) (ret CItem) {

	if v := self.cons.getObj(&key{self.objPre, id}, &Bytes{}, true, false, false); v != nil {
		stateHash := *v.(*Bytes)
		if v := self.cons.getObj(&key{self.statePre, stateHash}, item, false, true, true); v != nil {
			return v.(CItem)
		}
	}
	return nil
}
