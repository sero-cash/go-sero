package witness

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

type PathGen struct {
	Leaf  merkle.Leaf
	Path  [merkle.DEPTH]merkle.Leaf
	Index uint64
}

func GenRoots(pg *PathGen) (roots [merkle.DEPTH + 1]merkle.Leaf) {
	roots[0] = pg.Leaf
	for j := uint64(1); j <= merkle.DEPTH; j++ {
		flag := pg.Index >> (j - 1)
		flag &= 0x1
		p := pg.Path[merkle.DEPTH-j]
		r := roots[j-1]
		if flag == 0 {
			roots[j] = merkle.Leaf(cpt.Combine(r.ToUint256(), p.ToUint256()))
		} else {
			roots[j] = merkle.Leaf(cpt.Combine(p.ToUint256(), r.ToUint256()))
		}
	}
	return
}

func GenRoot(pg *PathGen) merkle.Leaf {
	roots := GenRoots(pg)
	return roots[merkle.DEPTH]
}

func NewPathGen(tr *merkle.Tree) (pg PathGen, roots [merkle.DEPTH + 1]merkle.Leaf) {
	l := merkle.Last(*tr)
	w := Witness{Tree: tr.Clone()}
	path, index := w.Path()
	pg.Leaf = l
	pg.Index = index
	copy(pg.Path[:], path[:])

	roots = GenRoots(&pg)
	return
}

type IndexCur struct {
	Left  uint64
	Count uint64
}

func NewIndexCur(gen *PathGen) (ret IndexCur) {
	ret.Left = gen.Index
	ret.Count = 0
	return
}

func ParseIndex(pc *IndexCur, index uint64) uint64 {
	for {
		c_index := index >> pc.Count
		if c_index == pc.Left {
			return pc.Count - 1
		} else {
			pc.Left >>= 1
			pc.Count++
		}
		if pc.Count > merkle.DEPTH {
			panic("Index depth > Merkle.DEPTH")
		}
	}
}

func NextPathGen(pc *IndexCur, pg *PathGen, roots *[merkle.DEPTH + 1]merkle.Leaf) {
	start := ParseIndex(pc, pg.Index)
	pg.Path[merkle.DEPTH-start-1] = roots[start]
}
