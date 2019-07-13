package app

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/robfig/cron"
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"github.com/sero-cash/go-sero/zero/txtool/prepare"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	numPrefix  = []byte("NUM")
	pkPrefix   = []byte("PK")
	pkrPrefix  = []byte("PKR")
	utxoPrefix = []byte("UTXO")
	rootPrefix = []byte("ROOT")
	nilPrefix  = []byte("NIL")
	syncNilKEY = []byte("SYNCNILNUM")

	blockPrefix   = []byte("BLOCK")
	outUtxoPrefix = []byte("OUTUTXO")
	txPrefix      = []byte("TX")
	nilRootPrefix = []byte("NOILTOROOT")
)

const maxUint64 = ^uint64(0)

// PKR + PK + r
func pkrKey(pk keys.Uint512, r keys.Uint256) []byte {
	key := append(pkrPrefix, pk[:]...)
	key = append(key, r[:]...)
	return key
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

type SEROLight struct {
	db *serodb.LDBDatabase

	accountManager *accounts.Manager

	accounts     sync.Map
	accountIndex int8

	pkrAccounts sync.Map
	usedFlag    sync.Map

	numbers sync.Map

	sli flight.SLI

	feed    event.Feed
	updater event.Subscription        // Wallet update subscriptions for all backends
	update  chan accounts.WalletEvent // Subscription sink for backend wallet changes
	quit    chan chan error
	lock    sync.RWMutex
}

func makeAccountManager() (*accounts.Manager, string, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	keydir := GetKeystorePath()
	var ephemeral string
	var err error
	if err != nil {
		return nil, "", err
	}
	// Assemble the account manager and supported backends
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}
	return accounts.NewManager(backends...), ephemeral, nil
}

var current_client *SEROLight

func CurrentClient() *SEROLight {
	return current_client
}

func NewSeroLight() {

	accountManager, _, err := makeAccountManager()
	if err != nil {
		logex.Fatalf("makeAccountManager, err=[%v]", err)
	}

	update := make(chan accounts.WalletEvent, 1)
	updater := accountManager.Subscribe(update)

	client := &SEROLight{
		accountManager: accountManager,
		update:         update,
		updater:        updater,
		accountIndex:   0,
	}
	current_client = client

	db, err := serodb.NewLDBDatabase(GetDataPath(), 1024, 1024)
	if err != nil {
		panic(err)
	}
	client.db = db

	client.numbers = sync.Map{}
	client.accounts = sync.Map{}
	client.usedFlag = sync.Map{}

	for _, w := range accountManager.Wallets() {
		client.initWallet(w)
	}
	AddJob("0/10 * * * * ?", client.SyncOut)

	AddJob("0/10 * * * * ?", client.CheckNil)

	client.pkrAccounts = sync.Map{}

	go client.keystoreListener()
}

var fetchCount = uint64(5000)

type PkKey struct {
	PK  keys.Uint512
	Num uint64
}

func (self *SEROLight) CurrentBlock() uint64 {
	number := uint64(0)
	self.numbers.Range(func(key, value interface{}) bool {
		num := value.(uint64)
		if number < num {
			number = num
		}
		return true
	})
	return number
}

func (self *SEROLight) keystoreListener() {
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
				fmt.Println("WalletDropped: ", base58.Encode(pk[:]))
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

func (self *SEROLight) initWallet(w accounts.Wallet) {

	if _, ok := self.accounts.Load(*w.Accounts()[0].Address.ToUint512()); !ok {
		account := Account{}
		account.wallet = w
		account.pk = w.Accounts()[0].Address.ToUint512()
		account.tk = w.Accounts()[0].Tk.ToUint512()
		copy(account.skr[:], account.tk[:])

		fmt.Println("initWallet: ", base58.Encode(account.pk[:]))
		account.mainPkr = self.createPkr(account.pk, 1)

		self.accountIndex = self.accountIndex + 1
		account.index = self.accountIndex

		self.accounts.Store(*account.pk, &account)
		account.isChanged = true
		if num := self.starNum(account.pk); num > w.Accounts()[0].At {
			self.numbers.Store(*account.pk, num)
		} else {
			self.numbers.Store(*account.pk, w.Accounts()[0].At)
		}
		logex.Info("Add PK", "address", w.Accounts()[0].Address, "At", self.GetCurrencyNumber(*account.pk))
	}
}

type pkrAndIndex struct {
	pkr   keys.PKr
	index uint64
}

func (self *SEROLight) getPKrsForQueryByPk(pk keys.Uint512) (pais []pkrAndIndex) {
	prefix := append(pkrPrefix, pk[:]...)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	count := 0
	for iterator.Next() {

		pai := pkrAndIndex{}
		key := iterator.Key()
		keyLen := len(key)
		pai.index = decodeNumber(key[keyLen-8:])
		// remove index=0 , save latest five pkrs
		if count > 5 {
			pais = append(pais[:1], pais[2:]...)
		}
		value := iterator.Value()
		var pkr keys.PKr
		copy(pkr[:], value[:])
		pai.pkr = pkr
		pais = append(pais, pai)
		count++
	}
	return pais
}

func (self *SEROLight) starNum(pk *keys.Uint512) uint64 {
	value, err := self.db.Get(numKey(*pk))
	if err != nil {
		return 0
	}
	return decodeNumber(value)
}

func (self *SEROLight) GetCurrencyNumber(pk keys.Uint512) uint64 {
	value, ok := self.numbers.Load(pk)
	if !ok {
		return 0
	}
	if value.(uint64) == 0 {
		return value.(uint64)
	}
	return value.(uint64) - 1
}

func encodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func (self *SEROLight) createPkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(encodeNumber(index), 32))
	pkr := keys.Addr2PKr(pk, &r)
	self.db.Put(pkrKey(*pk, r), pkr[:])
	logex.Info("createPkr::: ", base58.Encode(pkr[:]))

	return pkr
}

func (self *SEROLight) createPkrBatch(batch serodb.Batch, pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(encodeNumber(index), 32))
	pkr := keys.Addr2PKr(pk, &r)
	batch.Put(pkrKey(*pk, r), pkr[:])
	return pkr
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

type BlockOutResp struct {
	CurrencyNum uint64
	BlockOuts   []BlockOut
}

type BlockOut struct {
	Num  uint64
	Outs []txtool.Out
}

type BlockInfo struct {
	Num  uint64
	Hash keys.Uint256
	Ins  []keys.Uint256
	Outs []Utxo
}

func nilToRootKey(nil keys.Uint256) []byte {
	return append(nilRootPrefix, nil[:]...)
}

func (self *SEROLight) SyncOut() {

	fmt.Println("SyncOut begin");
	self.numbers.Range(func(key, value interface{}) bool {
		pk := key.(keys.Uint512)
		current := value.(uint64)
		var start = current + 1
		var end = start + fetchCount
		for {
			logex.Infof("sync begin,start=[%d],end=[%d],pk=[%v]", start, end, base58.Encode(pk[:]))
			current = self.FetchAndIndex(start, end, pk)
			if current < end {
				break
			} else {
				start = current + 1
				end = start + fetchCount
			}
		}

		return true
	})
}

func (self *SEROLight) getMinBlockNum() uint64 {
	minBlockNum := uint64(0)
	self.numbers.Range(func(key, value interface{}) bool {
		//pk := key.(keys.Uint512)
		current := value.(uint64)
		if current > minBlockNum {
			minBlockNum = current
		}
		return true
	})
	return minBlockNum
}

func (self *SEROLight) FetchAndIndex(start, end uint64, pk keys.Uint512) (current uint64) {

	account := self.getAccountByPk(pk)
	pkrAndIndexs := self.getPKrsForQueryByPk(pk)

	var pkrs []string
	//确保取到prk的最大下标
	var maxPkrIndex uint64
	for _, pkrAndIndex := range pkrAndIndexs {
		if maxPkrIndex < pkrAndIndex.index {
			maxPkrIndex = pkrAndIndex.index
		}
		pkrs = append(pkrs, base58.Encode(pkrAndIndex.pkr[:]))
	}

	sync := Sync{RpcHost: "http://127.0.0.1:8545", Method: "light_getOutsByPKr", Params: []interface{}{pkrs, start, end}}

	jsonResp, err := sync.Do()
	if err != nil {
		logex.Error("jsonRep err=[%s]", err.Error())
		return
	}

	bor := BlockOutResp{}
	if err = json.Unmarshal(*jsonResp.Result, &bor); err != nil {
		logex.Error("json.Unmarshal err=[%s]", err.Error())
		return
	}
	utxosMap := map[PkKey][]Utxo{}

	currentBlockNum := self.CurrentBlock()
	blockOuts := bor.BlockOuts
	if currentBlockNum < bor.CurrencyNum {
		currentBlockNum = bor.CurrencyNum
	}

	var utxoNum uint64 = 0
	cUtxoNum := self.GetUtxoNum(pk)
	for _, blockOut := range blockOuts {
		outs := blockOut.Outs
		for _, out := range outs {
			var pkr keys.PKr
			if out.State.OS.Out_Z != nil {
				pkr = out.State.OS.Out_Z.PKr
			}
			if out.State.OS.Out_O != nil {
				pkr = out.State.OS.Out_O.Addr
			}

			dout := DecOuts([]txtool.Out{out}, &account.skr)[0]
			num := uint64(0)
			if dout.Asset.Tkn != nil {
				currency := common.BytesToString(dout.Asset.Tkn.Currency[:])
				fmt.Println("cUtxoNum:", cUtxoNum)
				num = maxUint64 - cUtxoNum[currency] - utxoNum
			} else {
				num = maxUint64 - out.State.Num
			}
			fmt.Println("num: ", num)
			key := PkKey{PK: *account.pk, Num: num}

			utxo := Utxo{Pkr: pkr, Root: out.Root, Nil: dout.Nil, TxHash: out.State.TxHash, Num: out.State.Num, Asset: dout.Asset, IsZ: out.State.OS.Out_Z != nil, Out: out}
			//log.Info("DecOuts", "PK", base58.Encode(account.pk[:]), "root", common.Bytes2Hex(out.Root[:]), "currency", common.BytesToString(utxo.Asset.Tkn.Currency[:]), "value", utxo.Asset.Tkn.Value)
			if list, ok := utxosMap[key]; ok {
				utxosMap[key] = append(list, utxo)
			} else {
				utxosMap[key] = []Utxo{utxo}
			}
			utxoNum++
		}
	}

	batch := self.db.NewBatch()
	if len(utxosMap) > 0 {
		ops,err := self.indexUtxo(utxosMap, batch)
		if err !=nil{
			return
		}
		for key, value := range ops {
			batch.Put(common.Hex2Bytes(key), common.Hex2Bytes(value))
		}
		//PRK+1
		self.createPkrBatch(batch, &pk, maxPkrIndex+1)
	}

	// "NUM"+PK  => Num
	batch.Put(numKey(pk), encodeNumber(currentBlockNum))

	err = batch.Write()

	if err == nil {
		if ac := self.getAccountByPk(pk); ac != nil {
			ac.isChanged = true
		}
		self.numbers.Store(pk, currentBlockNum)
	}

	return
}

func (self *SEROLight) indexUtxo(utxosMap map[PkKey][]Utxo, batch serodb.Batch) (opsReturn map[string]string ,err error) {
	ops := map[string]string{}
	for key, list := range utxosMap {
		roots := []keys.Uint256{}
		for _, utxo := range list {
			data, err := rlp.EncodeToBytes(utxo)
			if err != nil {
				return nil,err
			}

			// "ROOT" + root
			batch.Put(rootKey(utxo.Root), data)
			//nil => root
			batch.Put(nilToRootKey(utxo.Nil), utxo.Root[:])

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
			//log.Info("Index add", "PK", base58.Encode(key.PK[:]), "Nil", common.Bytes2Hex(utxo.Nil[:]), "root", common.Bytes2Hex(utxo.Root[:]), "Value", utxo.Asset.Tkn.Value)
		}
		data, err := rlp.EncodeToBytes(roots)
		if err != nil {
			return nil,err
		}

		// utxo PK + index  => [roots]
		batch.Put(utxoKey(key.Num, key.PK), data)
	}
	return ops,nil
}

type BlockDelNil struct {
	Num uint64
	Nil keys.Uint256
}

func (self *SEROLight) CheckNil() {

	minNum := self.getMinBlockNum()
	nilNum, err := self.db.Get(syncNilKEY)
	start := uint64(0)
	if len(nilNum) > 0 {
		start = decodeNumber(nilNum)
	}
	//start = start + 1

	end := minNum
	if start >= minNum {
		return
	}
	if start >= end {
		return
	}

	iterator := self.db.NewIteratorWithPrefix(nilPrefix)
	Nils := []keys.Uint256{}

	for iterator.Next() {
		key := iterator.Key()
		var Nil keys.Uint256
		copy(Nil[:], key[3:])
		nilkey := nilKey(Nil)
		value, _ := self.db.Get(nilkey)
		if value != nil {
			Nils = append(Nils, Nil)
		}
	}

	sync := Sync{RpcHost: "http://127.0.0.1:8545", Method: "light_checkNil", Params: []interface{}{Nils, start, end}}

	jsonResp, err := sync.Do()
	if err != nil {
		logex.Error("jsonRep err=[%s]", err.Error())
		return
	}
	batch := self.db.NewBatch()
	maxNilNum := end
	if jsonResp.Result != nil {
		bdns := []BlockDelNil{}
		if err = json.Unmarshal(*jsonResp.Result, &bdns); err != nil {
			logex.Error("json.Unmarshal err=[%s]", err.Error())
			return
		}
		for _, bdn := range bdns {
			var pk keys.Uint512
			Nil := bdn.Nil
			value, _ := self.db.Get(nilKey(Nil))
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
				self.usedFlag.Delete(root)
				copy(pk[:], value[2:66])
				logex.Info("delete :", root)
			}
			if account := self.getAccountByPk(pk); account != nil {
				account.isChanged = true
			}
			if maxNilNum < bdn.Num {
				maxNilNum = bdn.Num
			}
		}
	}
	batch.Put(syncNilKEY, encodeNumber(maxNilNum))
	batch.Write()
}

func (self *SEROLight) getAccountByPk(pk keys.Uint512) *Account {
	if value, ok := self.accounts.Load(pk); ok {
		return value.(*Account)
	}
	return nil
}

// "UTXO" + pk + number
func utxoKey(number uint64, pk keys.Uint512) []byte {
	return append(utxoPrefix, append(pk[:], encodeNumber(number)...)...)
}

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

func (self *SEROLight) ownPkr(pks []keys.Uint512, pkr keys.PKr) (account *Account, ok bool) {
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

func (self *SEROLight) GetUtxoNum(pk keys.Uint512) map[string]uint64 {
	if account := self.getAccountByPk(pk); account != nil {
		return account.utxoNums
	}
	return map[string]uint64{}
}

func (self *SEROLight) GetBalances(pk keys.Uint512) (balances map[string]*big.Int) {
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
						if utxo.flag == 0 { //available utxo
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

func (self *SEROLight) GetRecordsByPk(PK *keys.Uint512, begin, end uint64) (records []Utxo, err error) {
	err = self.iteratorUtxo(PK, begin, end, func(utxo Utxo) {
		records = append(records, utxo)
	})
	return
}

type (
	HandleUtxoFunc func(utxo Utxo)
)

func (self *SEROLight) iteratorUtxo(PK *keys.Uint512, begin, end uint64, handler HandleUtxoFunc) (e error) {
	var pk keys.Uint512
	if PK != nil {
		pk = *PK
	}

	//get the first num
	iteratorT := self.db.NewIteratorWithPrefix(append(utxoPrefix, pk[:]...))
	firstNum := uint64(0)
	for iteratorT.Next() {
		key := iteratorT.Key()
		firstNum = decodeNumber(key[68:76])
		break
	}

	begin = begin + firstNum
	end = end + firstNum - 1

	iterator := self.db.NewIteratorWithPrefix(utxoPrefix)
	for ok := iterator.Seek(utxoKey(begin, pk)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := decodeNumber(key[68:76])
		//num := decodeNumber(key[4:12])
		if num > end {
			break
		}
		copy(pk[:], key[4:68])

		if PK != nil && *PK != pk {
			continue
		}

		value := iterator.Value()
		roots := []keys.Uint256{}
		if err := rlp.Decode(bytes.NewReader(value), &roots); err != nil {
			logex.Error("Invalid roots RLP", "PK", common.Bytes2Hex(pk[:]), "blockNumber", num, "err", err)
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

func (self *SEROLight) getUtxo(root keys.Uint256) (utxo Utxo, e error) {
	data, err := self.db.Get(rootKey(root))
	if err != nil {
		return
	}
	if err := rlp.Decode(bytes.NewReader(data), &utxo); err != nil {
		logex.Error("Light Invalid utxo RLP", "root", common.Bytes2Hex(root[:]), "err", err)
		e = err
		return
	}
	if value, ok := self.usedFlag.Load(utxo.Root); ok {
		utxo.flag = value.(int)
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

type Utxo struct {
	Pkr    keys.PKr
	Root   keys.Uint256
	TxHash keys.Uint256
	Nil    keys.Uint256
	Num    uint64
	Asset  assets.Asset
	IsZ    bool
	flag   int
	Out    txtool.Out
}

func NewBigIntFromString(s string, base int) (*big.Int, error) {
	val, flag := big.NewInt(0).SetString(s, base)
	if !flag {
		return nil, fmt.Errorf("can't convert %s to BigInt", s)
	}
	return val, nil
}

func (self *SEROLight) CommitTx(from, to, currency, passwd string, amount, gasprice *big.Int) (hash keys.Uint256, err error) {

	fee := new(big.Int).Mul(big.NewInt(25000), gasprice)
	fromPk := address.Base58ToAccount(from).ToUint512()

	var RefundTo *keys.PKr
	ac := self.getAccountByPk(*fromPk)
	if ac != nil {
		RefundTo = &ac.mainPkr
	}

	preTxParam := prepare.PreTxParam{}
	preTxParam.From = *fromPk
	preTxParam.RefundTo = RefundTo
	preTxParam.GasPrice = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(9), nil)
	preTxParam.Fee = assets.Token{Currency: utils.CurrencyToUint256("SERO"), Value: utils.NewU256(fee.Uint64())}

	var toPkr keys.PKr
	copy(toPkr[:], base58.Decode(to)[:])
	tkn := assets.Token{Currency: utils.CurrencyToUint256(currency), Value: utils.NewU256(amount.Uint64())}
	asset := assets.NewAsset(&tkn, nil)
	reception := prepare.Reception{
		Addr:  toPkr,
		Asset: asset,
	}

	preTxParam.Receptions = []prepare.Reception{reception}
	param, err := self.GenTx(preTxParam)

	if err != nil {
		return hash, err
	}

	account := accounts.Account{Address: ac.wallet.Accounts()[0].Address}

	wallet, err := self.accountManager.Find(account)
	if err != nil {
		return hash, err
	}
	seed, err := wallet.GetSeedWithPassphrase(passwd)
	if err != nil {
		return hash, err
	}
	sk := keys.Seed2Sk(seed.SeedToUint256())

	gtx, err := flight.SignTx(&sk, param)
	if err != nil {
		return hash, err
	}

	hash = gtx.Hash

	sync := Sync{RpcHost: "http://127.0.0.1:8545", Method: "sero_commitTx", Params: []interface{}{gtx}}
	if _, err := sync.Do(); err != nil {
		return hash, err
	}

	return hash, nil
}

func (self *SEROLight) buildOut(to string, currency string, amount *big.Int) txtool.GOut {
	var pkr keys.PKr
	copy(pkr[:], base58.Decode(to)[:])
	asset := assets.Asset{Tkn: &assets.Token{
		Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
		Value:    utils.U256(*amount),
	},
	}
	return txtool.GOut{PKr: pkr, Asset: asset}
}

// === jsonrpc post

type Sync struct {
	RpcHost string
	Method  string
	Params  interface{}
}

func (sync Sync) Do() (*JSONRpcResp, error) {
	client := &http.Client{
		Timeout: 900 * time.Second,
	}
	//logex.Info("sync.Params=", sync.Params)
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": sync.Method, "params": sync.Params, "id": 0}
	data, err := json.Marshal(jsonReq)
	if err != nil {
		logex.Error(err.Error())
		return nil, err
	}
	logex.Info(string(data))

	req, err := http.NewRequest("POST", sync.RpcHost, bytes.NewBuffer(data))
	if err != nil {
		logex.Error(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logex.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp *JSONRpcResp
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		logex.Error(err.Error())
		return nil, err
	}
	if rpcResp.Error != nil {
		logex.Error(rpcResp.Error)
		return nil, fmt.Errorf(rpcResp.Error["message"].(string))
	}
	return rpcResp, err
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}
