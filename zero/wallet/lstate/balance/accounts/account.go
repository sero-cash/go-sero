package accounts

import (
	"bytes"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-czero-import/keys"
)

type Account struct {
	Tk      keys.Uint512
	NextNum uint64
	Token   map[keys.Uint256]*utils.U256
	Ticket  map[keys.Uint256][]keys.Uint256
	dirty   bool
}

func NewAccount(tk *keys.Uint512, next_num uint64) (ret Account) {
	ret.Tk = *tk
	ret.NextNum = next_num
	ret.Token = make(map[keys.Uint256]*utils.U256)
	ret.Ticket = make(map[keys.Uint256][]keys.Uint256)
	ret.dirty = true
	return
}

func (self *Account) Next() {
	self.NextNum++
	self.dirty = true
}

func (self *Account) AddTicket(a *assets.Ticket) {
	if a.Category == keys.Empty_Uint256 {
		return
	}
	if a.Value == keys.Empty_Uint256 {
		return
	}
	if v, ok := self.Ticket[a.Category]; ok {
		self.Ticket[a.Category] = append(v, a.Value)
		self.dirty = true
	} else {
		self.Ticket[a.Category] = append([]keys.Uint256{}, a.Value)
		self.dirty = true
	}
}

func (self *Account) DelTicket(a *assets.Ticket) {
	if a.Category == keys.Empty_Uint256 {
		return
	}
	if a.Value == keys.Empty_Uint256 {
		return
	}
	if v, ok := self.Ticket[a.Category]; ok {
		for k, t := range v {
			if t == a.Value {
				self.Ticket[a.Category] = append(v[:k], v[k+1:]...)
				self.dirty = true
				return
			}
		}
		panic(errors.New("Del Token but balance is nil"))
	} else {
		panic(errors.New("Del Token but balance is nil"))
	}
}

func (self *Account) AddToken(a *assets.Token) {
	if a.Currency == keys.Empty_Uint256 {
		return
	}
	if a.Value.Cmp(&utils.U256_0) == 0 {
		return
	}
	if v, ok := self.Token[a.Currency]; ok {
		v.AddU(&a.Value)
		self.dirty = true
	} else {
		self.Token[a.Currency] = a.Value.ToRef()
		self.dirty = true
	}
}

func (self *Account) DelToken(a *assets.Token) {
	if a.Currency == keys.Empty_Uint256 {
		return
	}
	if a.Value.Cmp(&utils.U256_0) == 0 {
		return
	}
	if v, ok := self.Token[a.Currency]; ok {
		v.SubU(&a.Value)
		self.dirty = true
	} else {
		panic(errors.New("Del Token but balance is nil"))
	}
}

const TK_LASTNUM_KEY = "TK$LASTNUM$KEY-"

func (self *DB) GetAccount(k *keys.Uint512) (ret Account) {
	if v, err := self.db.Get(Bytes2Key(TK_LASTNUM_KEY, k[:]), nil); err != nil {
		return NewAccount(k, 0)
	} else {
		utils.DeepUnserial(bytes.NewBuffer(v), &ret)
		return
	}
}

func (self *DB) SetAccount(batch *leveldb.Batch, a *Account) {
	if a.dirty {
		v := utils.DeepSerial(a)
		a.dirty = false
		batch.Put(Bytes2Key(TK_LASTNUM_KEY, a.Tk[:]), v.Bytes())
	}
}

func (self *DB) AddAccount(tk *keys.Uint512, next_num uint64) (ret bool) {
	if _, err := self.db.Get(Bytes2Key(TK_LASTNUM_KEY, tk[:]), nil); err != nil {
		new_account := NewAccount(tk, next_num)
		v := utils.DeepSerial(&new_account)
		if err := self.db.Put(Bytes2Key(TK_LASTNUM_KEY, tk[:]), v.Bytes(), nil); err != nil {
			panic(err)
		} else {
			return true
		}
	} else {
		return false
	}
}
