package exchange

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	sk            *keys.PKr
	skr           keys.PKr
	mainPkr       keys.PKr
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
	Receptions []Reception
	Gas        uint64
	GasPrice   *big.Int
	Roots      []keys.Uint256
}

type (
	HandleUtxoFunc func(utxo Utxo)
)

type PkKey struct {
	Pkr keys.PKr
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

	accounts    map[keys.Uint512]*Account
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
	exchange.accounts = map[keys.Uint512]*Account{}
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
	if _, ok := self.accounts[*w.Accounts()[0].Address.ToUint512()]; !ok {
		account := Account{}
		account.wallet = w
		account.pk = w.Accounts()[0].Address.ToUint512()
		account.tk = w.Accounts()[0].Tk.ToUint512()
		copy(account.skr[:], account.tk[:])
		account.mainPkr = self.createPkr(account.pk, 1)
		account.nextMergeTime = time.Now()
		self.accounts[*account.pk] = &account

		if num := self.starNum(account.pk); num > 0 {
			self.numbers.Store(*account.pk, num)
		} else {
			self.numbers.Store(*account.pk, w.Accounts()[0].At)
		}

		log.Info("Add PK", "address", w.Accounts()[0].Address)
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

func (self *Exchange) GetCurrencyNumber(pk keys.Uint512) uint64 {
	value, ok := self.numbers.Load(pk)
	if !ok {
		return 0
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
	if _, ok := self.accounts[*pk]; !ok {
		return pkr, errors.New("not found Pk")
	}

	return keys.Addr2PKr(pk, index), nil
}

func (self *Exchange) GetBalances(pk keys.Uint512) (balances map[string]*big.Int) {

	prefix := append(pkPrefix, pk[:]...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	balances = map[string]*big.Int{}
	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if utxo.Asset.Tkn != nil {
				currency := common.BytesToString(utxo.Asset.Tkn.Currency[:])
				if amount, ok := balances[currency]; ok {
					amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
				} else {
					balances[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
				}
			}
		}
	}
	return
}

func (self *Exchange) GetRecords(pkr keys.PKr, begin, end uint64) (records []Utxo, err error) {
	err = self.iteratorUtxo(pkr, begin, end, func(utxo Utxo) {
		records = append(records, utxo)
	})
	return
}

func (self *Exchange) GenTx(param TxParam) (txParam *light_types.GenTxParam, e error) {
	utxos, err := self.preGenTx(param)
	if err != nil {
		return nil, err
	}

	if _, ok := self.accounts[param.From]; !ok {
		return nil, errors.New("not found Pk")
	}

	txParam, e = self.buildTxParam(utxos, self.accounts[param.From], param.Receptions, param.Gas, param.GasPrice)
	return
}

func (self *Exchange) GenTxWithSign(param TxParam) (*light_types.GTx, error) {
	utxos, err := self.preGenTx(param)
	if err != nil {
		return nil, err
	}

	var account *Account
	if _, ok := self.accounts[param.From]; !ok {
		return nil, errors.New("not found Pk")
	} else {
		account = self.accounts[param.From]
	}

	gtx, err := self.genTx(utxos, account, param.Receptions, param.Gas, param.GasPrice)
	if err != nil {
		log.Error("Exchange genTx", "error", err)
		return nil, err
	}
	gtx.Hash = gtx.Tx.ToHash()
	log.Info("Exchange genTx success")
	return gtx, nil
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
			amount = new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), param.GasPrice)
		}
		for currency, amount := range amounts {
			if list, err := self.findUtxos(&param.From, currency, amount); err != nil {
				return utxos, err
			} else {
				utxos = append(utxos, list...)
			}
		}
	}
	return
}

//func (self *Exchange) CommitTx(gtx light_types.GTx) (err error) {
//	return self.commitTx(&gtx)
//}

func (self *Exchange) isPk(addr keys.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func (self *Exchange) createPkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(encodeNumber(index), 32))
	return keys.Addr2PKr(pk, &r)
}

func (self *Exchange) genTx(utxos []Utxo, account *Account, receptions []Reception, gas uint64, gasPrice *big.Int) (*light_types.GTx, error) {
	txParam, err := self.buildTxParam(utxos, account, receptions, gas, gasPrice)
	if err != nil {
		return nil, err
	}

	if account.sk == nil {
		seed, err := account.wallet.GetSeed()
		if err != nil {
			return nil, err
		}
		sk := keys.Seed2Sk(seed.SeedToUint256())
		account.sk = new(keys.PKr)
		copy(account.sk[:], sk[:])
	}

	txParam.From.SKr = *account.sk
	for index := range txParam.Ins {
		txParam.Ins[index].SKr = *account.sk
	}

	gtx, err := self.sli.GenTx(txParam)
	if err != nil {
		return nil, err
	}
	return &gtx, nil
}

func (self *Exchange) buildTxParam(utxos []Utxo, account *Account, receptions []Reception, gas uint64, gasPrice *big.Int) (txParam *light_types.GenTxParam, e error) {
	txParam = new(light_types.GenTxParam)
	txParam.Gas = gas
	txParam.GasPrice = *gasPrice

	txParam.From = light_types.Kr{PKr: account.mainPkr}

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
		e = fmt.Errorf("SSI GenTx Error: not enough")
		return
	} else {
		amount.Sub(amount, fee)
		if amount.Sign() == 0 {
			delete(amounts, "SERO")
		}
	}

	if len(amounts) > 0 {
		for currency, value := range amounts {
			Outs = append(Outs, light_types.GOut{PKr: account.mainPkr, Asset: assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*value),
			}}})
		}
	}
	if len(ticekts) > 0 {
		for value, category := range ticekts {
			Outs = append(Outs, light_types.GOut{PKr: account.mainPkr, Asset: assets.Asset{Tkt: &assets.Ticket{
				Category: category,
				Value:    value,
			}}})
		}
	}

	txParam.Ins = Ins
	txParam.Outs = Outs

	for _, utxo := range utxos {
		self.usedFlag.Store(utxo.Nil, 1)
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

func (self *Exchange) initAccount(pkr keys.PKr) (err error) {
	if _, ok := self.isMyPkr(pkr); !ok {
		return
	}

	var account *PkrAccount
	if value, ok := self.pkrAccounts.Load(pkr); ok {
		account = value.(*PkrAccount)

	} else {
		account = &PkrAccount{}
		account.Pkr = pkr
		account.balances = map[string]*big.Int{}
		self.pkrAccounts.Store(pkr, account)
	}

	err = self.iteratorUtxo(pkr, account.num+1, math.MaxUint64, func(utxo Utxo) {
		if utxo.Asset.Tkn != nil {
			curency := strings.ToUpper(common.BytesToString(utxo.Asset.Tkn.Currency[:]))
			if balance, ok := account.balances[curency]; ok {
				balance.Add(balance, utxo.Asset.Tkn.Value.ToIntRef())
			} else {
				account.balances[curency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
			}
			account.num = utxo.Num
		}
	})

	return
}

func (self *Exchange) iteratorUtxo(pkr keys.PKr, begin, end uint64, handler HandleUtxoFunc) (e error) {
	iterator := self.db.NewIteratorWithPrefix(append(pkrPrefix, pkr[:]...))
	for ok := iterator.Seek(utxoPkrKey(pkr, begin)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := decodeNumber(key[99:107])
		if num > end {
			break
		}

		value := iterator.Value()
		roots := []keys.Uint256{}
		if err := rlp.Decode(bytes.NewReader(value), &roots); err != nil {
			log.Error("Invalid roots RLP", "pkr", common.Bytes2Hex(pkr[:]), "blockNumber", num, "err", err)
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

	if value, ok := self.usedFlag.Load(utxo.Nil); ok {
		utxo.flag = value.(int)
	}
	return
}

func (self *Exchange) findUtxos(pk *keys.Uint512, currency string, amount *big.Int) (utxos []Utxo, e error) {
	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	list := UtxoList{}
	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if _, ok := self.usedFlag.Load(utxo.Nil); !ok {
				utxos = append(utxos, utxo)
				amount.Sub(amount, utxo.Asset.Tkn.Value.ToIntRef())
			} else {
				list = append(list, utxo)
			}
		}
		if amount.Sign() <= 0 {
			break
		}
	}

	if amount.Sign() > 0 {
		if list.Len() > 0 {
			sort.Sort(list)
			for _, utxo := range list {
				utxos = append(utxos, utxo)
				amount.Sub(amount, utxo.Asset.Tkn.Value.ToIntRef())
				if amount.Sign() <= 0 {
					break
				}
			}
		}
	}

	if amount.Sign() > 0 {
		e = errors.New("not enough token")
	}
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
		//list := []string{}
		//for _, pk := range pks {
		//	list = append(list, base58.EncodeToString(pk[:]))
		//}

		for end > start {
			count := fetchCount
			if end-start < fetchCount {
				count = end - start
			}
			if count == 0 {
				return
			}
			//log.Info("fetchAndIndexUtxo", "start", start, "count", count, "pks", list)
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

	utxosMap := map[keys.Uint512]map[PkrKey][]Utxo{}
	nils := []keys.Uint256{}
	num := start
	for _, block := range blocks {
		for _, out := range block.Outs {
			var pkr keys.PKr

			if out.State.OS.Out_Z != nil {
				pkr = out.State.OS.Out_Z.PKr
			}
			if out.State.OS.Out_O != nil {
				pkr = out.State.OS.Out_O.Addr
			}

			key := PkrKey{pkr: pkr, num: out.State.Num}
			account, ok := self.ownPkr(pks, pkr)
			if !ok {
				continue
			}

			dout := DecOuts([]light_types.Out{out}, &account.skr)[0]
			utxo := Utxo{Pkr: pkr, Root: out.Root, Nil: dout.Nil, TxHash: out.State.TxHash, Num: out.State.Num, Asset: dout.Asset}
			if pkrMap, ok := utxosMap[*account.pk]; ok {
				if list, ok := pkrMap[key]; ok {
					pkrMap[key] = append(list, utxo)
				} else {
					pkrMap[key] = []Utxo{utxo}
				}
			} else {
				utxosMap[*account.pk] = map[PkrKey][]Utxo{key: {utxo}}
			}
		}
		if len(block.Nils) > 0 {
			nils = append(nils, block.Nils...)
		}
	}

	batch := self.db.NewBatch()
	if len(utxosMap) > 0 || len(nils) > 0 {
		if err := self.indexBlocks(batch, utxosMap, nils); err != nil {
			log.Error("indexBlocks ", "error", err)
		}
	}

	count = len(blocks)
	num = uint64(blocks[count-1].Num) + 1
	// "NUM"+pk  => num
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

func (self *Exchange) indexBlocks(batch serodb.Batch, utxosMap map[keys.Uint512]map[PkrKey][]Utxo, nils []keys.Uint256) (err error) {
	ops := map[string]string{}
	for pk, pkrMap := range utxosMap {
		rootsMap := map[uint64][]keys.Uint256{}
		for key, list := range pkrMap {
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
					// "PK" + pk + currency + root
					pkKey = utxoPkKey(pk, utxo.Asset.Tkn.Currency[:], &utxo.Root)

				} else if utxo.Asset.Tkt != nil {
					// "PK" + pk + tkt + root
					pkKey = utxoPkKey(pk, utxo.Asset.Tkt.Value[:], &utxo.Root)
				}
				// "PK" + pk + currency + root => 0
				ops[common.Bytes2Hex(pkKey)] = common.Bytes2Hex([]byte{0})

				// "NIL" + pk + tkt + root => "PK" + pk + currency + root
				nilkey := nilKey(utxo.Nil)
				rootkey := nilKey(utxo.Root)

				// "NIL" +nil/root => pkKey
				ops[common.Bytes2Hex(nilkey)] = common.Bytes2Hex(pkKey)
				ops[common.Bytes2Hex(rootkey)] = common.Bytes2Hex(pkKey)

				roots = append(roots, utxo.Root)
				//log.Info("Index add", "PK", base58.EncodeToString(pk[:]), "Nil", common.Bytes2Hex(utxo.Nil[:]), "Key", common.Bytes2Hex(pkKey[:]), "Value", utxo.Asset.Tkn.Value)
			}

			data, err := rlp.EncodeToBytes(roots)
			if err != nil {
				return err
			}
			// "PKR" + prk + blockNumber => [roots]
			batch.Put(utxoPkrKey(key.pkr, key.num), data)
			if list, ok := rootsMap[key.num]; ok {
				rootsMap[key.num] = append(list, roots...)
			} else {
				rootsMap[key.num] = roots
			}
		}

		for num, roots := range rootsMap {
			data, err := rlp.EncodeToBytes(roots)
			if err != nil {
				return err
			}
			pkr := keys.PKr{}
			copy(pkr[:], pk[:])
			batch.Put(utxoPkrKey(pkr, num), data)
		}
	}

	for _, Nil := range nils {

		key := nilKey(Nil)
		hex := common.Bytes2Hex(key)
		if value, ok := ops[hex]; ok {
			delete(ops, hex)
			delete(ops, value)
			var root keys.Uint256
			copy(root[:], value[98:130])
			delete(ops, common.Bytes2Hex(nilKey(root)))

			//log.Info("Index del", "nil", common.Bytes2Hex(Nil[:]), "key", value)
		} else {
			data, _ := self.db.Get(key)
			if data != nil {
				batch.Delete(data)
				batch.Delete(nilKey(Nil))

				var root keys.Uint256
				copy(root[:], data[98:130])
				batch.Delete(nilKey(root))
				//log.Info("Index del", "nil", common.Bytes2Hex(Nil[:]), "key", common.Bytes2Hex(data))
			}
		}
		self.usedFlag.Delete(common.Bytes2Hex(Nil[:]))
	}

	for key, value := range ops {
		batch.Put(common.Hex2Bytes(key), common.Hex2Bytes(value))
	}

	return nil
}

func (self *Exchange) ownPkr(pks []keys.Uint512, pkr keys.PKr) (account *Account, ok bool) {
	for _, pk := range pks {
		account = self.accounts[pk]
		if account == nil {
			continue
		}
		if keys.IsMyPKr(account.tk, &pkr) {
			return account, true
		}
	}
	return
}

func (self *Exchange) isMyPkr(pkr keys.PKr) (account *Account, ok bool) {
	for _, account := range self.accounts {
		if keys.IsMyPKr(account.tk, &pkr) {
			return account, true
		}
	}
	return nil, false
}

func (self *Exchange) merge() {
	for _, account := range self.accounts {
		seed, err := account.wallet.GetSeed()
		if err != nil || seed == nil {
			continue
		}
		for {
			prefix := utxoPkKey(*account.pk, common.LeftPadBytes([]byte("SERO"), 32), nil)
			iterator := self.db.NewIteratorWithPrefix(prefix)
			utxos := UtxoList{}
			for iterator.Next() {
				key := iterator.Key()
				var root keys.Uint256
				copy(root[:], key[98:130])

				if utxo, err := self.getUtxo(root); err == nil {
					if _, ok := self.usedFlag.Load(utxo.Nil); !ok {
						utxos = append(utxos, utxo)
					}
				}

				if utxos.Len() >= 108 {
					break
				}
			}
			if utxos.Len() > 100 || time.Now().After(account.nextMergeTime) {
				sort.Sort(utxos)
				utxos = utxos[0 : utxos.Len()-8]
				if utxos.Len() > 1 {
					amount := new(big.Int)
					for _, utxo := range utxos {
						amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
					}
					amount.Sub(amount, new(big.Int).Mul(big.NewInt(25000), big.NewInt(1000000000)))
					gtx, err := self.genTx(utxos, account, []Reception{{Value: amount, Currency: "SERO", Addr: account.mainPkr}}, 25000, big.NewInt(1000000000))
					if err != nil {
						log.Error("Exchange merge utxo", "error", err)
						continue
					}
					log.Info("Exchange merge utxo success ", "count", utxos.Len())
					self.commitTx(gtx)
				}
				if utxos.Len() < 100 {
					account.nextMergeTime = time.Now().Add(time.Hour * 6)
				}
			} else {
				break
			}
		}

	}
}

var (
	numPrefix  = []byte("NUM")
	pkPrefix   = []byte("PK")
	pkrPrefix  = []byte("PKR")
	rootPrefix = []byte("ROOT")
	nilPrefix  = []byte("NIL")

	Prefix = []byte("Out")
)

func numKey(pk keys.Uint512) []byte {
	return append(numPrefix, pk[:]...)
}

func nilKey(nil keys.Uint256) []byte {
	return append(nilPrefix, nil[:]...)
}

func rootKey(root keys.Uint256) []byte {
	return append(rootPrefix, root[:]...)
}

// utxoKey = pk + currency +root
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

func utxoPkrKey(pkr keys.PKr, number uint64) []byte {
	return append(pkrPrefix, append(pkr[:], encodeNumber(number)...)...)
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
