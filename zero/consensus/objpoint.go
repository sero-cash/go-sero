package consensus

import "errors"

type ObjPoint struct {
	objPre   string
	statePre string
	inblock  string
	cons     *Cons
}

func NewObjPt(cons *Cons, objPre string, statePre string, inblock string) (ret ObjPoint) {
	ret.objPre = objPre
	ret.statePre = statePre
	ret.inblock = inblock
	ret.cons = cons
	return
}

func (self *ObjPoint) AddObj(item PItem) {
	if item == nil {
		panic(errors.New("item can not be nil"))
	}
	stateHash := Bytes(item.State())
	self.cons.addObj(&key{self.objPre, item.Id()}, &stateHash, true, nil, nil)
	self.cons.addObj(&key{self.statePre, stateHash}, item, false, &inBlock{self.inblock, item.Id()}, &inDB{item.Id()})
	return
}

func (self *ObjPoint) GetObj(id []byte, item PItem) (ret CItem) {

	if v := self.cons.getObj(&key{self.objPre, id}, &Bytes{}, true, false); v != nil && v.item != nil {
		stateHash := *v.item.(*Bytes)
		if v := self.cons.getObj(&key{self.statePre, stateHash}, item, false, true); v != nil && v.item != nil {
			return v.item.(CItem)
		}
	}
	return nil
}
