package exchange

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sero-cash/go-sero/common/base58"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero/zero/light/light_ref"

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
	"github.com/sero-cash/go-sero/zero/light"
	"github.com/sero-cash/go-sero/zero/light/light_types"
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

type Reception struct {
	Addr     keys.PKr
	Currency string
	Value    *big.Int
}

type TxParam struct {
	From       keys.Uint512
	RefundTo   *keys.PKr
	Receptions []Reception
	Gas        uint64
	GasPrice   *big.Int
	Roots      []keys.Uint256
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

	sri light.SRI
	sli light.SLI

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
		sri:            light.SRI_Inst,
		sli:            light.SLI_Inst,
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
		account.mainPkr = self.createPkr(account.pk, 1)
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
	return decodeNumber(value)
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
		num := decodeNumber(key[5:13])
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

func (self *Exchange) GenTx(param TxParam) (txParam *light_types.GenTxParam, e error) {
	utxos, err := self.preGenTx(param)
	if err != nil {
		return nil, err
	}

	if value, ok := self.accounts.Load(param.From); ok {
		var refundTo keys.PKr
		if param.RefundTo == nil {
			account := value.(*Account)
			refundTo = account.mainPkr
		} else {
			refundTo = *param.RefundTo
		}
		txParam, e = self.buildTxParam(utxos, &refundTo, param.Receptions, param.Gas, param.GasPrice)
	} else {
		return nil, errors.New("not found Pk")
	}

	return
}

func (self *Exchange) GenTxWithSign(param TxParam) (pretx *light_types.GenTxParam, tx *light_types.GTx, e error) {
	if self == nil {
		e = errors.New("exchange instance is nil")
		return
	}
	var utxos []Utxo
	if utxos, e = self.preGenTx(param); e != nil {
		return
	}

	var account *Account
	if value, ok := self.accounts.Load(param.From); ok {
		account = value.(*Account)
	} else {
		e = errors.New("not found Pk")
		return
	}

	if pretx, tx, e = self.genTx(utxos, account, param.Receptions, param.Gas, param.GasPrice); e != nil {
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

func (self *Exchange) isPk(addr keys.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func (self *Exchange) createPkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(encodeNumber(index), 32))
	return keys.Addr2PKr(pk, &r)
}

func (self *Exchange) preGenTx(param TxParam) (utxos []Utxo, err error) {
	var roots []keys.Uint256
	if len(param.Roots) > 0 {
		roots = param.Roots
		for _, root := range roots {
			utxo, err := self.getUtxo(root)
			if err != nil {
				return utxos, err
			}
			utxos = append(utxos, utxo)
		}
	} else {
		amounts := map[string]*big.Int{}
		for _, each := range param.Receptions {
			if amount, ok := amounts[each.Currency]; ok {
				amount.Add(amount, each.Value)
			} else {
				amounts[each.Currency] = new(big.Int).Set(each.Value)
			}
		}
		if amount, ok := amounts["SERO"]; ok {
			amount.Add(amount, new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), param.GasPrice))
		} else {
			amounts["SERO"] = new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), param.GasPrice)
		}
		for currency, amount := range amounts {
			list, remain := self.findUtxos(&param.From, currency, amount)
			if remain.Sign() > 0 {
				return utxos, errors.New(fmt.Sprintf("not enough token, maximum available token is %s", new(big.Int).Sub(amount, remain).String()))
			} else {
				utxos = append(utxos, list...)
			}
		}
	}
	count := 0
	for _, each := range utxos {
		if !each.IsZ {
			count++
		}
	}
	if count > 2500 {
		err = errors.New("ins.len > 2500")
	}
	return
}

func (self *Exchange) ClearTxParam(txParam *light_types.GenTxParam) (count int) {
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

func (self *Exchange) genTx(utxos []Utxo, account *Account, receptions []Reception, gas uint64, gasPrice *big.Int) (txParam *light_types.GenTxParam, tx *light_types.GTx, e error) {
	if txParam, e = self.buildTxParam(utxos, &account.mainPkr, receptions, gas, gasPrice); e != nil {
		return
	}

	var seed *address.Seed
	if seed, e = account.wallet.GetSeed(); e != nil {
		self.ClearTxParam(txParam)
		return
	}

	sk := keys.Seed2Sk(seed.SeedToUint256())
	gtx, err := light.SignTx(&sk, txParam)
	if err != nil {
		self.ClearTxParam(txParam)
		e = err
		return
	} else {
		tx = &gtx
		return
	}
}

func (self *Exchange) buildTxParam(
	utxos []Utxo,
	refundTo *keys.PKr,
	receptions []Reception,
	gas uint64,
	gasPrice *big.Int) (txParam *light_types.GenTxParam, e error) {

	txParam = new(light_types.GenTxParam)
	txParam.Gas = gas
	txParam.GasPrice = *gasPrice

	txParam.From = light_types.Kr{PKr: *refundTo}

	roots := []keys.Uint256{}
	for _, utxo := range utxos {
		roots = append(roots, utxo.Root)
	}
	Ins := []light_types.GIn{}
	wits, err := self.sri.GetAnchor(roots)
	if err != nil {
		e = err
		return
	}

	amounts := make(map[string]*big.Int)
	ticekts := make(map[keys.Uint256]keys.Uint256)
	for index, utxo := range utxos {
		if out := light.GetOut(&utxo.Root, 0); out != nil {
			Ins = append(Ins, light_types.GIn{Out: light_types.Out{Root: utxo.Root, State: *out}, Witness: wits[index]})

			if utxo.Asset.Tkn != nil {
				currency := strings.Trim(string(utxo.Asset.Tkn.Currency[:]), string([]byte{0}))
				if amount, ok := amounts[currency]; ok {
					amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
				} else {
					amounts[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
				}

			}
			if utxo.Asset.Tkt != nil {
				ticekts[utxo.Asset.Tkt.Value] = utxo.Asset.Tkt.Category
			}
		}
	}

	Outs := []light_types.GOut{}
	for _, reception := range receptions {
		currency := strings.ToUpper(reception.Currency)
		if amount, ok := amounts[currency]; ok && amount.Cmp(reception.Value) >= 0 {

			if self.isPk(reception.Addr) {
				pk := reception.Addr.ToUint512()
				pkr := self.createPkr(&pk, 1)
				Outs = append(Outs, light_types.GOut{PKr: pkr, Asset: assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
					Value:    utils.U256(*reception.Value),
				}}})
			} else {
				Outs = append(Outs, light_types.GOut{PKr: reception.Addr, Asset: assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
					Value:    utils.U256(*reception.Value),
				}}})
			}

			amount.Sub(amount, reception.Value)
			if amount.Sign() == 0 {
				delete(amounts, currency)
			}
		}

	}

	fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), gasPrice)
	if amount, ok := amounts["SERO"]; !ok || amount.Cmp(fee) < 0 {
		e = fmt.Errorf("Exchange Error: not enough")
		return
	} else {
		amount.Sub(amount, fee)
		if amount.Sign() == 0 {
			delete(amounts, "SERO")
		}
	}

	if len(amounts) > 0 {
		for currency, value := range amounts {
			Outs = append(Outs, light_types.GOut{PKr: txParam.From.PKr, Asset: assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*value),
			}}})
		}
	}
	if len(ticekts) > 0 {
		for value, category := range ticekts {
			Outs = append(Outs, light_types.GOut{PKr: txParam.From.PKr, Asset: assets.Asset{Tkt: &assets.Ticket{
				Category: category,
				Value:    value,
			}}})
		}
	}

	txParam.Ins = Ins
	txParam.Outs = Outs

	for _, utxo := range utxos {
		self.usedFlag.Store(utxo.Root, 1)
	}

	return
}

func (self *Exchange) commitTx(tx *light_types.GTx) (err error) {
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
		num := decodeNumber(key[4:12])
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

func (self *Exchange) findUtxos(pk *keys.Uint512, currency string, amount *big.Int) (utxos []Utxo, remain *big.Int) {
	remain = new(big.Int).Set(amount)

	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	//list := []Utxo{}
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

			//else {
			//	list = append(list, utxo)
			//}
		}
	}
	//
	//if remain.Sign() > 0 {
	//	for _, utxo := range list {
	//		utxos = append(utxos, utxo)
	//		remain.Sub(remain, utxo.Asset.Tkn.Value.ToIntRef())
	//		if remain.Sign() <= 0 {
	//			break
	//		}
	//	}
	//}
	return
}

func DecOuts(outs []light_types.Out, skr *keys.PKr) (douts []light_types.DOut) {
	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	for _, out := range outs {
		dout := light_types.DOut{}

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
	if light_ref.Ref_inst.Bc == nil || !light_ref.Ref_inst.Bc.IsValid() {
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

	blocks, err := self.sri.GetBlocksInfo(start, countBlock)
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
			dout := DecOuts([]light_types.Out{out}, &account.skr)[0]
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
			nils = append(nils, block.Nils...)
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
	if len(utxosMap) > 0 || len(nils) > 0 {
		if err := self.indexBlocks(batch, utxosMap, blockMap, nils); err != nil {
			log.Error("indexBlocks ", "error", err)
			return
		}
	}

	count = len(blocks)
	num := uint64(blocks[count-1].Num) + 1
	// "NUM"+PK  => Num
	data := encodeNumber(num)
	for _, pk := range pks {
		batch.Put(numKey(pk), data)
	}

	err = batch.Write()
	if err == nil {
		for _, pk := range pks {
			self.numbers.Store(pk, num)
		}
	}
	log.Info("Exchange indexed", "blockNumber", num-1)
	return
}

func (self *Exchange) indexBlocks(batch serodb.Batch, utxosMap map[PkKey][]Utxo, blockMap map[uint64]*BlockInfo, nils []keys.Uint256) (err error) {
	ops := map[string]string{}

	for num, blockInfo := range blockMap {
		data, err := rlp.EncodeToBytes(blockInfo)
		if err != nil {
			return err
		}
		batch.Put(blockKey(num), data)
	}

	txMap := map[keys.Uint256][]Utxo{}
	for key, list := range utxosMap {
		roots := []keys.Uint256{}
		for _, utxo := range list {
			data, err := rlp.EncodeToBytes(utxo)
			if err != nil {
				return err
			}

			// "ROOT" + root
			batch.Put(rootKey(utxo.Root), data)

			var pkKey []byte
			if utxo.Asset.Tkn != nil {
				// "PK" + PK + currency + root
				pkKey = utxoPkKey(key.PK, utxo.Asset.Tkn.Currency[:], &utxo.Root)

			} else if utxo.Asset.Tkt != nil {
				// "PK" + PK + tkt + root
				pkKey = utxoPkKey(key.PK, utxo.Asset.Tkt.Value[:], &utxo.Root)
			}
			// "PK" + PK + currency + root => 0
			ops[common.Bytes2Hex(pkKey)] = common.Bytes2Hex([]byte{0})

			// "NIL" + PK + tkt + root => "PK" + PK + currency + root
			nilkey := nilKey(utxo.Nil)
			rootkey := nilKey(utxo.Root)

			// "NIL" +nil/root => pkKey
			ops[common.Bytes2Hex(nilkey)] = common.Bytes2Hex(pkKey)
			ops[common.Bytes2Hex(rootkey)] = common.Bytes2Hex(pkKey)

			roots = append(roots, utxo.Root)

			if list, ok := txMap[utxo.TxHash]; ok {
				txMap[utxo.TxHash] = append(list, utxo)
			} else {
				txMap[utxo.TxHash] = []Utxo{utxo}
			}

			//log.Info("Index add", "PK", base58.EncodeToString(key.PK[:]), "Nil", common.Bytes2Hex(utxo.Nil[:]), "root", common.Bytes2Hex(utxo.Root[:]), "Value", utxo.Asset.Tkn.Value)
		}

		data, err := rlp.EncodeToBytes(roots)
		if err != nil {
			return err
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
			return err
		}
		batch.Put(txKey(txHash), data)
	}

	for _, Nil := range nils {

		var pk keys.Uint512
		key := nilKey(Nil)
		hex := common.Bytes2Hex(key)
		if value, ok := ops[hex]; ok {
			delete(ops, hex)
			delete(ops, value)
			var root keys.Uint256
			copy(root[:], value[98:130])
			delete(ops, common.Bytes2Hex(nilKey(root)))
			self.usedFlag.Delete(root)

			copy(pk[:], value[2:66])
		} else {
			value, _ := self.db.Get(key)
			if value != nil {
				batch.Delete(value)
				batch.Delete(nilKey(Nil))

				var root keys.Uint256
				copy(root[:], value[98:130])
				batch.Delete(nilKey(root))
				self.usedFlag.Delete(root)

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

	return nil
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
	list   UtxoList
	amount big.Int
	zcount int
	ocount int
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

func (self *Exchange) GenMergeTx(mp *MergeParam) (txParam *light_types.GenTxParam, e error) {
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
	txParam, e = self.buildTxParam(mu.list, mp.To, []Reception{{Value: &mu.amount, Currency: mp.Currency, Addr: *mp.To}}, 25000, big.NewInt(1000000000))
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
		pretx, gtx, err := self.genTx(mu.list, account, []Reception{{Value: &mu.amount, Currency: currency, Addr: account.mainPkr}}, 25000, big.NewInt(1000000000))
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
	if light_ref.Ref_inst.Bc == nil || !light_ref.Ref_inst.Bc.IsValid() {
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
	Prefix        = []byte("Out")
)

func txKey(txHash keys.Uint256) []byte {
	return append(txPrefix, txHash[:]...)
}

func blockKey(number uint64) []byte {
	return append(blockPrefix, encodeNumber(number)...)
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
	return append(utxoPrefix, append(encodeNumber(number), pk[:]...)...)
}

func encodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func decodeNumber(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
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
