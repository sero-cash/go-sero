package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Data struct {
	Num      uint64
	Id2Hash  utils.H2Hash
	IdDirtys utils.Dirtys
	Hash2Pkg map[keys.Uint256]localdb.ZPkg
}

func NewData(num uint64) (ret *Data) {
	return &Data{
		Num:     num,
		Id2Hash: utils.NewH2Hash("$ZState0$Pkg$Id2Hash$"),
	}
}
func (self *Data) Clear() {
	self.Id2Hash.Clear()
	self.IdDirtys.Clear()
	self.Hash2Pkg = make(map[keys.Uint256]localdb.ZPkg)
}
func (self *Data) Add(pkg *localdb.ZPkg) {
	hash := pkg.ToHash()
	self.Hash2Pkg[hash] = *pkg
	self.Id2Hash.Add(&pkg.Pack.Id, &hash)
	self.IdDirtys.Append(&pkg.Pack.Id)
}

func (self *Data) GetHashes() (ret []keys.Uint256) {
	for _, id := range self.IdDirtys.List() {
		hash := self.Id2Hash.Get(&id)
		ret = append(ret, hash)
	}
	return
}
