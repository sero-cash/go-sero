package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

func (self *Data) Add(pkg *localdb.ZPkg) {
	hash := pkg.ToHash_V1()
	self.Id2Hash.Add(&pkg.Pack.Id, &hash)
	self.Dirtys.Append(&pkg.Pack.Id)
}

func (self *Data) Del(id *keys.Uint256) {
	self.Id2Hash.Del(id)
	self.Dirtys.Append(id)
}

func (self *Data) SaveIndex(tr tri.Tri) {
	G2pkgs_dirty := self.Dirtys.List()
	for _, k := range G2pkgs_dirty {
		self.Id2Hash.Save(tr, &k)
	}
}

func (state *Data) GetHash(tr tri.Tri, id *keys.Uint256) (ret *keys.Uint256) {
	return state.Id2Hash.Get(tr, id)
}
