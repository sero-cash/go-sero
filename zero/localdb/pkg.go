package localdb

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type ZPkg struct {
	High   uint64
	From   c_type.PKr
	Pack   stx.PkgCreate
	Closed bool
}

func (self *ZPkg) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(big.NewInt(int64(self.High)).Bytes())
	d.Write(self.From[:])
	d.Write(self.Pack.ToHash().NewRef()[:])
	if self.Closed {
		d.Write([]byte{1})
	} else {
		d.Write([]byte{0})
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *ZPkg) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}

type PkgGet struct {
	Out *ZPkg
}

func (self *PkgGet) Unserial(v []byte) (e error) {
	if len(v) < 2 {
		self.Out = nil
		return
	} else {
		self.Out = &ZPkg{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			self.Out = nil
			return
		} else {
			return
		}
	}
}

func PkgKey(root *c_type.Uint256) []byte {
	key := []byte("$SERO_LOCALDB_PKG_HASH$")
	key = append(key, root[:]...)
	return key
}

func PutPkg(db serodb.Putter, hash *c_type.Uint256, pkg *ZPkg) {
	key := PkgKey(hash)
	tri.UpdateDBObj(db, key, pkg)
}

func GetPkg(db serodb.Getter, hash *c_type.Uint256) (ret *ZPkg) {
	key := PkgKey(hash)
	get := PkgGet{}
	tri.GetDBObj(db, key, &get)
	ret = get.Out
	return
}
