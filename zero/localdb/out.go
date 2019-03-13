package localdb

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/zstate/tri"
	"github.com/sero-cash/go-sero/zero/utils"
)

type OutState struct {
	Index  uint64
	Out_O  *stx.Out_O    `rlp:"nil"`
	Out_Z  *stx.Out_Z    `rlp:"nil"`
	OutCM  *keys.Uint256 `rlp:"nil"`
	RootCM *keys.Uint256 `rlp:"nil"`
}

func (self *OutState) Clone() (ret OutState) {
	utils.DeepCopy(&ret, self)
	return
}

func (out *OutState) IsO() bool {
	if out.Out_Z == nil {
		return true
	} else {
		return false
	}
}

func (self *OutState) ToIndexRsk() (ret keys.Uint256) {
	ret = utils.NewU256(self.Index).ToRef().ToUint256()
	return
}
func (self *OutState) ToOutCM() *keys.Uint256 {
	if self.IsO() {
		if self.OutCM == nil {
			asset := self.Out_O.Asset.ToFlatAsset()
			cm := cpt.GenOutCM(
				asset.Tkn.Currency.NewRef(),
				asset.Tkn.Value.ToUint256().NewRef(),
				asset.Tkt.Category.NewRef(),
				asset.Tkt.Value.NewRef(),
				&self.Out_O.Memo,
				&self.Out_O.Addr,
				self.ToIndexRsk().NewRef(),
			)
			self.OutCM = &cm
		}
		return self.OutCM
	} else {
		return self.Out_Z.OutCM.NewRef()
	}
}

func (self *OutState) ToRootCM() *keys.Uint256 {
	if self.RootCM == nil {
		out_cm := self.ToOutCM()
		cm := cpt.GenRootCM(self.Index, out_cm)
		self.RootCM = &cm
	}
	return self.RootCM
}

func (self *OutState) ToPKr() *keys.PKr {
	if self.IsO() {
		return &self.Out_O.Addr
	} else {
		return &self.Out_Z.PKr
	}
}

func (self *OutState) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type OutState0Get struct {
	Out *OutState
}

func (self *OutState0Get) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.Out = nil
		return
	} else {
		self.Out = &OutState{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}

func OutKey(root *keys.Uint256) []byte {
	key := []byte("$SERO_LOCALDB_OUT$")
	key = append(key, root[:]...)
	return key
}

func PutOut(db serodb.Putter, root *keys.Uint256, out *OutState) {
	outkey := OutKey(root)
	tri.UpdateDBObj(db, outkey, out)
}

func GetOut(db serodb.Database, root *keys.Uint256) (ret *OutState) {
	outkey := OutKey(root)
	outget := OutState0Get{}
	tri.GetDBObj(db, outkey, &outget)
	ret = outget.Out
	return
}
