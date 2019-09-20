package localdb

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/utils"
)

type OutState struct {
	Index  uint64
	Out_O  *stx_v0.Out_O   `rlp:"nil"`
	Out_Z  *stx_v0.Out_Z   `rlp:"nil"`
	Out_P  *stx_v1.Out_P   `rlp:"nil"`
	Out_C  *stx_v1.Out_C   `rlp:"nil"`
	OutCM  *c_type.Uint256 `rlp:"nil"`
	RootCM *c_type.Uint256 `rlp:"nil"`
}

func (out *OutState) TxType() string {
	if out.Out_O != nil {
		return "Out_O"
	}
	if out.Out_Z != nil {
		return "Out_Z"
	}
	if out.Out_P != nil {
		return "Out_P"
	}
	if out.Out_C != nil {
		return "Out_C"
	}
	return "EMPTY"
}

func (out *OutState) IsZero() bool {
	if out.Out_Z != nil || out.Out_C != nil {
		return true
	} else {
		return false
	}
}
func (out *OutState) IsSzk() bool {
	if out.Out_P != nil || out.Out_C != nil {
		return true
	}
	return false
}

func (self *OutState) Clone() (ret OutState) {
	utils.DeepCopy(&ret, self)
	return
}

func (self *OutState) ToIndexRsk() (ret c_type.Uint256) {
	ret = utils.NewU256(self.Index).ToRef().ToUint256()
	return
}
func (self *OutState) toOutCM() *c_type.Uint256 {
	if self.OutCM == nil {
		if self.Out_O != nil {
			asset := self.Out_O.Asset.ToFlatAsset()
			cm := c_czero.GenOutCM(
				asset.Tkn.Currency.NewRef(),
				asset.Tkn.Value.ToUint256().NewRef(),
				asset.Tkt.Category.NewRef(),
				asset.Tkt.Value.NewRef(),
				&self.Out_O.Memo,
				&self.Out_O.Addr,
				self.ToIndexRsk().NewRef(),
			)
			self.OutCM = &cm
		} else if self.Out_Z != nil {
			self.OutCM = self.Out_Z.OutCM.NewRef()
		}
	}
	return self.OutCM
}

func (self *OutState) ToRootCM() *c_type.Uint256 {
	if self.RootCM == nil {
		if self.Out_O != nil || self.Out_Z != nil {
			out_cm := self.toOutCM()
			cm := c_czero.GenRootCM(self.Index, out_cm)
			self.RootCM = &cm
		} else if self.Out_P != nil {
			rsk_bs := crypto.Keccak256(big.NewInt(int64(self.Index)).Bytes())
			rsk := c_type.Uint256{}
			copy(rsk[:], rsk_bs)
			type_asset := self.Out_P.Asset.ToTypeAsset()
			self.RootCM = c_superzk.GenRootCM_P(
				self.Index,
				&type_asset,
				&self.Out_P.PKr,
				&rsk,
			).NewRef()
		} else if self.Out_C != nil {
			self.RootCM = c_superzk.GenRootCM_C(
				self.Index,
				&self.Out_C.AssetCM,
				&self.Out_C.PKr,
				&self.Out_C.RPK,
			).NewRef()
		}
	}
	return self.RootCM
}

func (self *OutState) ToPKr() *c_type.PKr {
	if self.Out_O != nil {
		return &self.Out_O.Addr
	} else if self.Out_Z != nil {
		return &self.Out_Z.PKr
	} else if self.Out_P != nil {
		return &self.Out_P.PKr
	} else if self.Out_C != nil {
		return &self.Out_C.PKr
	}
	return nil
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
