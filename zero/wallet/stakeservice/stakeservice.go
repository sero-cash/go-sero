package stakeservice

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/stake"
	"github.com/sero-cash/go-sero/zero/txtool/prepare"
	"github.com/sero-cash/go-sero/zero/wallet/utils"
	"sync"
)

type Account struct {
	pk *keys.Uint512
	tk *keys.Uint512
}

type StakeService struct {
	bc             *core.BlockChain
	accountManager *accounts.Manager
	db             *serodb.LDBDatabase

	nextBlockNumber uint64

	accounts sync.Map

	feed    event.Feed
	updater event.Subscription        // Wallet update subscriptions for all backends
	update  chan accounts.WalletEvent // Subscription sink for backend wallet changes
	quit    chan chan error
	lock    sync.RWMutex
}

func NewStakeService(dbpath string, bc *core.BlockChain, accountManager *accounts.Manager) *StakeService {
	update := make(chan accounts.WalletEvent, 1)
	updater := accountManager.Subscribe(update)

	stakeService := &StakeService{
		bc:             bc,
		accountManager: accountManager,
		update:         update,
		updater:        updater,
	}

	db, err := serodb.NewLDBDatabase(dbpath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	stakeService.db = db

	stakeService.accounts = sync.Map{}
	for _, w := range accountManager.Wallets() {
		stakeService.initWallet(w)
	}

	value, err := db.Get(nextKey)
	if err != nil {
		stakeService.nextBlockNumber = uint64(1)
	} else {
		stakeService.nextBlockNumber = prepare.DecodeNumber(value)
	}

	utils.AddJob("0/10 * * * * ?", stakeService.stakeIndex)
	return stakeService
}

func (self *StakeService) StakePools() (pools []*stake.StakePool) {
	iterator := self.db.NewIteratorWithPrefix(poolPrefix)
	if iterator.Next() {
		key := iterator.Key()
		pool := stake.StakePoolDB.GetObject(self.bc.GetDB(), key[4:], &stake.StakePool{})
		pools = append(pools, pool.(*stake.StakePool))
	}
	return
}

func (self *StakeService) Shares() (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(sharePrefix)
	if iterator.Next() {
		key := iterator.Key()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), key[5:], &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) SharesByPk(pk keys.Uint512) (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(pk[:])
	if iterator.Next() {
		key := iterator.Key()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), key[64:], &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) GetBlockRecords(blockNumber uint64) (shares []*stake.Share, pools []*stake.StakePool) {
	header := self.bc.GetHeaderByNumber(blockNumber)
	return stake.GetBlockRecords(self.bc.GetDB(), header.Hash(), blockNumber)
}

func (self *StakeService) stakeIndex() {
	header := self.bc.CurrentHeader()
	batch := self.db.NewBatch()
	blockNumber := self.nextBlockNumber

	sharesCount := 0
	poolsCount := 0
	for blockNumber+seroparam.DefaultConfirmedBlock() <= header.Number.Uint64() {
		shares, pools := self.GetBlockRecords(blockNumber)
		for _, share := range shares {
			batch.Put(sharekey(share.Id()), share.State())
			if pk, ok := self.ownPkr(share.PKr); ok {
				batch.Put(pkShareKey(pk, share.Id()), share.State())
			}
		}

		for _, pool := range pools {
			batch.Put(poolKey(pool.Id()), pool.State())
		}
		sharesCount += len(shares)
		poolsCount += len(pools)
		blockNumber += 1

	}
	batch.Put(nextKey, prepare.EncodeNumber(blockNumber+1))
	err := batch.Write()
	if err == nil {
		self.nextBlockNumber = blockNumber + 1
		log.Info("StakeIndex", "blockNumber", blockNumber, "sharesCount", sharesCount, "poolsCount", poolsCount)
	}
}

func (self *StakeService) ownPkr(pkr keys.PKr) (pk *keys.Uint512, ok bool) {
	var account *Account
	self.accounts.Range(func(key, value interface{}) bool {
		a := value.(*Account)
		if keys.IsMyPKr(a.tk, &pkr) {
			account = a
			return false
		}
		return true
	})
	if account != nil {
		return account.pk, true
	}
	return
}

var (
	sharePrefix = []byte("SHARE")
	poolPrefix  = []byte("POOL")
	nextKey     = []byte("NEXT")
)

func pkShareKey(pk *keys.Uint512, key []byte) []byte {
	return append(pk[:], key[:]...)
}

func sharekey(key []byte) []byte {
	return append(sharePrefix, key[:]...)
}

func poolKey(key []byte) []byte {
	return append(poolPrefix, key[:]...)
}

func (self *StakeService) initWallet(w accounts.Wallet) {

	if _, ok := self.accounts.Load(*w.Accounts()[0].Address.ToUint512()); !ok {
		account := Account{}
		account.pk = w.Accounts()[0].Address.ToUint512()
		account.tk = w.Accounts()[0].Tk.ToUint512()
		self.accounts.Store(*account.pk, &account)
	}
}

//
//func (self *StakeService) updateAccount() {
//	// Close all subscriptions when the manager terminates
//	defer func() {
//		self.lock.Lock()
//		self.updater.Unsubscribe()
//		self.updater = nil
//		self.lock.Unlock()
//	}()
//
//	// Loop until termination
//	for {
//		select {
//		case event := <-self.update:
//			// Wallet event arrived, update local cache
//			self.lock.Lock()
//			switch event.Kind {
//			case accounts.WalletArrived:
//				//wallet := event.Wallet
//				self.initWallet(event.Wallet)
//			case accounts.WalletDropped:
//			}
//			self.lock.Unlock()
//
//		case errc := <-self.quit:
//			// Manager terminating, return
//			errc <- nil
//			return
//		}
//	}
//}
