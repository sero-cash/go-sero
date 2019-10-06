package stakeservice

import (
	"sync"
	"sync/atomic"

	"github.com/robfig/cron"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/stake"
)

type Account struct {
	accountKey *common.AccountKey
	tk         *c_type.Tk
}

type StakeService struct {
	bc             *core.BlockChain
	accountManager *accounts.Manager
	db             *serodb.LDBDatabase

	accounts sync.Map
	numbers  sync.Map

	feed    event.Feed
	updater event.Subscription        // Wallet update subscriptions for all backends
	update  chan accounts.WalletEvent // Subscription sink for backend wallet changes
	quit    chan chan error
	lock    sync.RWMutex
}

var current_StakeService *StakeService

func CurrentStakeService() *StakeService {
	return current_StakeService
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
	current_StakeService = stakeService

	db, err := serodb.NewLDBDatabase(dbpath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	stakeService.db = db

	stakeService.numbers = sync.Map{}
	stakeService.accounts = sync.Map{}
	for _, w := range accountManager.Wallets() {
		stakeService.initWallet(w)
	}

	AddJob("0/10 * * * * ?", stakeService.stakeIndex)
	go stakeService.updateAccount()
	return stakeService
}

func (self *StakeService) StakePools() (pools []*stake.StakePool) {
	iterator := self.db.NewIteratorWithPrefix(poolPrefix)
	for iterator.Next() {

		value := iterator.Value()
		pool := stake.StakePoolDB.GetObject(self.bc.GetDB(), value, &stake.StakePool{})
		pools = append(pools, pool.(*stake.StakePool))
	}
	return
}

func (self *StakeService) Shares() (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(sharePrefix)
	for iterator.Next() {
		value := iterator.Value()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), value, &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) SharesById(id common.Hash) *stake.Share {
	hash, err := self.db.Get(sharekey(id[:]))
	if err != nil {
		return nil
	}
	return self.getShareByHash(hash)
}

func (self *StakeService) getShareByHash(hash []byte) *stake.Share {
	ret := stake.ShareDB.GetObject(self.bc.GetDB(), hash, &stake.Share{})
	if ret == nil {
		return nil
	}
	return ret.(*stake.Share)
}

func (self *StakeService) SharesByAccountKey(accountKey common.AccountKey) (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(accountKey[:])
	for iterator.Next() {
		value := iterator.Value()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), value, &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) SharesByPkr(pkr c_type.PKr) (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(pkr[:])
	for iterator.Next() {
		value := iterator.Value()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), value, &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) GetBlockRecords(blockNumber uint64) (shares []*stake.Share, pools []*stake.StakePool) {
	header := self.bc.GetHeaderByNumber(blockNumber)
	return stake.GetBlockRecords(self.bc.GetDB(), header.Hash(), blockNumber)
}

func (self *StakeService) stakeIndex() {
	start := uint64(math.MaxUint64)
	self.numbers.Range(func(key, value interface{}) bool {
		num := value.(uint64)
		if start > num {
			start = num
		}
		return true
	})
	if start == uint64(math.MaxUint64) {
		return
	}
	if start < 1300000 {
		start = 1300000
	}

	header := self.bc.CurrentHeader()
	sharesCount := 0
	poolsCount := 0
	batch := self.db.NewBatch()
	blocNumber := start
	for blocNumber+seroparam.DefaultConfirmedBlock() <= header.Number.Uint64() {
		shares, pools := self.GetBlockRecords(blocNumber)
		for _, share := range shares {
			batch.Put(sharekey(share.Id()), share.State())
			batch.Put(pkrShareKey(share.PKr, share.Id()), share.State())
			if accountKey, ok := self.ownPkr(share.PKr); ok {
				batch.Put(pkShareKey(accountKey, share.Id()), share.State())
			}
		}

		for _, pool := range pools {
			batch.Put(poolKey(pool.Id()), pool.State())
		}
		sharesCount += len(shares)
		poolsCount += len(pools)
		blocNumber++
		if blocNumber-start >= 10000 {
			break
		}
	}
	if blocNumber == start {
		return
	}

	self.numbers.Range(func(key, value interface{}) bool {
		accountKey := key.(common.AccountKey)
		batch.Put(numKey(accountKey), utils.EncodeNumber(blocNumber))
		return true
	})
	err := batch.Write()
	if err == nil {
		self.numbers.Range(func(key, value interface{}) bool {
			pk := key.(common.AccountKey)
			self.numbers.Store(pk, blocNumber)
			return true
		})
		log.Info("StakeIndex", "blockNumber", blocNumber, "sharesCount", sharesCount, "poolsCount", poolsCount)
	}
}

func (self *StakeService) ownPkr(pkr c_type.PKr) (pk *common.AccountKey, ok bool) {
	var account *Account
	self.accounts.Range(func(key, value interface{}) bool {
		a := value.(*Account)
		if superzk.IsMyPKr(a.tk, &pkr) {
			account = a
			return false
		}
		return true
	})
	if account != nil {
		return account.accountKey, true
	}
	return
}

func (self *StakeService) updateAccount() {
	// Close all subscriptions when the manager terminates
	defer func() {
		self.lock.Lock()
		self.updater.Unsubscribe()
		self.updater = nil
		self.lock.Unlock()
	}()

	// Loop until termination
	for {
		select {
		case event := <-self.update:
			// Wallet event arrived, update local cache
			self.lock.Lock()
			switch event.Kind {
			case accounts.WalletArrived:
				self.initWallet(event.Wallet)
			case accounts.WalletDropped:
				accountKey := event.Wallet.Accounts()[0].Key
				self.numbers.Delete(accountKey)
			}
			self.lock.Unlock()

		case errc := <-self.quit:
			// Manager terminating, return
			errc <- nil
			return
		}
	}
}

func (self *StakeService) initWallet(w accounts.Wallet) {
	if _, ok := self.accounts.Load(w.Accounts()[0].Key); !ok {
		account := Account{}
		account.accountKey = &w.Accounts()[0].Key
		account.tk = w.Accounts()[0].Tk.ToTk().NewRef()
		self.accounts.Store(*account.accountKey, &account)

		var num uint64
		if num = self.starNum(account.accountKey); num < w.Accounts()[0].At {
			num = w.Accounts()[0].At
		}
		self.numbers.Store(*account.accountKey, num)
		log.Info("Add PK", "accountKey", w.Accounts()[0].Key, "At", num)
	}
}

func (self *StakeService) starNum(accountKey *common.AccountKey) uint64 {
	value, err := self.db.Get(numKey(*accountKey))
	if err != nil {
		return 0
	}
	return utils.DecodeNumber(value)
}

var (
	numPrefix   = []byte("NUM")
	sharePrefix = []byte("SHARE")
	poolPrefix  = []byte("POOL")
)

func pkShareKey(accountKey *common.AccountKey, key []byte) []byte {
	return append(accountKey[:], key[:]...)
}

func pkrShareKey(pk c_type.PKr, key []byte) []byte {
	return append(pk[:], key[:]...)
}

func sharekey(key []byte) []byte {
	return append(sharePrefix, key[:]...)
}

func poolKey(key []byte) []byte {
	return append(poolPrefix, key[:]...)
}

func numKey(accountKey common.AccountKey) []byte {
	return append(numPrefix, accountKey[:]...)
}

func AddJob(spec string, run RunFunc) *cron.Cron {
	c := cron.New()
	c.AddJob(spec, &RunJob{run: run})
	c.Start()
	return c
}

type (
	RunFunc func()
)

type RunJob struct {
	runing int32
	run    RunFunc
}

func (r *RunJob) Run() {
	x := atomic.LoadInt32(&r.runing)
	if x == 1 {
		return
	}

	atomic.StoreInt32(&r.runing, 1)
	defer func() {
		atomic.StoreInt32(&r.runing, 0)
	}()

	r.run()
}
