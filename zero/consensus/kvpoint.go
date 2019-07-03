package consensus

type KVPoint struct {
	keyPre string
	cons   *Cons
}

func NewKVPt(cons *Cons, keyPre string) (ret KVPoint) {
	ret.keyPre = keyPre
	ret.cons = cons
	return
}

func (self *KVPoint) SetValue(id []byte, value []byte) {
	v := Bytes(value)
	self.cons.addObj(&key{self.keyPre, id}, &v, true, false, false)
	return
}

func (self *KVPoint) GetValue(id []byte) (ret []byte) {

	if v := self.cons.getObj(&key{self.keyPre, id}, &Bytes{}, true, false, false); v != nil {
		ret = append([]byte{}, *v.(*Bytes)...)
		return
	}
	return nil
}
