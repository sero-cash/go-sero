package assets

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Asset struct {
	Tkn *Token   `rlp:"nil"`
	Tkt *Ticket  `rlp:"nil"`
	Pkg *Package `rlp:"nil"`
}

func NewAsset(tkn *Token, tkt *Ticket, pkg *Package) (ret Asset) {
	if tkn != nil {
		if tkn.Value.Cmp(&utils.U256_0) > 0 {
			ret.Tkn = tkn.Clone().ToRef()
		}
	}
	if tkt != nil {
		if tkt.Value != keys.Empty_Uint256 {
			ret.Tkt = tkt.Clone().ToRef()
		}
	}
	if pkg != nil {
		ret.Pkg = pkg.Clone().ToRef()
	}
	return
}

func (self *Asset) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	if self.Tkn != nil {
		d.Write(self.Tkn.ToHash().NewRef()[:])
	}
	if self.Tkt != nil {
		d.Write(self.Tkt.ToHash().NewRef()[:])
	}
	if self.Pkg != nil {
		d.Write(self.Pkg.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Asset) Clone() (ret Asset) {
	utils.DeepCopy(&ret, self)
	return
}
