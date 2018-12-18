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

package zstate

import (
	"runtime"
	"runtime/debug"

	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

func Need_debug() bool {
	return false
	if false {
		return true
	} else {
		return zconfig.Is_Dev()
	}
}

func Debug_State0_addout_assert(state *State0, os *OutState0) {
	if Need_debug() {
		trees := state.GenState0Trees()
		leaf := os.ToRootCM()
		tree := trees.Trees[os.Index]
		root := tree.RootKey()
		if out, err := state.GetOut(&root); err != nil {
			Debug_Weak_panic("Debug: add out get out by root err: %v", err)
		} else {
			if out != nil {
				Debug_Weak_panic("Debug: add out get out by root is not nil: %v\n%v\n", out, os)
			} else {
			}
		}

		if out, err := state.GetOut(leaf); err != nil {
			Debug_Weak_panic("get out by leaf err: %v", err)
		} else {
			if out != nil {
				Debug_Weak_panic("get out by leaf is not nil: %v\n%v\n", out, os)
			} else {
			}
		}
	}
}

func Debug_Weak_panic(msg string, ctx ...interface{}) {
	if Need_debug() {
		log.Debug(">========debug_painc:=======>"+msg, ctx...)
		debug.PrintStack()
		runtime.Breakpoint()
	}
}
