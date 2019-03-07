package localdb

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
)

type RootState struct {
	OS     OutState
	TxHash keys.Uint256
	Num    uint64
}

func (self *RootState) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type RootStateGet struct {
	Out *RootState
}

func (self *RootStateGet) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.Out = nil
		return
	} else {
		self.Out = &RootState{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}

func Root2TxHashKey(root *keys.Uint256) []byte {
	key := []byte("$SERO_LOCALDB_ROOTSTATE$")
	key = append(key, root[:]...)
	return key
}

func PutRoot(db serodb.Putter, root *keys.Uint256, rs *RootState) {
	rootkey := Root2TxHashKey(root)
	tri.UpdateDBObj(db, rootkey, rs)
}

func GetRoot(db serodb.Getter, root *keys.Uint256) (ret *RootState) {
	rootkey := Root2TxHashKey(root)
	rootget := RootStateGet{}
	tri.GetDBObj(db, rootkey, &rootget)
	ret = rootget.Out
	return
}
