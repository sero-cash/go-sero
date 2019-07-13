package app

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/serodb"
	"math/big"
	"sync"
)

type Balance struct {
	Id        uint64
	Hash      keys.Uint256
	Block     uint64
	From      keys.Uint512
	To        keys.PKr
	Currency  keys.Uint256
	Amount    big.Int
	Fee       big.Int
	Timestamp uint64

	mu sync.Mutex

	db *serodb.LDBDatabase
}

func (self *Balance) Store() {

}



func (self *Balance) GetNextId() (id uint64) {
	self.mu.Lock()
	defer self.mu.Unlock()

	return id

}