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

import "io/ioutil"

var is_dev = false

func Init_Dev(dev bool) {
	is_dev = dev
}
func Is_Dev() bool {
	return is_dev
}

func DefaultDelayNum() uint64 {
	if is_dev {
		return 0
	} else {
		return 12
	}
}

var dir = ""

var last_remove_time = int64(0)

func Init_State_dir(d string) {
	dir = d
}

var VP0 = uint64(788888)

var MAX_O_INS_LENGTH = 2500

var MAX_TX_OUT_COUNT_LENGTH = 256

func IsDirExists(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	} else {
		if len(files) > 0 {
			return true
		} else {
			return false
		}
	}
}
