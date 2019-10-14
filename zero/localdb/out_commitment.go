package localdb

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/utils"
)

func HashIndexRsk(index uint64) (ret c_type.Uint256) {
	ret = utils.NewU256(index).ToRef().ToUint256()
	return
}

func HashIndexAr(index uint64) (ret c_type.Uint256) {
	index_bytes := utils.NewU256(index)
	pre_ar := index_bytes.ToUint256()
	ar_bytes := crypto.Keccak256(pre_ar[:])
	copy(ret[:], ar_bytes)
	ret = c_superzk.ForceFr(&ret)
	return
}

func genOutCM(self *OutState) (cm c_type.Uint256, e error) {
	if self.Out_O != nil {
		cm, e = c_superzk.Czero_genOutCM(
			self.Out_O.Asset.ToTypeAsset().NewRef(),
			&self.Out_O.Memo,
			&self.Out_O.Addr,
			HashIndexRsk(self.Index).NewRef(),
		)
		return
	} else if self.Out_Z != nil {
		cm = self.Out_Z.OutCM
		return
	} else {
		e = errors.New("no output for out cm")
		return
	}
}

func genRootCM(self *OutState) (cm c_type.Uint256, e error) {
	if self.Out_O != nil {
		out_cm := self.OutCM
		cm = c_superzk.Czero_genRootCM(self.Index, out_cm)
		return
	} else if self.Out_Z != nil {
		out_cm := self.Out_Z.OutCM
		cm = c_superzk.Czero_genRootCM(self.Index, &out_cm)
		return
	} else if self.Out_P != nil {
		ar := HashIndexAr(self.Index)
		type_asset := self.Out_P.Asset.ToTypeAsset()
		cm, e = c_superzk.GenRootCM_P(
			self.Index,
			&type_asset,
			&ar,
			&self.Out_P.PKr,
		)
		return
	} else if self.Out_C != nil {
		cm, e = c_superzk.GenRootCM_C(
			self.Index,
			&self.Out_C.AssetCM,
			&self.Out_C.PKr,
		)
		return
	} else {
		e = errors.New("no output for root cm")
		return
	}
}
