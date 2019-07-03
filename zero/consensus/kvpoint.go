package consensus

type KVPoint struct {
	keyPre  string
	inblock string
	cons    *Cons
}

func NewKVPt(cons *Cons, keyPre string, inblock string) (ret KVPoint) {
	ret.keyPre = keyPre
	ret.inblock = inblock
	ret.cons = cons
	return
}

func (self *KVPoint) SetValue(id []byte, value []byte) {
	v := Bytes(value)
	self.cons.addObj(&key{self.keyPre, id}, &v, true, self.inblock, false)
	return
}

func (self *KVPoint) GetValue(id []byte) (ret []byte) {
	if v := self.cons.getObj(&key{self.keyPre, id}, &Bytes{}, true, self.inblock, false); v != nil && v.item != nil {
		ret = append([]byte{}, *v.item.(*Bytes)...)
		return
	}
	return nil
}
