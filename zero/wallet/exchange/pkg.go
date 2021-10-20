package exchange

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txtool"
)

var (
	pk_from_id_2_id_KeyPrefix = []byte("PK_FROM_ID_2_ID")
	id_2_pkg_KeyPrefix        = []byte("ID_2_PKG")
)

func pk_from_id_2_id_Key(pk *c_type.Uint512, from *bool, id *c_type.Uint256) []byte {
	ret := append(pk_from_id_2_id_KeyPrefix, pk[:]...)
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

type Pkg struct {
	z    localdb.ZPkg
	to   *c_type.Uint512 `rlp: nil`
	from *c_type.Uint512 `rlp: nil`
}

func id_2_pkg_key(id *c_type.Uint256) []byte {
	ret := append(id_2_pkg_KeyPrefix, id[:]...)
	return ret
}

func (self *Exchange) FindPkgs(pk *c_type.Uint512, from bool) (pkgs []Pkg) {
	prefix := pk_from_id_2_id_Key(pk, &from, nil)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	for iterator.Next() {
		if id := iterator.Value(); len(id) == 32 {
			i := c_type.Uint256{}
			copy(i[:], id[:])
			if pkg := self.FindPkgById(&i); pkg != nil {
				pkgs = append(pkgs, *pkg)
			} else {
				log.Error("find pkg error", "pkg", id)
			}
		} else {
			log.Error("pkg id error", "pkg", id)
		}
	}
	iterator.Release()
	return
}

func (self *Exchange) FindPkgById(id *c_type.Uint256) (pkg *Pkg) {
	if bs, e := self.db.Get(id_2_pkg_key(id)); e != nil {
		return
	} else {
		pkg := Pkg{}
		if e := rlp.DecodeBytes(bs, &pkg); e == nil {
			return &pkg
		} else {
			panic(e)
		}
	}
}

type pkgIndexes struct {
	deleteKeys      [][]byte
	pk_from_id_maps map[string]c_type.Uint256
}

func (self *Exchange) indexPkgs(pks []c_type.Uint512, batch serodb.Batch, blocks []txtool.Block) {
	for _, block := range blocks {
		for _, pkg := range block.Pkgs {
			if p := self.FindPkgById(&pkg.Pack.Id); p != nil {
				if p.to != nil {
					from := false
					batch.Delete(pk_from_id_2_id_Key(p.to, &from, &p.z.Pack.Id))
				}
				if p.from != nil {
					from := true
					batch.Delete(pk_from_id_2_id_Key(p.to, &from, &p.z.Pack.Id))
				}
				batch.Delete(id_2_pkg_key(&p.z.Pack.Id))
			}
			var p Pkg
			if account, ok := self.ownPkr(pks, pkg.Pack.PKr); ok {
				p.to = account.pk
			}
			if account, ok := self.ownPkr(pks, pkg.From); ok {
				p.from = account.pk
			}
			if p.from != nil || p.to != nil {
				if !pkg.Closed {
					p.z = pkg
					if bs, e := rlp.EncodeToBytes(&p); e == nil {
						if p.to != nil {
							from := false
							if e := batch.Put(pk_from_id_2_id_Key(p.to, &from, &p.z.Pack.Id), p.z.Pack.Id[:]); e != nil {
								panic(e)
							}
						}
						if p.from != nil {
							from := true
							if e := batch.Put(pk_from_id_2_id_Key(p.to, &from, &p.z.Pack.Id), p.z.Pack.Id[:]); e != nil {
								panic(e)
							}
						}
						if e := batch.Put(id_2_pkg_key(&p.z.Pack.Id), bs); e != nil {
							panic(e)
						}
					} else {
						panic(e)
					}
				}
			}
		}
	}
	return
}
