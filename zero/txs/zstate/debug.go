// Copyright 2015 The sero.cash Authors
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

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
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
		leaf := os.ToCommitment()
		tree := trees.trees[os.Index]
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
		log_el := tree.Logs[len(tree.Logs)-1]
		log_leaf := log_el.ToUint256()
		if *log_leaf != *leaf {
			Debug_Weak_panic("")
		}
	}
}

func Debug_State1_addout_assert(state *State1, os *OutState0) {
	if Need_debug() {
		wmap := make(map[keys.Uint256]int)
		for i, wout := range state.G2wouts {
			if v, ok := wmap[wout]; ok {
				Debug_Weak_panic("add out but wouts already exists i,v:%v,%v", i, v)
			} else {
				wmap[wout] = i
			}
		}
		trees := state.State0.GenState0Trees()
		leaf := os.ToCommitment()
		tree := trees.trees[os.Index]
		root := tree.RootKey()
		if out, err := state.GetOut(&root); err != nil {
			Debug_Weak_panic("get out err: %v", err)
		} else {
			if out != nil {
				Debug_Weak_panic("get out but out is not nil %v", out)
			} else {
			}
		}
		if out, err := state.GetOut(leaf); err != nil {
			Debug_Weak_panic("get out by leaf err: %v", err)
		} else {
			if out != nil {
				Debug_Weak_panic("get out by leaf but out is not nil %v", out)
			} else {
			}
		}
	}
}

func Debug_State1_addout_end_assert(state *State1, os *OutState0) {
	if Need_debug() {
		trees := state.State0.GenState0Trees()

		leaf := os.ToCommitment()
		tree := trees.trees[os.Index]
		root := tree.RootKey()

		for i, wout := range state.G2wouts {
			if out, err := state.GetOut(&wout); err != nil {
				Debug_Weak_panic("get out err: %v,%v", err, i)
			} else {
				w_root := out.Witness.Root()
				w_leaf := out.Witness.Logs[len(out.Witness.Logs)-1]
				w_el := w_leaf.ToUint256()
				if w_root != merkle.Leaf(root) {
					Debug_Weak_panic("w_root!=root")
				}
				if *w_el != *leaf {
					Debug_Weak_panic("w_el!=leaf")
				}
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
