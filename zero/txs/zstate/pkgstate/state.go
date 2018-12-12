package pkgstate

import (
	"fmt"
	"sort"
	"sync"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-sero/rlp"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/zero/txs/stx"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-sero/zero/txs/pkg"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type ZPkg struct {
	High uint64
	From keys.Uint512
	Pack stx.PkgCreate
}

func (self *ZPkg) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}

type PkgGet struct {
	out *ZPkg
}

func (self *PkgGet) Out() *ZPkg {
	return self.out
}

func (self *PkgGet) Unserial(v []byte) (e error) {
	if len(v) < 2 {
		self.out = nil
		return
	} else {
		self.out = &ZPkg{}
		if err := rlp.DecodeBytes(v, &self.out); err != nil {
			e = err
			self.out = nil
			return
		} else {
			return
		}
	}
}

type PkgState struct {
	tri          tri.Tri
	rw           *sync.RWMutex
	num          uint64
	G2pkgs       map[keys.Uint256]*ZPkg
	G2pkgs_dirty map[keys.Uint256]bool
}

func NewPkgState(tri tri.Tri, num uint64) (state PkgState) {
	state = PkgState{tri: tri, num: num}
	state.rw = new(sync.RWMutex)
	state.clear()
	return
}

func (self *PkgState) Update() {
	G2pkgs_dirty := utils.Uint256s{}
	for k := range self.G2pkgs_dirty {
		G2pkgs_dirty = append(G2pkgs_dirty, k)
	}
	sort.Sort(G2pkgs_dirty)

	for _, k := range G2pkgs_dirty {
		v := self.G2pkgs[k]
		tri.UpdateObj(self.tri, pkgName(&k), v)
	}
	return
}

func (self *PkgState) Revert() {
	self.clear()
	return
}

func (state *PkgState) clear() {
	state.G2pkgs = make(map[keys.Uint256]*ZPkg)
	state.G2pkgs_dirty = make(map[keys.Uint256]bool)
}

func (state *PkgState) add_pkg_dirty(pkg *ZPkg) {
	state.G2pkgs[pkg.Pack.Id] = pkg
	state.G2pkgs_dirty[pkg.Pack.Id] = true
}

func (state *PkgState) del_pkg_dirty(id *keys.Uint256) {
	state.G2pkgs[*id] = nil
	state.G2pkgs_dirty[*id] = true
}

func pkgName(k *keys.Uint256) (ret []byte) {
	ret = []byte("ZState0_PkgName")
	ret = append(ret, k[:]...)
	return
}

func (state *PkgState) getPkg(id *keys.Uint256) (pg *ZPkg) {
	if pg = state.G2pkgs[*id]; pg != nil {
		return
	} else {
		get := PkgGet{}
		tri.GetObj(state.tri, pkgName(id), &get)
		pg = get.Out()
		return
	}
}

func (self *PkgState) GetPkg(id *keys.Uint256) (pg *ZPkg) {
	self.rw.Lock()
	defer self.rw.Unlock()
	return self.getPkg(id)
}

func (self *PkgState) Force_del(id *keys.Uint256) {
	self.rw.Lock()
	defer self.rw.Unlock()
	self.del_pkg_dirty(id)
}

func (self *PkgState) Force_add(from *keys.Uint512, pack *stx.PkgCreate) {
	self.rw.Lock()
	defer self.rw.Unlock()
	zpkg := ZPkg{
		self.num,
		*from,
		pack.Clone(),
	}
	self.add_pkg_dirty(&zpkg)
}

func (self *PkgState) Force_transfer(id *keys.Uint256, to *keys.Uint512) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.getPkg(id); pg == nil {
		return
	} else {
		pg.Pack.Pkr = *to
		self.add_pkg_dirty(pg)
		return
	}
}

type OPkg struct {
	Z ZPkg
	O pkg.Pkg_O
}

func (self *PkgState) Close(id *keys.Uint256, pkr *keys.Uint512, key *keys.Uint256) (ret OPkg, e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.getPkg(id); pg == nil {
		e = fmt.Errorf("Pkg is nil: %v", hexutil.Encode(id[:]))
		return
	} else {
		if pg.Pack.Pkr != *pkr {
			e = fmt.Errorf("Pkg Owner Check Failed: %v", hexutil.Encode(id[:]))
			return
		} else {
			if ret.O, e = pkg.DePkg(key, &pg.Pack.Pkg); e != nil {
				return
			} else {
				ret.Z = *pg
				self.del_pkg_dirty(id)
				return
			}
		}
	}
}

func (self *PkgState) Transfer(id *keys.Uint256, pkr *keys.Uint512, to *keys.Uint512) (e error) {
	self.rw.Lock()
	defer self.rw.Unlock()
	if pg := self.getPkg(id); pg == nil {
		e = fmt.Errorf("Pkg is nil: %v", hexutil.Encode(id[:]))
		return
	} else {
		if pg.Pack.Pkr != *pkr {
			e = fmt.Errorf("Pkg Owner Check Failed: %v", hexutil.Encode(id[:]))
			return
		} else {
			pg.Pack.Pkr = *to
			self.add_pkg_dirty(pg)
			return
		}
	}
}
