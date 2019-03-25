package data

import (
	"fmt"
	"sort"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
)

func (self *Data) addInByNilOrRoot(in *keys.Uint256) {
	self.G2ins[*in] = true
	self.Dirty_G2ins[*in] = true
}

func (self *Data) addOutByRoot(k *keys.Uint256, out *localdb.OutState) {
	self.G2outs[*k] = out
	self.Dirty_G2outs[*k] = true
}

func inName(k *keys.Uint256) (ret []byte) {
	ret = []byte("ZState0_InName")
	ret = append(ret, k[:]...)
	return
}
func outName0(k *keys.Uint256) (ret []byte) {
	ret = []byte("ZState0_OutName")
	ret = append(ret, k[:]...)
	return
}
func (self *Data) SaveIndex(tr tri.Tri) {
	g2ins_dirty := utils.Uint256s{}
	for k := range self.Dirty_G2ins {
		g2ins_dirty = append(g2ins_dirty, k)
	}
	sort.Sort(g2ins_dirty)

	for _, k := range g2ins_dirty {
		v := []byte{1}
		if err := tr.TryUpdate(inName(&k), v); err != nil {
			panic(err)
			return
		}
	}

	g2outs_dirty := utils.Uint256s{}
	for k := range self.Dirty_G2outs {
		g2outs_dirty = append(g2outs_dirty, k)
	}
	sort.Sort(g2outs_dirty)

	for _, k := range g2outs_dirty {
		if v := self.G2outs[k]; v != nil {
			tri.UpdateObj(tr, outName0(&k), v)
		} else {
			panic("state0 update g2outs can not find dirty out")
		}
	}
}

func (self *Data) AddOut(root *keys.Uint256, out *localdb.OutState) {
	self.addOutByRoot(root, out)
	self.appendRoot(root)
	if self.Cur.Index != int64(out.Index) {
		panic("add out but cur.index != current_index")
	}
	if self.Cur.Index < 0 {
		panic("add out but cur.index < 0")
	}
	return
}

func (self *Data) HasIn(tr tri.Tri, hash *keys.Uint256) (exists bool) {
	if v, ok := self.G2ins[*hash]; ok {
		exists = v
		return
	} else {
		if v, err := tr.TryGet(inName(hash)); err != nil {
			panic(err)
			return
		} else {
			if v != nil && v[0] == 1 {
				exists = true
			} else {
				exists = false
			}
			self.G2ins[*hash] = exists
		}
	}
	return
}

func (self *Data) AddIn(tr tri.Tri, nil_or_root *keys.Uint256) (e error) {
	if exists := self.HasIn(tr, nil_or_root); exists {
		e = fmt.Errorf("add in but exists")
		return
	} else {
		self.addInByNilOrRoot(nil_or_root)
		return
	}
}

func (self *Data) GetOut(tr tri.Tri, root *keys.Uint256) (src *localdb.OutState) {
	if out := self.G2outs[*root]; out != nil {
		return out
	} else {
		get := localdb.OutState0Get{}
		tri.GetObj(tr, outName0(root), &get)
		if get.Out != nil {
			self.G2outs[*root] = get.Out
			return get.Out
		} else {
			return nil
		}
	}
}
