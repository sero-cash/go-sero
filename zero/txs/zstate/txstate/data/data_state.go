package data

import (
	"fmt"
	"sort"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
)

const LAST_OUTSTATE0_NAME = tri.KEY_NAME("ZState0_Cur")
const BLOCK_NAME = "ZState0_BLOCK"
const ZSTATE0_INNAME = "ZState0_InName"
const ZSTATE0_OUTNAME = "ZState0_OutName"

func Name2BKey(name string, num uint64) (ret []byte) {
	key := fmt.Sprintf("%s_%d", name, num)
	ret = []byte(key)
	return
}
func InName(k *keys.Uint256) (ret []byte) {
	ret = []byte(ZSTATE0_INNAME)
	ret = append(ret, k[:]...)
	return
}
func OutName0(k *keys.Uint256) (ret []byte) {
	ret = []byte(ZSTATE0_OUTNAME)
	ret = append(ret, k[:]...)
	return
}

func (self *Data) RecordState(putter serodb.Putter, root *keys.Uint256) {
	if int64(self.Num) > int64(seroparam.SIP2())-13000 {
		if out, ok := self.G2outs[*root]; ok {
			rs := localdb.RootState{}
			rs.Num = self.Num
			rs.OS = *out
			if txhash, ok := self.H2tx[*root]; ok {
				if txhash != nil {
					rs.TxHash = *txhash
				}
				localdb.PutRoot(putter, root, &rs)
			} else {
				panic(fmt.Errorf("data record state h2tx error for : %v", self.Num))
			}
		} else {
			panic(fmt.Errorf("data record state G2outs error for : %v", self.Num))
		}
	}
	return
}

func (self *Data) LoadState(tr tri.Tri) {
	if self.Num >= 0 {
		get := CurrentGet{}
		tri.GetObj(
			tr,
			LAST_OUTSTATE0_NAME.Bytes(),
			&get,
		)
		self.Cur = get.Out

		blockget := State0BlockGet{}
		tri.GetObj(
			tr,
			Name2BKey(BLOCK_NAME, uint64(self.Num)),
			&blockget,
		)
		self.Block = blockget.Out
	}
	return
}

func (self *Data) SaveState(tr tri.Tri) {
	tri.UpdateObj(tr, LAST_OUTSTATE0_NAME.Bytes(), &self.Cur)
	tri.UpdateObj(
		tr,
		Name2BKey(BLOCK_NAME, self.Num),
		&self.Block,
	)
	g2ins_dirty := utils.Uint256s{}
	for k := range self.Dirty_G2ins {
		g2ins_dirty = append(g2ins_dirty, k)
	}
	sort.Sort(g2ins_dirty)

	for _, k := range g2ins_dirty {
		v := []byte{1}
		if err := tr.TryUpdate(InName(&k), v); err != nil {
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
			tri.UpdateObj(tr, OutName0(&k), v)
		} else {
			panic("state0 update g2outs can not find dirty out")
		}
	}
	return
}

func (self *Data) HasIn(tr tri.Tri, hash *keys.Uint256) (exists bool) {
	if v, err := tr.TryGet(InName(hash)); err != nil {
		panic(err)
		return
	} else {
		if v != nil && v[0] == 1 {
			exists = true
		} else {
			exists = false
		}
	}
	return
}

func (self *Data) GetOut(tr tri.Tri, root *keys.Uint256) (src *localdb.OutState) {
	if out := self.G2outs[*root]; out != nil {
		return out
	} else {
		get := localdb.OutState0Get{}
		tri.GetObj(tr, OutName0(root), &get)
		if get.Out != nil {
			self.G2outs[*root] = get.Out
			return get.Out
		} else {
			return nil
		}
	}
}
