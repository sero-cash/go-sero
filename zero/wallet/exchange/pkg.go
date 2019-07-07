package exchange

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txtool"
)

func (self *Exchange) findPkgs(pk *keys.Uint512) (pkgs []localdb.ZPkg) {
	prefix := append(pk2pkgKeyPrefix, pk[:]...)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	for iterator.Next() {
		if hash := iterator.Value(); len(hash) == 32 {
			h := keys.Uint256{}
			copy(h[:], hash[:])
			if pkg := localdb.GetPkg(txtool.Ref_inst.Bc.GetDB(), &h); pkg == nil {
				log.Error("the pkg is empty", "pkg", hash)
			} else {
				pkgs = append(pkgs, *pkg)
			}
		} else {
			log.Error("the pkg hash is error", "pkg", hash)
		}
	}
	return
}

type pkg struct {
	id   keys.Uint256
	hash keys.Uint256
	pk   keys.Uint512
	from bool
}

var (
	pk2pkgKeyPrefix = []byte("PK2PKG")
)

type pkgIndexes struct {
}

func (self *Exchange) resolvePkgs(pkgs []localdb.ZPkg) (indexes *pkgIndexes) {
	return
}

func (self *Exchange) indexPkgs(batch serodb.Batch, indexes *pkgIndexes) {
	return
}

func pk2pkgKey(pk keys.Uint512, from *bool, id *keys.Uint256) []byte {
	ret := append(pk2pkgKeyPrefix, pk[:]...)
	if from != nil {
		f := byte(0)
		if *from {
			f = byte(1)
		}
		ret = append(ret, f)
	}
	if id != nil {
		ret = append(ret, id[:]...)
	}
	return ret
}
