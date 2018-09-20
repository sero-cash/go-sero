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

package merkle

import (
    "testing"
)

/*
func TestMerkle(t *testing.T) {
    var tree Tree
    tree.Append(Leaf("1"));
    tree.Append(Leaf("2"));
    tree.Append(Leaf("3"));
    tree.Append(Leaf("4"));
    tree.Append(Leaf("5"));
    tree.Append(Leaf("6"));
    tree.Append(Leaf("7"));
    tree.Append(Leaf("8"));
    tree.Append(Leaf("9"));
    tree.Append(Leaf("10"));
    tree.Append(Leaf("11"));
    tree.Append(Leaf("12"));
    tree.Append(Leaf("13"));
    tree.Append(Leaf("14"));
    depth:=NextDepth(tree,0)
    t.Logf("%v",depth);
    if depth!=2 {
        t.Fail()
    }
    size:=Size(tree);
    if size!=14 {
        t.Fail()
    }
    t.Logf("%v",size);
}
*/

type BigStruct struct {
    C01 uint64
    C02 uint64
    C03 uint64
    C04 uint64
    C05 uint64
    C06 uint64
    C07 uint64
    C08 uint64
    C09 uint64
    C10 uint64
    C11 uint64
    C12 uint64
    C13 uint64
    C14 uint64
    C15 uint64
    C16 uint64
    C17 uint64
    C18 uint64
    C19 uint64
    C20 uint64
    C21 uint64
    C22 uint64
    C23 uint64
    C24 uint64
    C25 uint64
    C26 uint64
    C27 uint64
    C28 uint64
    C29 uint64
    C30 uint64
}

func Invoke1() *BigStruct {
    return new(BigStruct)
}

func Invoke2() (BigStruct) {
    return BigStruct{}
}

func f(r uint64) {
}

func Benchmark_Invoke1(b *testing.B) {
    for i := 0; i < b.N; i++ {
        r:=Invoke1()
        f(r.C01)
    }
}

func Benchmark_Invoke2(b *testing.B) {
    for i := 0; i < b.N; i++ {
        r:=Invoke2()
        f(r.C01)
    }
}
