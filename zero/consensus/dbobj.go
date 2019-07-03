package consensus

import (
	"math/big"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
)

type DBObj struct {
	Pre string
}

func makeBlockName(pre string, num uint64, hash *common.Hash) (ret []byte) {
	ret = []byte(pre)
	ret = append(ret, big.NewInt(int64(num)).Bytes()...)
	ret = append(ret, hash[:]...)
	return
}

func (self DBObj) setBlockRecords(batch serodb.Putter, num uint64, hash *common.Hash, records []*Record) {
	if b, err := rlp.EncodeToBytes(&records); err != nil {
		panic(err)
	} else {
		name := makeBlockName(self.Pre, num, hash)
		if err := batch.Put(name, b); err != nil {
			panic(err)
		} else {
			return
		}
	}
}

func (self DBObj) GetBlockRecords(getter serodb.Getter, num uint64, hash *common.Hash) (records []*Record) {
	if b, err := getter.Get(makeBlockName(self.Pre, num, hash)); err != nil {
		return
	} else {
		if err := rlp.DecodeBytes(b, &records); err != nil {
			panic(err)
		} else {
			return
		}
	}
}

func (self DBObj) GetObject(getter serodb.Getter, id []byte, item CItem) (ret CItem) {
	k := key{self.Pre, id}
	if v, err := getter.Get([]byte(k.k())); err != nil {
		return
	} else {
		if e := rlp.DecodeBytes(v, item); e != nil {
			return nil
		}
		return item
	}
}
