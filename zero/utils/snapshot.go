package utils

import (
	"github.com/mohae/deepcopy"
)

type Snapshot struct {
	id  int
	buf interface{}
}

type Snapshots struct {
	objs []*Snapshot
}

func (self *Snapshots) Push(id int, obj interface{}) {
	cp := deepcopy.Copy(obj)
	self.objs = append(self.objs, &Snapshot{id, cp})
	return
}

func (self *Snapshots) Revert(id int) (to interface{}) {
	var temp []*Snapshot
	var max_temp *Snapshot
	for _, s := range self.objs {
		if s.id < id {
			temp = append(temp, s)
			if max_temp == nil {
				max_temp = s
			} else {
				if max_temp.id < s.id {
					max_temp = s
				}
			}
		}
	}
	self.objs = temp
	to = max_temp.buf
	return
}
