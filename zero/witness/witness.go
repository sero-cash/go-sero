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

package witness

import (
    "github.com/sero-cash/go-sero/zero/witness/merkle"
    "unsafe"
    "github.com/sero-cash/go-sero/zero/zconfig"
    "github.com/sero-cash/go-sero/zero/utils"
)

type WitnessBase struct {
    Filled       []merkle.Leaf
    Cursor       *merkle.Tree `rlp:"nil"`
    Cursor_depth uint
    Parents      []merkle.Leaf
    Logs []merkle.Leaf
}
type Witness struct {
    Tree merkle.Tree
    WitnessBase
}

func (w *Witness)Clone() (ret Witness) {
    utils.DeepCopy(&ret,w)
    return
}

func (w *Witness) Element() (merkle.Leaf) {
    return merkle.Last(w.Tree)
}

func PartialPath(w Witness)(*merkle.PathFiller) {
    uncles:=merkle.NewPathFilter(w.Filled)
    if w.Cursor !=nil {
        uncles.Push(w.Cursor.TempRoot(&merkle.PathFiller{},w.Cursor_depth))
    }
    return uncles
}


func (w *Witness) Root() (merkle.Leaf) {
    return w.Tree.TempRoot(PartialPath(*w),merkle.DEPTH)
}


func (w *Witness) Append(leaf merkle.Leaf) {
    if zconfig.Is_Dev() {
        w.Logs = append(w.Logs, leaf)
        if len(w.Logs) > 10 {
            w.Logs = append(w.Logs[:0], w.Logs[1:]...)
        }
    }
    if w.Cursor !=nil {
        w.Cursor.Append(leaf)
        if w.Cursor.TempIsComplete(w.Cursor_depth) {
            w.Filled =append(w.Filled,w.Cursor.TempRoot(PartialPath(*w),w.Cursor_depth))
            w.Cursor = nil;
        }
    } else {
        w.Cursor_depth =w.nextDepth()
        
        if w.Cursor_depth > merkle.DEPTH {
            panic("tree is full");
        }
        
        if w.Cursor_depth == 0 {
            w.Filled =append(w.Filled,leaf)
        } else {
            w.Cursor =new(merkle.Tree)
            w.Cursor.Append(leaf)
        }
    }
}
func (w* Witness) nextDepth() (uint) {
    return merkle.NextDepth(w.Tree,uint(len(w.Filled)))
}
func (w *Witness) IsComplete() (bool) {
    if w.nextDepth() == merkle.DEPTH {
        return true
    } else {
        return false
    }
}

func reverse(ps unsafe.Pointer) {
    s:=*(*[]interface{})(ps)
    for i,j:=0,len(s)-1; i<j; i,j=i+1,j-1 {
        s[i],s[j]=s[j],s[i]
    }
}

func bools2int(v []bool)(ret uint64) {
    if len(v)>64 {
        panic("boolean vector can't be larger than 64 bits")
    }
    for i:=0;i<len(v);i++ {
        if v[i] {
            ret |= uint64(1) << uint(((len(v) - 1) - i))
        }
    }
    return
}

func (w *Witness)Path()(path []merkle.Leaf,index uint64) {
    tree:=w.Tree
    if tree.Left==nil {
        panic("can't create an authentication path for the beginning of the tree");
    }
    
    filler:=PartialPath(*w)
    
    b_index:=[]bool{}
    if tree.Right!=nil {
        b_index=append(b_index,true)
        path=append(path,*tree.Left)
    } else {
        b_index=append(b_index,false);
        path=append(path,filler.Next(0))
    }
    
    d:=uint(1);
    
    for _,parent:=range tree.Pats {
        if parent.L!=nil {
            b_index=append(b_index,true)
            path=append(path,*parent.L)
        } else {
            b_index=append(b_index,false)
            path=append(path,filler.Next(d))
        }
        d++;
    }
    
    for d<merkle.DEPTH {
        b_index=append(b_index,false)
        path=append(path,filler.Next(d))
        d++;
    }
    for i:=0;i<len(path)/2;i++ {
        path[i],path[len(path)-i-1]=path[len(path)-i-1],path[i]
    }
    for i:=0;i<len(b_index)/2;i++ {
        b_index[i],b_index[len(path)-i-1]=b_index[len(path)-i-1],b_index[i]
    }
    index=bools2int(b_index)
    return
}