// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package state1

import (
	"sort"
	"time"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
)

type OutStat struct {
	Time  int64
	Value utils.U256
	Z     bool
}

func (self *OutStat) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type OutStatGet struct {
	out *OutStat
}

func (self *OutStatGet) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.out = nil
		return
	} else {
		self.out = &OutStat{}
		if err := rlp.DecodeBytes(v, self.out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}

type OutStatWrap struct {
	stat OutStat
	out  *OutState1
}

type OutStats []OutStatWrap

func (self OutStats) Len() int {
	return len(self)
}
func (self OutStats) Less(i, j int) bool {
	left := self[i]
	right := self[j]
	if left.stat.Time != right.stat.Time {
		return left.stat.Time < right.stat.Time
	}
	if left.stat.Value.Cmp(&right.stat.Value) != 0 {
		return left.stat.Value.Cmp(&right.stat.Value) > 0
	}
	if left.stat.Z != right.stat.Z {
		return !left.stat.Z
	}
	if left.out.Pg.Index != right.out.Pg.Index {
		return left.out.Pg.Index < right.out.Pg.Index
	}
	return false
}
func (self OutStats) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func outStatName(root *keys.Uint256) (ret []byte) {
	ret = []byte("ZSTATE_OUT_STAT_")
	ret = append(ret, root[:]...)
	return
}

func UpdateOutStat(st *zstate.State0, out *OutState1) {
	os := OutStat{}
	os.Z = out.Z
	if out.Out_O.Out.Pkg.Tkn != nil {
		os.Value = out.Out_O.Out.Pkg.Tkn.Value
	} else {
		os.Value = utils.U256_0
	}
	os.Time = time.Now().UnixNano()
	tri.UpdateGlobalObj(st.Tri(), outStatName(out.Pg.Root.ToUint256()), &os)
}

func SortOutStats(st *zstate.State0, outs []*OutState1) {
	wraps := OutStats{}
	for _, out := range outs {
		out_root := out.Pg.Root.ToUint256()
		get := OutStatGet{}
		tri.GetGlobalObj(st.Tri(), outStatName(out_root), &get)
		if get.out != nil {
			wraps = append(
				wraps,
				OutStatWrap{
					*get.out,
					out,
				},
			)
		} else {
			os := OutStat{}
			os.Z = out.Z
			if out.Out_O.Out.Pkg.Tkn != nil {
				os.Value = out.Out_O.Out.Pkg.Tkn.Value
			} else {
				os.Value = utils.U256_0
			}
			os.Time = 0
			wraps = append(
				wraps,
				OutStatWrap{
					os,
					out,
				},
			)
		}
	}
	sort.Sort(wraps)
	for i, wrap := range wraps {
		outs[i] = wrap.out
	}
}
