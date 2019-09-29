package stx

import (
	"math/big"
	"sync/atomic"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/crypto/sha3"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type BuyShareCmd struct {
	Value utils.U256
	Vote  c_type.PKr
	Pool  *c_type.Uint256 `rlp:"nil"`
}

func (self *BuyShareCmd) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Value.ToBEBytes())
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
	Vote    c_type.PKr
	FeeRate uint32
}

func (self *RegistPoolCmd) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Value.ToBEBytes())
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

type ClosePoolCmd struct {
	None byte
}

func (self *ClosePoolCmd) ToHash() (ret c_type.Uint256) {
	ret[0] = 1
	return
}

type ContractCmd struct {
	Asset assets.Asset
	To    *c_type.PKr `rlp:"nil"`
	Data  []byte
}

func (self *ContractCmd) ToHash() (ret c_type.Uint256) {
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
	assetCC_Czero atomic.Value
	assetCC_Szk   atomic.Value
}

func (self *DescCmd) ToPkr() *c_type.PKr {
	if self.BuyShare != nil {
		return &self.BuyShare.Vote
	}
	if self.RegistPool != nil {
		return &self.RegistPool.Vote
	}
	return nil
}

func (self *DescCmd) ToAssetCC_Czero() *c_type.Uint256 {
	if asset := self.OutAsset(); asset != nil {
		if cc, ok := self.assetCC_Czero.Load().(c_type.Uint256); ok {
			return &cc
		}
		asset_desc := c_czero.AssetDesc{
			Asset: asset.ToTypeAsset(),
		}
		c_czero.GenAssetCC(&asset_desc)
		v := asset_desc.Asset_cc
		self.assetCC_Czero.Store(v)
		return &v
	} else {
		return nil
	}
}

func (self *DescCmd) ToAssetCC_Szk() *c_type.Uint256 {
	if asset := self.OutAsset(); asset != nil {
		if cc, ok := self.assetCC_Szk.Load().(c_type.Uint256); ok {
			return &cc
		}
		v, _ := c_superzk.GenAssetCC(asset.ToTypeAsset().NewRef())
		self.assetCC_Szk.Store(v)
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

func (self *DescCmd) ToHash() (ret c_type.Uint256) {
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
