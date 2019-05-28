package accounts

import (
	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/syndtr/goleveldb/leveldb"
)

type DB struct {
	db *leveldb.DB
}

func NewDB(dir string) (ret DB) {
	if db, err := leveldb.OpenFile(dir, nil); err != nil {
		panic(err)
	} else {
		ret.db = db
	}
	return
}

func (self *DB) DB() (ret *leveldb.DB) {
	return self.db
}

func Bytes2Key(prefix string, bytes []byte) (ret []byte) {
	ret = append([]byte(prefix), bytes...)
	return
}

func (self *DB) AddPkg(a *Account, pg *localdb.ZPkg) (ret bool) {
	return false
}
