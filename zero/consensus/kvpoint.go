package consensus

type KVPoint struct {
	keyPre string
	cons   *Cons
}

func NewKVPt(cons *Cons, keyPre string, inblock string) (ret KVPoint) {
	ret.keyPre = keyPre
	ret.cons = cons
	return
}

func (self *KVPoint) SetValue(id []byte, value []byte) {
	v := Bytes(value)
	self.cons.addObj(&key{self.keyPre, id}, &v, true, nil, nil)
	return
}

func (self *KVPoint) GetValue(id []byte) (ret []byte) {
	if v := self.cons.getObj(&key{self.keyPre, id}, &Bytes{}, true, false); v != nil && v.item != nil {
		ret = append([]byte{}, *v.item.(*Bytes)...)
		return
	}
	return nil
}
