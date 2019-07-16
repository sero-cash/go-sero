package consensus

type BytePair struct {
	Key   []byte
	Value []byte
}

func (self *BytePair) CopyTo() CItem {
	ret := &BytePair{}
	ret.Key = append([]byte{}, self.Key...)
	ret.Value = append([]byte{}, self.Value...)
	return ret
}

func (self *BytePair) CopyFrom(from CItem) {
	if from != nil {
		f := from.(*BytePair)
		self.Key = append([]byte{}, f.Key...)
		self.Value = append([]byte{}, f.Value...)
	} else {
		self.Key = []byte{}
		self.Value = []byte{}
	}
	return
}

func (self *BytePair) Id() (ret []byte) {
	ret = append([]byte{}, self.Key...)
	return
}

func (self *BytePair) State() (ret []byte) {
	ret = append([]byte{}, self.Value...)
	return
}

type Bytes []byte

func (self *Bytes) CopyTo() CItem {
	ret := append(Bytes{}, *self...)
	return &ret
}

func (self *Bytes) CopyFrom(from CItem) {
	if from != nil {
		(*self) = append(Bytes{}, (*from.(*Bytes))...)
	} else {
		(*self) = []byte{}
	}
	return
}

func (self *Bytes) Id() (ret []byte) {
	ret = append(ret, *self...)
	return
}
