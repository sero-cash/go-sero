package exchange

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/robfig/cron"
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Account struct {
	wallet        accounts.Wallet
	pk            *keys.Uint512
	tk            *keys.Uint512
	skr           keys.PKr
	mainPkr       keys.PKr
	balances      map[string]*big.Int
	utxoNums      map[string]uint64
	isChanged     bool
	nextMergeTime time.Time
}

type PkrAccount struct {
	Pkr      keys.PKr
	balances map[string]*big.Int
	num      uint64
}

type Utxo struct {
	Pkr    keys.PKr
	Root   keys.Uint256
	TxHash keys.Uint256
	Nil    keys.Uint256
	Num    uint64
	Asset  assets.Asset
	IsZ    bool
	flag   int
}

type UtxoList []Utxo

func (list UtxoList) Len() int {
	return len(list)
}

func (list UtxoList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list UtxoList) Less(i, j int) bool {
	if list[i].flag == list[j].flag {
		return list[i].Asset.Tkn.Value.ToIntRef().Cmp(list[j].Asset.Tkn.Value.ToIntRef()) < 0
	} else {
		return list[i].flag < list[j].flag
	}
}

func (list UtxoList) Roots() (roots prepare.Utxos) {
	for _, utxo := range list {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	return
}

type (
	HandleUtxoFunc func(utxo Utxo)
)

type PkKey struct {
	PK  keys.Uint512
	Num uint64
}

type PkrKey struct {
	pkr keys.PKr
	num uint64
}

type FetchJob struct {
	start    uint64
	accounts []Account
}

type Exchange struct {
	db             *serodb.LDBDatabase
	txPool         *core.TxPool
	accountManager *accounts.Manager

	accounts    sync.Map
	pkrAccounts sync.Map

	usedFlag sync.Map
	numbers  sync.Map

	feed    event.Feed
	updater event.Subscription        // Wallet update subscriptions for all backends
	update  chan accounts.WalletEvent // Subscription sink for backend wallet changes
	quit    chan chan error
	lock    sync.RWMutex
}

var current_exchange *Exchange

func CurrentExchange() *Exchange {
	return current_exchange
}

func NewExchange(dbpath string, txPool *core.TxPool, accountManager *accounts.Manager, autoMerge bool) (exchange *Exchange) {

	update := make(chan accounts.WalletEvent, 1)
	updater := accountManager.Subscribe(update)

	exchange = &Exchange{
		txPool:         txPool,
		accountManager: accountManager,
		update:         update,
		updater:        updater,
	}
	current_exchange = exchange

	db, err := serodb.NewLDBDatabase(dbpath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	exchange.db = db

	exchange.numbers = sync.Map{}
	exchange.accounts = sync.Map{}
	for _, w := range accountManager.Wallets() {
		exchange.initWallet(w)
	}

	exchange.pkrAccounts = sync.Map{}
	exchange.usedFlag = sync.Map{}

	AddJob("0/10 * * * * ?", exchange.fetchBlockInfo)

	if autoMerge {
		AddJob("0 0/5 * * * ?", exchange.merge)
	}

	go exchange.updateAccount()
	log.Info("Init NewExchange success")
	return
}

func (self *Exchange) initWallet(w accounts.Wallet) {

	if _, ok := self.accounts.Load(*w.Accounts()[0].Address.ToUint512()); !ok {
		account := Account{}
		account.wallet = w
		account.pk = w.Accounts()[0].Address.ToUint512()
		account.tk = w.Accounts()[0].Tk.ToUint512()
		copy(account.skr[:], account.tk[:])
		account.mainPkr = prepare.CreatePkr(account.pk, 1)
		account.isChanged = true
		account.nextMergeTime = time.Now()
		self.accounts.Store(*account.pk, &account)

		if num := self.starNum(account.pk); num > w.Accounts()[0].At {
			self.numbers.Store(*account.pk, num)
		} else {
			self.numbers.Store(*account.pk, w.Accounts()[0].At)
		}

		log.Info("Add PK", "address", w.Accounts()[0].Address, "At", self.GetCurrencyNumber(*account.pk))
	}
}

func (self *Exchange) starNum(pk *keys.Uint512) uint64 {
	value, err := self.db.Get(numKey(*pk))
	if err != nil {
		return 0
	}
	return utils.DecodeNumber(value)
}

func (self *Exchange) updateAccount() {
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
				//wallet := event.Wallet
				self.initWallet(event.Wallet)
			case accounts.WalletDropped:
				pk := *event.Wallet.Accounts()[0].Address.ToUint512()
				self.numbers.Delete(pk)
			}
			self.lock.Unlock()

		case errc := <-self.quit:
			// Manager terminating, return
			errc <- nil
			return
		}
	}
}

func (self *Exchange) GetUtxoNum(pk keys.Uint512) map[string]uint64 {
	if account := self.getAccountByPk(pk); account != nil {
		return account.utxoNums
	}
	return map[string]uint64{}
}

func (self *Exchange) GetRootByNil(Nil keys.Uint256) (root *keys.Uint256) {
	data, err := self.db.Get(nilToRootKey(Nil))
	if err != nil {
		return
	}
	root = &keys.Uint256{}
	copy(root[:], data[:])
	return
}

func (self *Exchange) GetCurrencyNumber(pk keys.Uint512) uint64 {
	value, ok := self.numbers.Load(pk)
	if !ok {
		return 0
	}
	if value.(uint64) == 0 {
		return value.(uint64)
	}
	return value.(uint64) - 1
}

func (self *Exchange) GetPkr(pk *keys.Uint512, index *keys.Uint256) (pkr keys.PKr, err error) {
	if index == nil {
		return pkr, errors.New("index must not be empty")
	}
	if new(big.Int).SetBytes(index[:]).Cmp(big.NewInt(100)) < 0 {
		return pkr, errors.New("index must > 100")
	}
	if _, ok := self.accounts.Load(*pk); !ok {
		return pkr, errors.New("not found Pk")
	}

	return keys.Addr2PKr(pk, index), nil
}

func (self *Exchange) ClearUsedFlagForPK(pk *keys.Uint512) (count int) {
	if _, ok := self.accounts.Load(*pk); ok {
		prefix := append(pkPrefix, pk[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)

		for iterator.Next() {
			key := iterator.Key()
			var root keys.Uint256
			copy(root[:], key[98:130])
			if _, flag := self.usedFlag.Load(root); flag {
				self.usedFlag.Delete(root)
				count++
			}
		}
	}
	return
}

func (self *Exchange) ClearUsedFlagForRoot(root keys.Uint256) (count int) {
	if _, flag := self.usedFlag.Load(root); flag {
		self.usedFlag.Delete(root)
		count++
	}
	return
}

func (self *Exchange) GetLockedBalances(pk keys.Uint512) (balances map[string]*big.Int) {
	if _, ok := self.accounts.Load(pk); ok {
		prefix := append(pkPrefix, pk[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)
		balances = map[string]*big.Int{}

		for iterator.Next() {
			key := iterator.Key()
			var root keys.Uint256
			copy(root[:], key[98:130])
			if utxo, err := self.getUtxo(root); err == nil {
				if utxo.Asset.Tkn != nil {
					if _, flag := self.usedFlag.Load(utxo.Root); flag {
						currency := common.BytesToString(utxo.Asset.Tkn.Currency[:])
						if amount, ok := balances[currency]; ok {
							amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
						} else {
							balances[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
						}
					}
				}
			}
		}
		return balances
	}
	return
}

func (self *Exchange) GetMaxAvailable(pk keys.Uint512, currency string) (amount *big.Int) {
	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	amount = new(big.Int)
	count := 0
	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if _, flag := self.usedFlag.Load(utxo.Root); !flag {
				if utxo.Asset.Tkn != nil {
					if utxo.IsZ {
						amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
					} else {
						if count < 2500 {
							amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
							count++
						}
					}
				}
			}
		}
	}
	return
}

func (self *Exchange) GetBalances(pk keys.Uint512) (balances map[string]*big.Int) {
	if value, ok := self.accounts.Load(pk); ok {
		account := value.(*Account)
		if account.isChanged {
			prefix := append(pkPrefix, pk[:]...)
			iterator := self.db.NewIteratorWithPrefix(prefix)
			balances = map[string]*big.Int{}
			utxoNums := map[string]uint64{}
			for iterator.Next() {
				key := iterator.Key()
				var root keys.Uint256
				copy(root[:], key[98:130])
				if utxo, err := self.getUtxo(root); err == nil {
					if utxo.Asset.Tkn != nil {
						currency := common.BytesToString(utxo.Asset.Tkn.Currency[:])
						if amount, ok := balances[currency]; ok {
							amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
							utxoNums[currency] += 1
						} else {
							balances[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
							utxoNums[currency] = 1
						}
					}
				}
			}
			account.balances = balances
			account.utxoNums = utxoNums
			account.isChanged = false
		} else {
			return account.balances
		}
	}

	return
}

type BlockInfo struct {
	Num  uint64
	Hash keys.Uint256
	Ins  []keys.Uint256
	Outs []Utxo
}

func (self *Exchange) GetBlocksInfo(start, end uint64) (blocks []BlockInfo, err error) {
	iterator := self.db.NewIteratorWithPrefix(blockPrefix)
	for ok := iterator.Seek(blockKey(start)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := utils.DecodeNumber(key[5:13])
		if num >= end {
			break
		}

		var block BlockInfo
		if err = rlp.Decode(bytes.NewReader(iterator.Value()), &block); err != nil {
			log.Error("Exchange Invalid block RLP", "Num", num, "err", err)
			return
		}
		blocks = append(blocks, block)
	}
	return
}

func (self *Exchange) GetRecordsByTxHash(txHash keys.Uint256) (records []Utxo, err error) {
	data, err := self.db.Get(txKey(txHash))
	if err != nil {
		return
	}
	if err = rlp.Decode(bytes.NewReader(data), &records); err != nil {
		log.Error("Invalid utxos RLP", "txHash", common.Bytes2Hex(txHash[:]), "err", err)
		return
	}
	return
}

func (self *Exchange) GetRecordsByPk(PK *keys.Uint512, begin, end uint64) (records []Utxo, err error) {
	err = self.iteratorUtxo(PK, begin, end, func(utxo Utxo) {
		records = append(records, utxo)
	})
	return
}

func (self *Exchange) GetRecordsByPkr(pkr keys.PKr, begin, end uint64) (records []Utxo, err error) {
	account := self.getAccountByPkr(pkr)
	if account == nil {
		err = errors.New("not found PK by pkr")
		return
	}
	PK := account.pk

	err = self.iteratorUtxo(PK, begin, end, func(utxo Utxo) {
		if pkr != utxo.Pkr {
			return
		}
		records = append(records, utxo)
	})
	return
}

func (self *Exchange) GenTxWithSign(param prepare.PreTxParam) (pretx *txtool.GTxParam, tx *txtool.GTx, e error) {
	if self == nil {
		e = errors.New("exchange instance is nil")
		return
	}
	var roots prepare.Utxos
	if roots, e = prepare.SelectUtxos(&param, self); e != nil {
		return
	}

	var account *Account
	if value, ok := self.accounts.Load(param.From); ok {
		account = value.(*Account)
	} else {
		e = errors.New("not found Pk")
		return
	}

	if pretx, tx, e = self.genTx(roots, account, param.RefundTo, param.Receptions, &param.Cmds, &param.Fee, param.GasPrice); e != nil {
		log.Error("Exchange genTx", "error", e)
		return
	}
	tx.Hash = tx.Tx.ToHash()
	log.Info("Exchange genTx success")
	return
}

func (self *Exchange) getAccountByPk(pk keys.Uint512) *Account {
	if value, ok := self.accounts.Load(pk); ok {
		return value.(*Account)
	}
	return nil
}

func (self *Exchange) getAccountByPkr(pkr keys.PKr) (a *Account) {
	self.accounts.Range(func(pk, value interface{}) bool {
		account := value.(*Account)
		if keys.IsMyPKr(account.tk, &pkr) {
			a = account
			return false
		}
		return true
	})
	return
}

func (self *Exchange) ClearTxParam(txParam *txtool.GTxParam) (count int) {
	if self == nil {
		return
	}
	if txParam == nil {
		return
	}
	for _, in := range txParam.Ins {
		count += self.ClearUsedFlagForRoot(in.Out.Root)
	}
	return
}

func (self *Exchange) genTx(utxos prepare.Utxos, account *Account, refundTo *keys.PKr, receptions []prepare.Reception, cmds *prepare.Cmds, fee *assets.Token, gasPrice *big.Int) (txParam *txtool.GTxParam, tx *txtool.GTx, e error) {
	if refundTo == nil {
		refundTo = &account.mainPkr
	}

	if txParam, e = self.buildTxParam(utxos, refundTo, receptions, cmds, fee, gasPrice); e != nil {
		return
	}

	var seed *address.Seed
	if seed, e = account.wallet.GetSeed(); e != nil {
		self.ClearTxParam(txParam)
		return
	}

	sk := keys.Seed2Sk(seed.SeedToUint256())
	gtx, err := flight.SignTx(&sk, txParam)
	if err != nil {
		self.ClearTxParam(txParam)
		e = err
		return
	} else {
		tx = &gtx
		return
	}
}

func (self *Exchange) commitTx(tx *txtool.GTx) (err error) {
	gasPrice := big.Int(tx.GasPrice)
	gas := uint64(tx.Gas)
	signedTx := types.NewTxWithGTx(gas, &gasPrice, &tx.Tx)
	log.Info("Exchange commitTx", "txhash", signedTx.Hash().String())
	err = self.txPool.AddLocal(signedTx)
	return err
}

func (self *Exchange) iteratorUtxo(PK *keys.Uint512, begin, end uint64, handler HandleUtxoFunc) (e error) {
	var pk keys.Uint512
	if PK != nil {
		pk = *PK
	}
	iterator := self.db.NewIteratorWithPrefix(utxoPrefix)
	for ok := iterator.Seek(utxoKey(begin, pk)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := utils.DecodeNumber(key[4:12])
		if num >= end {
			break
		}
		copy(pk[:], key[12:76])

		if PK != nil && *PK != pk {
			continue
		}

		value := iterator.Value()
		roots := []keys.Uint256{}
		if err := rlp.Decode(bytes.NewReader(value), &roots); err != nil {
			log.Error("Invalid roots RLP", "PK", common.Bytes2Hex(pk[:]), "blockNumber", num, "err", err)
			e = err
			return
		}
		for _, root := range roots {
			if utxo, err := self.getUtxo(root); err != nil {
				return
			} else {
				handler(utxo)
			}
		}
	}

	return
}

func (self *Exchange) getUtxo(root keys.Uint256) (utxo Utxo, e error) {
	data, err := self.db.Get(rootKey(root))
	if err != nil {
		return
	}
	if err := rlp.Decode(bytes.NewReader(data), &utxo); err != nil {
		log.Error("Exchange Invalid utxo RLP", "root", common.Bytes2Hex(root[:]), "err", err)
		e = err
		return
	}

	if value, ok := self.usedFlag.Load(utxo.Root); ok {
		utxo.flag = value.(int)
	}
	return
}

func (self *Exchange) findUtxosByTicket(pk *keys.Uint512, tickets map[keys.Uint256]keys.Uint256) (utxos []Utxo, remain map[keys.Uint256]keys.Uint256) {
	remain = map[keys.Uint256]keys.Uint256{}
	for value, category := range tickets {
		remain[value] = category
		prefix := append(pkPrefix, append(pk[:], value[:]...)...)
		iterator := self.db.NewIteratorWithPrefix(prefix)
		if iterator.Next() {
			key := iterator.Key()
			var root keys.Uint256
			copy(root[:], key[98:130])

			if utxo, err := self.getUtxo(root); err == nil {
				if utxo.Asset.Tkt != nil && utxo.Asset.Tkt.Category == category {
					if _, ok := self.usedFlag.Load(utxo.Root); !ok {
						utxos = append(utxos, utxo)
						delete(remain, value)
					}
				}
			}
		}
	}
	return
}

func (self *Exchange) findUtxos(pk *keys.Uint512, currency string, amount *big.Int) (utxos []Utxo, remain *big.Int) {
	remain = new(big.Int).Set(amount)

	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if utxo.Asset.Tkn != nil {
				if _, ok := self.usedFlag.Load(utxo.Root); !ok {
					utxos = append(utxos, utxo)
					remain.Sub(remain, utxo.Asset.Tkn.Value.ToIntRef())
					if remain.Sign() <= 0 {
						break
					}
				}
			}
		}
	}
	return
}

func DecOuts(outs []txtool.Out, skr *keys.PKr) (douts []txtool.DOut) {
	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := txtool.DOut{}

		if out.State.OS.Out_O != nil {
			dout.Asset = out.State.OS.Out_O.Asset.Clone()
			dout.Memo = out.State.OS.Out_O.Memo
			dout.Nil = cpt.GenTil(&sk, out.State.OS.RootCM)
		} else {
			key, flag := keys.FetchKey(&sk, &out.State.OS.Out_Z.RPK)
			info_desc := cpt.InfoDesc{}
			info_desc.Key = key
			info_desc.Flag = flag
			info_desc.Einfo = out.State.OS.Out_Z.EInfo
			cpt.DecOutput(&info_desc)

			if e := stx.ConfirmOut_Z(&info_desc, out.State.OS.Out_Z); e == nil {
				dout.Asset = assets.NewAsset(
					&assets.Token{
						info_desc.Tkn_currency,
						utils.NewU256_ByKey(&info_desc.Tkn_value),
					},
					&assets.Ticket{
						info_desc.Tkt_category,
						info_desc.Tkt_value,
					},
				)
				dout.Memo = info_desc.Memo

				dout.Nil = cpt.GenTil(&sk, out.State.OS.RootCM)
			}
		}
		douts = append(douts, dout)
	}
	return
}

type uint64Slice []uint64

func (c uint64Slice) Len() int {
	return len(c)
}
func (c uint64Slice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c uint64Slice) Less(i, j int) bool {
	return c[i] < c[j]
}

var fetchCount = uint64(5000)

func (self *Exchange) fetchBlockInfo() {
	if txtool.Ref_inst.Bc == nil || !txtool.Ref_inst.Bc.IsValid() {
		return
	}
	for {
		indexs := map[uint64][]keys.Uint512{}
		orders := uint64Slice{}
		self.numbers.Range(func(key, value interface{}) bool {
			pk := key.(keys.Uint512)
			num := value.(uint64)
			if list, ok := indexs[num]; ok {
				indexs[num] = append(list, pk)
			} else {
				indexs[num] = []keys.Uint512{pk}
				orders = append(orders, num)
			}
			return true
		})
		if orders.Len() == 0 {
			return
		}

		sort.Sort(orders)
		start := orders[0]
		end := start + fetchCount
		if orders.Len() > 1 {
			end = orders[1]
		}

		pks := indexs[start]
		for end > start {
			count := fetchCount
			if end-start < fetchCount {
				count = end - start
			}
			if count == 0 {
				return
			}
			if self.fetchAndIndexUtxo(start, count, pks) < int(count) {
				return
			}
			start += count
		}
	}
}

func (self *Exchange) fetchAndIndexUtxo(start, countBlock uint64, pks []keys.Uint512) (count int) {

	blocks, err := flight.SRI_Inst.GetBlocksInfo(start, countBlock)
	if err != nil {
		log.Info("Exchange GetBlocksInfo", "error", err)
		return
	}

	if len(blocks) == 0 {
		return
	}

	utxosMap := map[PkKey][]Utxo{}
	nilsMap := map[keys.Uint256]Utxo{}
	nils := []keys.Uint256{}
	blockMap := map[uint64]*BlockInfo{}
	for _, block := range blocks {
		num := uint64(block.Num)
		utxos := []Utxo{}
		for _, out := range block.Outs {
			var pkr keys.PKr

			if out.State.OS.Out_Z != nil {
				pkr = out.State.OS.Out_Z.PKr
			}
			if out.State.OS.Out_O != nil {
				pkr = out.State.OS.Out_O.Addr
			}

			account, ok := self.ownPkr(pks, pkr)
			if !ok {
				continue
			}

			key := PkKey{PK: *account.pk, Num: out.State.Num}
			dout := DecOuts([]txtool.Out{out}, &account.skr)[0]
			utxo := Utxo{Pkr: pkr, Root: out.Root, Nil: dout.Nil, TxHash: out.State.TxHash, Num: out.State.Num, Asset: dout.Asset, IsZ: out.State.OS.Out_Z != nil}
			//log.Info("DecOuts", "PK", base58.EncodeToString(account.pk[:]), "root", common.Bytes2Hex(out.Root[:]), "currency", common.BytesToString(utxo.Asset.Tkn.Currency[:]), "value", utxo.Asset.Tkn.Value)
			nilsMap[utxo.Root] = utxo
			nilsMap[utxo.Nil] = utxo

			if list, ok := utxosMap[key]; ok {
				utxosMap[key] = append(list, utxo)
			} else {
				utxosMap[key] = []Utxo{utxo}
			}
			utxos = append(utxos, utxo)
		}

		if len(utxos) > 0 {
			blockMap[num] = &BlockInfo{Num: num, Hash: block.Hash, Outs: utxos}
		}

		if len(block.Nils) > 0 {
			roots := []keys.Uint256{}
			for _, Nil := range block.Nils {
				var utxo Utxo
				if value, ok := nilsMap[Nil]; ok {
					utxo = value
				} else {
					value, _ := self.db.Get(nilKey(Nil))
					if value != nil {
						var root keys.Uint256
						copy(root[:], value[98:130])

						if utxo, err = self.getUtxo(root); err != nil {
							continue
						} else {
							var pk keys.Uint512
							copy(pk[:], value[2:66])
						}
					} else {
						continue
					}
				}
				nils = append(nils, Nil)
				roots = append(roots, utxo.Root)
			}
			if len(roots) > 0 {
				if blockInfo, ok := blockMap[num]; ok {
					blockInfo.Ins = roots
				} else {
					blockMap[num] = &BlockInfo{Num: num, Hash: block.Hash, Ins: roots}
				}
			}
		}
	}

	batch := self.db.NewBatch()

	self.indexPkgs(pks, batch, blocks)

	var roots []keys.Uint256
	if len(utxosMap) > 0 || len(nils) > 0 {
		if roots, err = self.indexBlocks(batch, utxosMap, blockMap, nils); err != nil {
			log.Error("indexBlocks ", "error", err)
			return
		}
	}

	count = len(blocks)
	num := uint64(blocks[count-1].Num) + 1
	// "NUM"+PK  => Num
	data := utils.EncodeNumber(num)
	for _, pk := range pks {
		batch.Put(numKey(pk), data)
	}

	err = batch.Write()
	if err == nil {
		for _, pk := range pks {
			self.numbers.Store(pk, num)
		}
	}

	for _, root := range roots {
		self.usedFlag.Delete(root)
	}
	log.Info("Exchange indexed", "blockNumber", num-1)
	return
}

func (self *Exchange) indexBlocks(batch serodb.Batch, utxosMap map[PkKey][]Utxo, blockMap map[uint64]*BlockInfo, nils []keys.Uint256) (delRoots []keys.Uint256, err error) {
	ops := map[string]string{}

	for num, blockInfo := range blockMap {
		data, e := rlp.EncodeToBytes(blockInfo)
		if e != nil {
			err = e
			return
		}
		batch.Put(blockKey(num), data)
	}

	txMap := map[keys.Uint256][]Utxo{}
	for key, list := range utxosMap {
		roots := []keys.Uint256{}
		for _, utxo := range list {
			data, e := rlp.EncodeToBytes(utxo)
			if e != nil {
				err = e
				return
			}

			// "ROOT" + root
			batch.Put(rootKey(utxo.Root), data)
			//nil => root
			batch.Put(nilToRootKey(utxo.Nil), utxo.Root[:])

			var pkKeys []byte
			if utxo.Asset.Tkn != nil {
				// "PK" + PK + currency + root
				pkKey := utxoPkKey(key.PK, utxo.Asset.Tkn.Currency[:], &utxo.Root)
				ops[common.Bytes2Hex(pkKey)] = common.Bytes2Hex([]byte{0})
				pkKeys = append(pkKeys, pkKey...)
			}

			if utxo.Asset.Tkt != nil {
				// "PK" + PK + tkt + root
				pkKey := utxoPkKey(key.PK, utxo.Asset.Tkt.Value[:], &utxo.Root)
				ops[common.Bytes2Hex(pkKey)] = common.Bytes2Hex([]byte{0})
				pkKeys = append(pkKeys, pkKey...)
			}
			// "PK" + PK + currency + root => 0

			// "NIL" + PK + tkt + root => "PK" + PK + currency + root
			nilkey := nilKey(utxo.Nil)
			rootkey := nilKey(utxo.Root)

			// "NIL" +nil/root => pkKey
			ops[common.Bytes2Hex(nilkey)] = common.Bytes2Hex(pkKeys)
			ops[common.Bytes2Hex(rootkey)] = common.Bytes2Hex(pkKeys)

			roots = append(roots, utxo.Root)

			if list, ok := txMap[utxo.TxHash]; ok {
				txMap[utxo.TxHash] = append(list, utxo)
			} else {
				txMap[utxo.TxHash] = []Utxo{utxo}
			}

			//log.Info("Index add", "PK", base58.EncodeToString(key.PK[:]), "Nil", common.Bytes2Hex(utxo.Nil[:]), "root", common.Bytes2Hex(utxo.Root[:]), "Value", utxo.Asset.Tkn.Value)
		}

		data, e := rlp.EncodeToBytes(roots)
		if e != nil {
			err = e
			return
		}
		// blockNumber + PK => [roots]
		batch.Put(utxoKey(key.Num, key.PK), data)

		if account := self.getAccountByPk(key.PK); account != nil {
			account.isChanged = true
		}
	}

	for txHash, list := range txMap {
		data, err := rlp.EncodeToBytes(list)
		if err != nil {
			return nil, err
		}
		batch.Put(txKey(txHash), data)
	}

	for _, Nil := range nils {

		var pk keys.Uint512
		key := nilKey(Nil)
		hex := common.Bytes2Hex(key)
		if value, ok := ops[hex]; ok {
			delete(ops, hex)
			if len(value) == 260 {
				delete(ops, value)
			} else {
				delete(ops, value[0:260])
				delete(ops, value[260:])
			}

			var root keys.Uint256
			copy(root[:], value[98:130])
			delete(ops, common.Bytes2Hex(nilKey(root)))
			//self.usedFlag.Delete(root)
			delRoots = append(delRoots, root)

			copy(pk[:], value[2:66])
		} else {
			value, _ := self.db.Get(key)
			if value != nil {
				if len(value) == 130 {
					batch.Delete(value)
				} else {
					batch.Delete(value[0:130])
					batch.Delete(value[130:260])
				}
				batch.Delete(nilKey(Nil))

				var root keys.Uint256
				copy(root[:], value[98:130])
				batch.Delete(nilKey(root))
				//self.usedFlag.Delete(root)
				delRoots = append(delRoots, root)

				copy(pk[:], value[2:66])
			}
		}

		if account := self.getAccountByPk(pk); account != nil {
			account.isChanged = true
		}
	}

	for key, value := range ops {
		batch.Put(common.Hex2Bytes(key), common.Hex2Bytes(value))
	}

	return
}

func (self *Exchange) ownPkr(pks []keys.Uint512, pkr keys.PKr) (account *Account, ok bool) {
	for _, pk := range pks {
		value, ok := self.accounts.Load(pk)
		if !ok {
			continue
		}
		account = value.(*Account)
		if keys.IsMyPKr(account.tk, &pkr) {
			return account, true
		}
	}
	return
}

type MergeUtxos struct {
	list    UtxoList
	amount  big.Int
	zcount  int
	ocount  int
	tickets map[keys.Uint256]keys.Uint256
}

func (self *Exchange) getMergeUtxos(from *keys.Uint512, currency string, zcount int, left int) (mu MergeUtxos, e error) {
	if zcount > 400 {
		e = errors.New("zout count must <= 400")
	}
	prefix := utxoPkKey(*from, common.LeftPadBytes([]byte(currency), 32), nil)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	outxos := UtxoList{}
	zutxos := UtxoList{}
	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if _, ok := self.usedFlag.Load(utxo.Root); !ok {
				if utxo.IsZ {
					zutxos = append(zutxos, utxo)
				} else {
					outxos = append(outxos, utxo)
				}

			}
		}
		if zutxos.Len() >= zcount+left || outxos.Len() >= 2400+left {
			break
		}
	}
	if outxos.Len() >= 2400 {
		zutxos = UtxoList{}
	}
	mu.ocount = outxos.Len()
	mu.zcount = zutxos.Len()
	utxos := append(zutxos, outxos...)
	if utxos.Len() <= left {
		e = fmt.Errorf("no need to merge the account, utxo count == %v", utxos.Len())
		return
	}
	sort.Sort(utxos)
	mu.list = utxos[0 : utxos.Len()-(left-1)]
	for _, utxo := range mu.list {
		mu.amount.Add(&mu.amount, utxo.Asset.Tkn.Value.ToIntRef())
		if utxo.Asset.Tkt != nil {
			mu.tickets[utxo.Asset.Tkt.Value] = utxo.Asset.Tkt.Category
		}
	}
	mu.amount.Sub(&mu.amount, new(big.Int).Mul(big.NewInt(25000), big.NewInt(1000000000)))
	return
}

type MergeParam struct {
	From     keys.Uint512
	To       *keys.PKr
	Currency string
	Zcount   uint64
	Left     uint64
}

func (self *Exchange) GenMergeTx(mp *MergeParam) (txParam *txtool.GTxParam, e error) {
	account := self.getAccountByPk(mp.From)
	if account == nil {
		e = errors.New("account is nil")
		return
	}
	if mp.To == nil {
		mp.To = &account.mainPkr
	}
	var mu MergeUtxos
	if mu, e = self.getMergeUtxos(&mp.From, mp.Currency, int(mp.Zcount), int(mp.Left)); e != nil {
		return
	}
	bytes := common.LeftPadBytes([]byte(mp.Currency), 32)
	var Currency keys.Uint256
	copy(Currency[:], bytes[:])

	receptions := []prepare.Reception{{Addr: *mp.To, Asset: assets.Asset{Tkn: &assets.Token{Currency: Currency, Value: utils.U256(mu.amount)}}}}

	if len(mu.tickets) > 0 {
		for value, category := range mu.tickets {
			receptions = append(receptions, prepare.Reception{Addr: *mp.To, Asset: assets.Asset{Tkt: &assets.Ticket{category, value}}})
		}
	}
	txParam, e = self.buildTxParam(
		mu.list.Roots(),
		mp.To,
		receptions,
		nil,
		&assets.Token{
			utils.CurrencyToUint256("SERO"),
			utils.NewU256(25000),
		},
		big.NewInt(1000000000),
	)
	if e != nil {
		return
	}
	return
}

func (self *Exchange) Merge(pk *keys.Uint512, currency string, force bool) (count int, txhash keys.Uint256, e error) {
	account := self.getAccountByPk(*pk)
	if account == nil {
		e = errors.New("account is nil")
		return
	}

	seed, err := account.wallet.GetSeed()
	if err != nil || seed == nil {
		e = errors.New("account is locked")
		return
	}

	var mu MergeUtxos
	if mu, e = self.getMergeUtxos(pk, currency, 100, 10); e != nil {
		return
	}

	if mu.zcount >= 100 || mu.ocount >= 2400 || time.Now().After(account.nextMergeTime) || force {

		bytes := common.LeftPadBytes([]byte(currency), 32)
		var Currency keys.Uint256
		copy(Currency[:], bytes[:])

		receptions := []prepare.Reception{{Addr: account.mainPkr, Asset: assets.Asset{Tkn: &assets.Token{Currency: Currency, Value: utils.U256(mu.amount)}}}}

		if len(mu.tickets) > 0 {
			for value, category := range mu.tickets {
				receptions = append(receptions, prepare.Reception{Addr: account.mainPkr, Asset: assets.Asset{Tkt: &assets.Ticket{category, value}}})
			}
		}

		pretx, gtx, err := self.genTx(
			mu.list.Roots(),
			account,
			nil,
			receptions,
			nil,
			&assets.Token{
				utils.CurrencyToUint256("SERO"),
				utils.NewU256(25000),
			},
			big.NewInt(1000000000),
		)
		if err != nil {
			account.nextMergeTime = time.Now().Add(time.Hour * 6)
			e = err
			return
		}
		txhash = gtx.Hash
		if err := self.commitTx(gtx); err != nil {
			account.nextMergeTime = time.Now().Add(time.Hour * 6)
			self.ClearTxParam(pretx)
			e = err
			return
		}
		if mu.list.Len() < 100 {
			account.nextMergeTime = time.Now().Add(time.Hour * 6)
		}
		return
	} else {
		e = fmt.Errorf("no need to merge the account, utxo count == %v", mu.list.Len())
		return
	}
}

func (self *Exchange) merge() {
	if txtool.Ref_inst.Bc == nil || !txtool.Ref_inst.Bc.IsValid() {
		return
	}
	self.accounts.Range(func(key, value interface{}) bool {
		account := value.(*Account)
		if count, txhash, err := self.Merge(account.pk, "SERO", false); err != nil {
			log.Error("autoMerge fail", "PK", cpt.Base58Encode(account.pk[:]), "count", count, "error", err)
		} else {
			log.Info("autoMerge succ", "PK", cpt.Base58Encode(account.pk[:]), "tx", hexutil.Encode(txhash[:]), "count", count)
		}
		return true
	})

}

var (
	numPrefix  = []byte("NUM")
	pkPrefix   = []byte("PK")
	utxoPrefix = []byte("UTXO")
	rootPrefix = []byte("ROOT")
	nilPrefix  = []byte("NIL")

	blockPrefix   = []byte("BLOCK")
	outUtxoPrefix = []byte("OUTUTXO")
	txPrefix      = []byte("TX")
	nilRootPrefix = []byte("NOILTOROOT")
)

func nilToRootKey(nil keys.Uint256) []byte {
	return append(nilRootPrefix, nil[:]...)
}

func txKey(txHash keys.Uint256) []byte {
	return append(txPrefix, txHash[:]...)
}

func blockKey(number uint64) []byte {
	return append(blockPrefix, utils.EncodeNumber(number)...)
}

func numKey(pk keys.Uint512) []byte {
	return append(numPrefix, pk[:]...)
}

func nilKey(nil keys.Uint256) []byte {
	return append(nilPrefix, nil[:]...)
}

func rootKey(root keys.Uint256) []byte {
	return append(rootPrefix, root[:]...)
}

//func outUtxoKey(number uint64, pk keys.Uint512) []byte {
//	return append(outUtxoPrefix, append(encodeNumber(number), pk[:]...)...)
//}

// utxoKey = PK + currency +root
func utxoPkKey(pk keys.Uint512, currency []byte, root *keys.Uint256) []byte {
	key := append(pkPrefix, pk[:]...)
	if len(currency) > 0 {
		key = append(key, currency...)
	}
	if root != nil {
		key = append(key, root[:]...)
	}
	return key
}

func utxoKey(number uint64, pk keys.Uint512) []byte {
	return append(utxoPrefix, append(utils.EncodeNumber(number), pk[:]...)...)
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
