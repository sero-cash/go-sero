package stx

import (
	"math/big"
	"sync/atomic"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type BuyShareCmd struct {
	Value utils.U256
	Vote  keys.PKr
	Pool  *keys.Uint256 `rlp:"nil"`
}

func (self *BuyShareCmd) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Value.ToInt().Bytes())
	d.Write(self.Vote[:])
	if self.Pool != nil {
		d.Write(self.Pool[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *BuyShareCmd) Asset() (ret assets.Asset) {
	ret.Tkn = &assets.Token{
		utils.CurrencyToUint256("SERO"),
		self.Value,
	}
	return
}

type RegistPoolCmd struct {
	Value   utils.U256
	Vote    keys.PKr
	FeeRate uint32
}

func (self *RegistPoolCmd) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Value.ToInt().Bytes())
	d.Write(self.Vote[:])
	d.Write(big.NewInt(int64(self.FeeRate)).Bytes())
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *RegistPoolCmd) Asset() (ret assets.Asset) {
	ret.Tkn = &assets.Token{
		utils.CurrencyToUint256("SERO"),
		self.Value,
	}
	return
}

type ClosePoolCmd struct{}

func (self *ClosePoolCmd) ToHash() (ret keys.Uint256) {
	ret[0] = 1
	return
}

type ContractCmd struct {
	Asset assets.Asset
	To    *keys.PKr
	Data  []byte
}

func (self *ContractCmd) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Asset.ToHash().NewRef()[:])
	if self.To != nil {
		d.Write(self.To[:])
	}
	d.Write(self.Data)
	copy(ret[:], d.Sum(nil))
	return
}

type DescCmd struct {
	BuyShare   *BuyShareCmd   `rlp:"nil"`
	RegistPool *RegistPoolCmd `rlp:"nil"`
	ClosePool  *ClosePoolCmd  `rlp:"nil"`
	Contract   *ContractCmd   `rlp:"nil"`

	//Cache
	assetCC atomic.Value
}

func (self *DescCmd) ToPkr() *keys.PKr {
	if self.BuyShare != nil {
		return &self.BuyShare.Vote
	}
	if self.RegistPool != nil {
		return &self.RegistPool.Vote
	}
	return nil
}

func (self *DescCmd) ToAssetCC() *keys.Uint256 {
	if asset := self.OutAsset(); asset != nil {
		if cc, ok := self.assetCC.Load().(keys.Uint256); ok {
			return &cc
		}
		asset := asset.ToFlatAsset()
		asset_desc := cpt.AssetDesc{
			Tkn_currency: asset.Tkn.Currency,
			Tkn_value:    asset.Tkn.Value.ToUint256(),
			Tkt_category: asset.Tkt.Category,
			Tkt_value:    asset.Tkt.Value,
		}
		cpt.GenAssetCC(&asset_desc)
		v := asset_desc.Asset_cc
		self.assetCC.Store(v)
		return &v
	} else {
		return nil
	}
}

func (self *DescCmd) OutAsset() *assets.Asset {
	if self.BuyShare != nil {
		asset := self.BuyShare.Asset()
		return &asset
	}
	if self.RegistPool != nil {
		asset := self.RegistPool.Asset()
		return &asset
	}
	if self.Contract != nil {
		return &self.Contract.Asset
	}
	return nil
}

func (self *DescCmd) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	if self.BuyShare != nil {
		d.Write(self.BuyShare.ToHash().NewRef()[:])
	}
	if self.RegistPool != nil {
		d.Write(self.RegistPool.ToHash().NewRef()[:])
	}
	if self.ClosePool != nil {
		d.Write(self.ClosePool.ToHash().NewRef()[:])
	}
	if self.Contract != nil {
		d.Write(self.Contract.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *DescCmd) Count() int {
	count := 0
	if self.BuyShare != nil {
		count++
	}
	if self.RegistPool != nil {
		count++
	}
	if self.ClosePool != nil {
		count++
	}
	if self.Contract != nil {
		count++
	}
	return count
}

func (self *DescCmd) Valid() bool {
	if self.Count() <= 1 {
		return true
	} else {
		return false
	}
}
