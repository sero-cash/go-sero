package lstate

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

func Debug_State1_addout_assert(state *State, os *zstate.OutState) {
	if zstate.Need_debug() {
		wmap := make(map[keys.Uint256]int)
		for i, wout := range state.G2wouts {
			if v, ok := wmap[wout]; ok {
				zstate.Debug_Weak_panic("add out but wouts already exists i,v:%v,%v", i, v)
			} else {
				wmap[wout] = i
			}
		}
		trees := state.State0.GenState0Trees()
		leaf := os.ToRootCM()
		tree := trees.Trees[os.Index]
		root := tree.RootKey()
		if out, err := state.GetOut(&root); err != nil {
			zstate.Debug_Weak_panic("get out err: %v", err)
		} else {
			if out != nil {
				zstate.Debug_Weak_panic("get out but out is not nil %v", out)
			} else {
			}
		}
		if out, err := state.GetOut(leaf); err != nil {
			zstate.Debug_Weak_panic("get out by leaf err: %v", err)
		} else {
			if out != nil {
				zstate.Debug_Weak_panic("get out by leaf but out is not nil %v", out)
			} else {
			}
		}
	}
}

func Debug_State1_addout_end_assert(state *State, os *zstate.OutState) {
	if zstate.Need_debug() {
		trees := state.State0.GenState0Trees()

		leaf := os.ToRootCM()
		tree := trees.Trees[os.Index]
		root := tree.RootKey()

		for i, wout := range state.G2wouts {
			if out, err := state.GetOut(&wout); err != nil {
				zstate.Debug_Weak_panic("get out err: %v,%v", err, i)
			} else {
				w_root := out.Pg.Anchor
				w_el := out.Pg.Leaf.ToUint256()
				if w_root != merkle.Leaf(root) {
					zstate.Debug_Weak_panic("w_root!=root")
				}
				if *w_el != *leaf {
					zstate.Debug_Weak_panic("w_el!=leaf")
				}
			}
		}
	}
}
