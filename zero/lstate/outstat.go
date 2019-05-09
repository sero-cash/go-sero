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

package lstate

import (
	"sort"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
)

type OutStatWrap struct {
	stat localdb.OutStat
	out  *OutState
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
		return left.stat.Value.Cmp(&right.stat.Value) < 0
	}
	if left.stat.Z != right.stat.Z {
		return !left.stat.Z
	}
	if left.out.OutIndex != right.out.OutIndex {
		return left.out.OutIndex < right.out.OutIndex
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

func UpdateOutStat(db serodb.Putter, out *OutState) {
	os := localdb.OutStat{}
	os.Z = out.Z
	if out.Out_O.Asset.Tkn != nil {
		os.Value = out.Out_O.Asset.Tkn.Value
	} else {
		os.Value = utils.U256_0
	}
	localdb.UpdateOutStat(db, &out.Root, &os)
}

func SortOutStats(db serodb.Getter, outs []*OutState) {
	wraps := OutStats{}
	for _, out := range outs {
		out_root := out.Root
		get := localdb.GetOutStat(db, &out_root)
		if get != nil {
			wraps = append(
				wraps,
				OutStatWrap{
					*get,
					out,
				},
			)
		} else {
			os := localdb.OutStat{}
			os.Z = out.Z
			if out.Out_O.Asset.Tkn != nil {
				os.Value = out.Out_O.Asset.Tkn.Value
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
