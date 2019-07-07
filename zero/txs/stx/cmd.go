package stx

import (
	"math/big"

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
	To    keys.PKr
	Data  []byte
}

func (self *ContractCmd) ToHash() (ret keys.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Asset.ToHash().NewRef()[:])
	d.Write(self.To[:])
	d.Write(self.Data)
	copy(ret[:], d.Sum(nil))
	return
}

type DescCmd struct {
	BuyShare   *BuyShareCmd   `rlp:"nil"`
	RegistPool *RegistPoolCmd `rlp:"nil"`
	ClosePool  *ClosePoolCmd  `rlp:"nil"`
	Contract   *ContractCmd   `rlp:"nil"`
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
