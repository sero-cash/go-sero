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

package witness

import (
	"testing"

	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

var genLeaf = merkle.GenLeaf
var getLeaf = merkle.GetLeaf

func test(v interface{}) {}

func TestMerklePath(t *testing.T) {
	tree := merkle.Tree{}
	pgs := []*PathGen{}
	wits := []*Witness{}
	for i := uint64(0); i < 100; i++ {
		leaf := merkle.Leaf{uint8(i)}
		for _, wit := range wits {
			wit.Append(leaf)
		}
		tree.Append(leaf)
		w := Witness{Tree: tree.Clone()}

		pg, roots := NewPathGenAndRoots(&tree)
		icur := NewIndexCur(&pg)
		for i := len(pgs) - 1; i > -1; i-- {
			wpt := pgs[i]
			NextPathGen(&icur, wpt, &roots)
			path_w, index_w := wits[i].Path()
			anchor_w := wits[i].Root()
			root_w := wits[i].Tree.Root()

			if root_w != wpt.Root {
				t.Fail()
			}

			if anchor_w != wpt.Anchor {
				t.Fail()
			}

			for j, w := range path_w {
				if wpt.Path[j] != w {
					t.Fail()
				}
			}

			if index_w != wpt.Index {
				t.Fail()
			}
		}

		pgs = append(pgs, &pg)
		wits = append(wits, &w)

	}
}

func TestMerkle(t *testing.T) {
	tree1 := merkle.Tree{}
	w1 := Witness{Tree: tree1.Clone()}
	m1 := make(map[merkle.Leaf]int)
	for i := 0; i < 100000; i++ {
		tree1.Append(merkle.Leaf{1})
		w1.Append(merkle.Leaf{1})
		t_root := tree1.Root()
		w_root := w1.Root()
		if t_root != w_root {
			panic("")
		}
		if v, ok := m1[t_root]; ok {
			panic(v)
		} else {
			m1[t_root] = i
		}
	}

	var tree merkle.Tree
	tree.Append(genLeaf())
	tree.Append(genLeaf())
	tree.Append(genLeaf())
	tree.Append(genLeaf())
	tree.Append(genLeaf())
	tree.Append(genLeaf())
	w := Witness{Tree: tree.Clone()}
	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root := tree.Root()
	w_root := w.Root()
	t.Log(t_root, w_root)

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	w.Append(genLeaf())
	tree.Append(getLeaf())
	t_root = tree.Root()
	w_root = w.Root()

	root := w.Root()
	t.Logf("root:%v", root)
	path := PartialPath(w)
	t.Logf("path:%v", path)
	last := merkle.Last(w.Tree)
	t.Logf("last:%v", last)
	elem := w.Element()
	t.Logf("elem:%v", elem)
	p, index := w.Path()
	t.Logf("path:%v", p)
	t.Logf("index:%v", index)
}
