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

package zconfig

var is_dev = false

func Init_Dev(dev bool) {
	is_dev = dev
}
func Is_Dev() bool {
	return is_dev
}

var dir = ""

var last_remove_time = int64(0)

func Init_State_dir(d string) {
	dir = d
	Init_State1()
	Init_State2()
}

var VP0 = uint64(788888)

var MAX_O_INS_LENGTH = 2500

var MAX_TX_OUT_COUNT_LENGTH = 256
