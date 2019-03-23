package pkgstate

import (
	"fmt"
	"sync"

	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/pkgstate/data"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type PkgState struct {
	tri tri.Tri
	rw  *sync.RWMutex
	num uint64

	data      data.Data
	snapshots utils.Snapshots
}

func NewPkgState(tri tri.Tri, num uint64) (state PkgState) {
	state = PkgState{tri: tri, num: num}
	state.data = *data.NewData(num)
	state.rw = new(sync.RWMutex)
	state.data.Clear()
	state.load()
	return
}

func (self *PkgState) Snapshot(revid int) {
	self.snapshots.Push(revid, &self.data)
}
func (self *PkgState) Revert(revid int) {
	self.data.Clear()
	self.data = *self.snapshots.Revert(revid).(*data.Data)
	return
}

func (self *PkgState) load() {
	self.data.LoadCur(self.tri)
}

func (self *PkgState) Update() {
	self.data.SaveIndex(self.tri)
	self.data.SaveCur(self.tri)
	return
}

func (state *PkgState) GetBlockDetails() (ret []keys.Uint256) {
	return state.data.Dirtys.List()
}

func (self *PkgState) GetPkg(id *keys.Uint256) (pg *localdb.ZPkg) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if hash := self.data.Id2Hash.Get(self.tri, id); hash != nil {
		pg = localdb.GetPkg(self.tri.GlobalGetter(), hash)
		return
	} else {
		return
	}
}

func (self *PkgState) Force_del(id *keys.Uint256) {
	self.rw.Lock()
	defer self.rw.Unlock()
	self.data.Del(id)
}

func (self *PkgState) Force_add(from *keys.PKr, pack *stx.PkgCreate) {
	self.rw.Lock()
	defer self.rw.Unlock()
	zpkg := localdb.ZPkg{
		self.num,
		*from,
		pack.Clone(),
	}
	self.data.Add(&zpkg)
}

func (self *PkgState) Force_transfer(id *keys.Uint256, to *keys.PKr) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.GetPkg(id); pg == nil {
		return
	} else {
		pg.Pack.PKr = *to
		self.data.Add(pg)
		return
	}
}

type OPkg struct {
	Z localdb.ZPkg
	O pkg.Pkg_O
}

func (self *PkgState) Close(id *keys.Uint256, pkr *keys.PKr, key *keys.Uint256) (ret OPkg, e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.GetPkg(id); pg == nil {
		e = fmt.Errorf("Pkg is nil: %v", hexutil.Encode(id[:]))
		return
	} else {
		if pg.Pack.PKr != *pkr {
			e = fmt.Errorf("Pkg Owner Check Failed: %v", hexutil.Encode(id[:]))
			return
		} else {
			if ret.O, e = pkg.DePkg(key, &pg.Pack.Pkg); e != nil {
				return
			} else {
				ret.Z = *pg
				if e = pkg.ConfirmPkg(&ret.O, &ret.Z.Pack.Pkg); e != nil {
					return
				} else {
					self.data.Del(id)
					return
				}
			}
		}
	}
}

func (self *PkgState) Transfer(id *keys.Uint256, pkr *keys.PKr, to *keys.PKr) (e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.GetPkg(id); pg == nil {
		e = fmt.Errorf("Pkg is nil: %v", hexutil.Encode(id[:]))
		return
	} else {
		if pg.Pack.PKr != *pkr {
			e = fmt.Errorf("Pkg Owner Check Failed: %v", hexutil.Encode(id[:]))
			return
		} else {
			pg.Pack.PKr = *to
			self.data.Add(pg)
			return
		}
	}
}
