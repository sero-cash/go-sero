package data

import "github.com/sero-cash/go-sero/zero/utils"

type Data struct {
	Num     uint64
	Dirtys  utils.Dirtys
	Id2Hash utils.H2Hash
}

func NewData(num uint64) (ret *Data) {
	return &Data{
		Num:     num,
		Id2Hash: utils.NewH2Hash("$ZState0_PkgName$"),
	}
}
func (self *Data) Clear() {
	self.Dirtys.Clear()
	self.Id2Hash.Clear()
}
