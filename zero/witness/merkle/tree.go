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
    "github.com/sero-cash/go-czero-import/keys"
    "github.com/sero-cash/go-czero-import/cpt"
    "container/list"
    "github.com/sero-cash/go-sero/zero/zconfig"
    "github.com/sero-cash/go-sero/zero/utils"
)


//====================
const (
    DEPTH=cpt.DEPTH
)

type Leaf keys.Uint256

func (self *Leaf) ToUint256() (ret *keys.Uint256){
    ret=&keys.Uint256{}
    copy(ret[:],self[:])
    return
}

func Combine(l *Leaf,r *Leaf) (ret Leaf) {
    
    ret=Leaf(cpt.Combine(
        l.ToUint256(),
        r.ToUint256(),
    ))
    return
}

var empty=[32]byte{}

var count byte=0;

func GenLeaf() Leaf {
    count++
    return Leaf{count}
}
func GetLeaf() Leaf {
    return Leaf{count}
}
//====================


/*
//=============
const (
    DEPTH=8
)

type Leaf string

func Combine(l *Leaf,r *Leaf) (ret Leaf) {
    return "("+*l+"+"+*r+")"
}

var empty=Leaf("#")

var count byte=0;

func GenLeaf() Leaf {
    count++
    return Leaf(fmt.Sprintf("%v",count))
}

func GetLeaf() Leaf {
    return Leaf(fmt.Sprintf("%v",count))
}
//=============
*/


type Parent struct {
    L *Leaf     `rlp:"nil"`
}
type Tree struct {
    Left    *Leaf       `rlp:"nil"`
    Right   *Leaf       `rlp:"nil"`
    Pats []Parent
    Logs []Leaf
}

func (t *Tree) Clone() (ret Tree) {
    utils.DeepCopy(&ret,t)
    return
}

func createEmpty()(ret [DEPTH+1]Leaf) {
    ret[0]=empty
    for i:=1;i<=DEPTH;i++ {
        ret[i]=Combine(&ret[i-1],&ret[i-1])
    }
    return
}

var EmptyRoots=createEmpty()

func (t* Tree) IsComplete()(bool) {
    return t.TempIsComplete(DEPTH)
}

func (t* Tree) TempIsComplete(depth uint)(bool) {
    if t.Left ==nil||t.Right ==nil {
        return false
    }
    if len(t.Pats)!=int(depth-1) {
        return false
    }
    for _,parent:=range t.Pats {
        if parent.L==nil {
            return false
        }
    }
    return true
}
func (t *Tree) Append(l Leaf)() {
    if(t.IsComplete()) {
        panic("tree is full")
    }
    if zconfig.Is_Dev() {
        t.Logs = append(t.Logs, l)
        if len(t.Logs) > 10 {
            t.Logs = append(t.Logs[:0], t.Logs[1:]...)
        }
    }
    if t.Left ==nil {
        t.Left =&l
    } else if t.Right ==nil {
        t.Right =&l
    } else {
        combined:=Combine(t.Left,t.Right)
        t.Left =&l
        t.Right =nil
        for i:=0;i<DEPTH;i++ {
            if i<len(t.Pats) {
                if t.Pats[i].L!=nil {
                    combined=Combine(t.Pats[i].L,&combined);
                    t.Pats[i].L=nil
                } else {
                    t.Pats[i]=Parent{&combined}
                    break
                }
            } else {
                t.Pats =append(t.Pats,Parent{&combined})
                break
            }
        }
    }
}

func Size(t Tree) (ret uint) {
    ret=0;
    if t.Left !=nil {
        ret++
    }
    if t.Right !=nil {
        ret++
    }
    for i,parent:=range t.Pats {
        if parent.L!=nil {
            ret+=(uint(1)<<uint(i+1))
        }
    }
    return
}

func Last(t Tree) (ret Leaf) {
    if t.Right !=nil {
        return *t.Right
    }
    if t.Left !=nil {
        return *t.Left
    }
    panic("tree has no cursor");
}

func NextDepth(t Tree,skip uint)(ret uint) {
    if t.Left ==nil {
        if skip>0 {
            skip--
        } else {
            return 0
        }
    }
    if t.Right ==nil {
        if skip>0 {
            skip--
        } else {
            return 0
        }
    }
    d:=uint(1)
    for _,parent:=range t.Pats {
        if parent.L==nil {
            if skip>0 {
                skip--
            } else {
                return d;
            }
        }
        d++
    }
    return d+skip
}

type PathFiller struct {
    list.List
}

func NewPathFilter(leafs []Leaf)(*PathFiller) {
    pf:=&PathFiller{}
    for _,leaf:=range leafs {
        pf.Push(leaf)
    }
    return pf
}
func (pf* PathFiller) Push(leaf Leaf) {
    pf.PushBack(leaf)
}
func (pf* PathFiller) Next(depth uint) (Leaf) {
    if pf.List.Len()>0 {
        e:=pf.Front()
        pf.Remove(e)
        return e.Value.(Leaf)
    } else {
        return EmptyRoots[depth]
    }
}

func (tree *Tree) Root() (Leaf) {
    return tree.TempRoot(&PathFiller{},DEPTH)
}
func (tree *Tree) RootKey() (ret keys.Uint256) {
    root:=tree.Root()
    copy(ret[:],root[:])
    return
}

func (tree *Tree) TempRoot(partial *PathFiller,depth uint) (Leaf) {
    filler:=partial
    
    var combine_left Leaf
    if tree.Left!=nil {
        combine_left=*tree.Left
    } else {
        combine_left=filler.Next(0)
    }
    var combine_right Leaf
    if tree.Right!=nil {
        combine_right=*tree.Right
    } else {
        combine_right=filler.Next(0)
    }
    
    root:=Combine(&combine_left,&combine_right)
    
    d:=uint(1)
    
    for _,parent:=range tree.Pats {
        if parent.L!=nil {
            root=Combine(parent.L,&root)
        } else {
            next:=filler.Next(d)
            root=Combine(&root,&next)
        }
        d++
    }
    
    for d<depth {
        next:=filler.Next(d)
        root=Combine(&root,&next)
        d++
    }
    return root
}
