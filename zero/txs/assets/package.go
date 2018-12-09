package assets

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/utils"
)

const (
	PACKAGE_STATE_CREATED = byte(0)
	PACKAGE_STATE_PACKED  = byte(1)
	PACKAGE_STATE_OPENED  = byte(2)
)

type Package struct {
	State      byte
	PKr        keys.Uint256
	LimitHeigh uint64
	Tkn        *Token  `rlp:"nil"`
	Tkt        *Ticket `rlp:"nil"`
}

func (this Package) ToRef() (ret *Package) {
	ret = &this
	return
}

func NewPackage(state uint8, pkr *keys.Uint256, limitHeigh uint64, tkn *Token, tkt *Ticket) (ret Package) {
	ret.State = state
	ret.PKr = *pkr
	ret.LimitHeigh = limitHeigh
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
	return
}

func (self *Package) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write([]byte{self.State})
	d.Write(self.PKr[:])
	d.Write(big.NewInt(int64(self.LimitHeigh)).Bytes())
	if self.Tkn != nil {
		d.Write(self.Tkn.ToHash().NewRef()[:])
	}
	if self.Tkt != nil {
		d.Write(self.Tkt.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Package) Clone() (ret Package) {
	utils.DeepCopy(&ret, self)
	return
}
