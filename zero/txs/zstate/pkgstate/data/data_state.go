package data

import (
	"errors"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

func (self *Data) SaveState(tr tri.Tri) {
	G2pkgs_dirty := self.IdDirtys.List()
	for _, k := range G2pkgs_dirty {
		self.Id2Hash.SaveByTri(tr, &k)
	}
}

func (self *Data) RecordState(putter serodb.Putter, hash *c_type.Uint256) {
	if pkg, ok := self.Hash2Pkg[*hash]; ok {
		localdb.PutPkg(putter, hash, &pkg)
	} else {
		panic(errors.New("PKG record index error: hash2pkg"))
	}
}

func (self *Data) GetPkgById(tr tri.Tri, id *c_type.Uint256) (pg *localdb.ZPkg) {
	if hash := self.Id2Hash.GetByTri(tr, id); hash != c_type.Empty_Uint256 {
		pg = self.GetPkgByHash(tr, &hash)
		return
	} else {
		return
	}
}

func (self *Data) GetPkgByHash(tr tri.Tri, hash *c_type.Uint256) (pg *localdb.ZPkg) {
	if p, ok := self.Hash2Pkg[*hash]; ok {
		pg = &p
		return
	} else {
		if p := localdb.GetPkg(tr.GlobalGetter(), hash); p != nil {
			self.Hash2Pkg[*hash] = *p
			pg = p
			return
		} else {
			return
		}
	}
}
