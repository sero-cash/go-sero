package pkgstate

import (
	"fmt"
	"sync"

	"github.com/sero-cash/go-sero/serodb"

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
}

func (self *PkgState) Update() {
	self.data.SaveState(self.tri)
	return
}

func (self *PkgState) RecordState(putter serodb.Putter, hash *keys.Uint256) {
	self.data.RecordState(putter, hash)
}

func (self *PkgState) GetPkgByHash(hash *keys.Uint256) (ret *localdb.ZPkg) {
	ret = self.data.GetPkgByHash(self.tri, hash)
	return
}

func (self *PkgState) GetPkgById(id *keys.Uint256) (ret *localdb.ZPkg) {
	ret = self.data.GetPkgById(self.tri, id)
	return
}

func (state *PkgState) GetPkgHashes() (ret []keys.Uint256) {
	return state.data.GetHashes()
}

func (self *PkgState) Force_del(hash *keys.Uint256, close *stx.PkgClose) (e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.data.GetPkgById(self.tri, &close.Id); pg == nil || pg.Closed {
		e = fmt.Errorf("Close Pkg is nil: %v", hexutil.Encode(close.Id[:]))
		return
	} else {
		if keys.VerifyPKr(hash, &close.Sign, &pg.Pack.PKr) {
			pg.Closed = true
			self.data.Add(pg)
		} else {
			e = fmt.Errorf("Close Pkg signed error: %v", hexutil.Encode(close.Id[:]))
			return
		}
		return
	}
}

func (self *PkgState) Force_add(from *keys.PKr, pack *stx.PkgCreate) (e error) {
	self.rw.Lock()
	defer self.rw.Unlock()

	if pg := self.data.GetPkgById(self.tri, &pack.Id); pg != nil {
		e = fmt.Errorf("Create Pkg is not nil: %v", hexutil.Encode(pack.Id[:]))
		return
	} else {
		zpkg := localdb.ZPkg{
			self.num,
			*from,
			pack.Clone(),
			false,
		}
		self.data.Add(&zpkg)
		return
	}

}

func (self *PkgState) Force_transfer(hash *keys.Uint256, trans *stx.PkgTransfer) (e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.data.GetPkgById(self.tri, &trans.Id); pg == nil || pg.Closed {
		e = fmt.Errorf("Transfer Pkg is nil: %v", hexutil.Encode(trans.Id[:]))
		return
	} else {
		if keys.VerifyPKr(hash, &trans.Sign, &pg.Pack.PKr) {
			pg.Pack.PKr = trans.PKr
			self.data.Add(pg)
		} else {
			e = fmt.Errorf("Transfer Pkg signed error: %v", hexutil.Encode(trans.Id[:]))
			return
		}
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
	if pg := self.data.GetPkgById(self.tri, id); pg == nil || pg.Closed {
		e = fmt.Errorf("Close Pkg is nil: %v", hexutil.Encode(id[:]))
		return
	} else {
		if pg.Pack.PKr != *pkr {
			e = fmt.Errorf("Close Pkg Owner Check Failed: %v", hexutil.Encode(id[:]))
			return
		} else {
			if ret.O, e = pkg.DePkg(key, &pg.Pack.Pkg); e != nil {
				return
			} else {
				ret.Z = *pg
				if e = pkg.ConfirmPkg(&ret.O, &ret.Z.Pack.Pkg); e != nil {
					return
				} else {
					pg.Closed = true
					self.data.Add(pg)
					return
				}
			}
		}
	}
}

func (self *PkgState) Transfer(id *keys.Uint256, pkr *keys.PKr, to *keys.PKr) (e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.data.GetPkgById(self.tri, id); pg == nil || pg.Closed {
		e = fmt.Errorf("Transfer Pkg is nil: %v", hexutil.Encode(id[:]))
		return
	} else {
		if pg.Pack.PKr != *pkr {
			e = fmt.Errorf("Transfer Pkg Owner Check Failed: %v", hexutil.Encode(id[:]))
			return
		} else {
			pg.Pack.PKr = *to
			self.data.Add(pg)
			return
		}
	}
}
