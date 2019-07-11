package app

import (
	"github.com/sero-cash/go-sero/accounts"
	"sync"
	"github.com/sero-cash/go-sero/event"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"encoding/binary"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/robfig/cron"
	"sync/atomic"
	"github.com/sero-cash/go-czero-import/cpt"
	"encoding/json"
	"math/big"
	"bytes"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
	"fmt"
	"time"
	"net/http"
	"github.com/sero-cash/go-sero/common/base58"
	"github.com/sero-cash/go-sero/zero/light"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-flight/util"
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

	accounts    sync.Map
	pkrAccounts sync.Map

	numbers sync.Map

	sli light.SLI

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
	}
	current_client = client

	db, err := serodb.NewLDBDatabase(GetDataPath(), 1024, 1024)
	if err != nil {
		panic(err)
	}
	client.db = db

	client.numbers = sync.Map{}
	client.accounts = sync.Map{}
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
				fmt.Println("WalletDropped: ", base58.EncodeToString(pk[:]))
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

		fmt.Println("initWallet: ", base58.EncodeToString(account.pk[:]))
		account.mainPkr = self.createPkr(account.pk, 1)

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
	logex.Info("createPkr::: ", base58.EncodeToString(pkr[:]))

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
	Outs []light_types.Out
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
			logex.Infof("sync begin,start=[%d],end=[%d],pk=[%v]",start,end,base58.EncodeToString(pk[:]))
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
		pkrs = append(pkrs, base58.EncodeToString(pkrAndIndex.pkr[:]))
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

			dout := DecOuts([]light_types.Out{out}, &account.skr)[0]
			num := uint64(0)
			if dout.Asset.Tkn != nil {
				currency := common.BytesToString(dout.Asset.Tkn.Currency[:])
				cUtxoNum := self.GetUtxoNum(pk)
				fmt.Println("cUtxoNum:", cUtxoNum)
				num = maxUint64 - cUtxoNum[currency] - utxoNum
			} else {
				num = maxUint64 - out.State.Num
			}
			fmt.Println("num: ", num)

			key := PkKey{PK: *account.pk, Num: num}

			utxo := Utxo{Pkr: pkr, Root: out.Root, Nil: dout.Nil, TxHash: out.State.TxHash, Num: out.State.Num, Asset: dout.Asset, IsZ: out.State.OS.Out_Z != nil, Flag: 0, Out: out}
			//log.Info("DecOuts", "PK", base58.EncodeToString(account.pk[:]), "root", common.Bytes2Hex(out.Root[:]), "currency", common.BytesToString(utxo.Asset.Tkn.Currency[:]), "value", utxo.Asset.Tkn.Value)
			if list, ok := utxosMap[key]; ok {
				utxosMap[key] = append(list, utxo)
			} else {
				utxosMap[key] = []Utxo{utxo}
			}
			utxoNum ++
		}
	}

	batch := self.db.NewBatch()
	if len(utxosMap) > 0 {
		ops := map[string]string{}
		for key, list := range utxosMap {
			roots := []keys.Uint256{}
			for _, utxo := range list {
				data, err := rlp.EncodeToBytes(utxo)
				if err != nil {
					return
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
				//log.Info("Index add", "PK", base58.EncodeToString(key.PK[:]), "Nil", common.Bytes2Hex(utxo.Nil[:]), "root", common.Bytes2Hex(utxo.Root[:]), "Value", utxo.Asset.Tkn.Value)
			}
			data, err := rlp.EncodeToBytes(roots)
			if err != nil {
				return
			}

			// utxo index + PK => [roots]
			batch.Put(utxoKey(key.Num, key.PK), data)
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
	start = start + 1

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
				var root keys.Uint256
				copy(root[:], value[98:130])
				copy(pk[:], value[2:66])

				batch.Delete(value)
				batch.Delete(nilKey(Nil))
				batch.Delete(nilKey(root))
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

// "UTXO" + number + pk
func utxoKey(number uint64, pk keys.Uint512) []byte {
	return append(utxoPrefix, append(encodeNumber(number), pk[:]...)...)
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
						if utxo.Flag == 0 {//available utxo
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
	iterator := self.db.NewIteratorWithPrefix(utxoPrefix)
	begin = maxUint64 - begin
	end = maxUint64 - end

	for ok := iterator.Seek(utxoKey(end, pk)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := decodeNumber(key[4:12])
		if num > begin {
			break
		}
		copy(pk[:], key[12:76])

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
		logex.Error("Exchange Invalid utxo RLP", "root", common.Bytes2Hex(root[:]), "err", err)
		e = err
		return
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

type Utxo struct {
	Pkr    keys.PKr
	Root   keys.Uint256
	TxHash keys.Uint256
	Nil    keys.Uint256
	Num    uint64
	Asset  assets.Asset
	IsZ    bool
	Flag   int
	Out    light_types.Out
}

func NewBigIntFromString(s string, base int) (*big.Int, error) {
	val, flag := big.NewInt(0).SetString(s, base)
	if !flag {
		return nil, fmt.Errorf("can't convert %s to BigInt", s)
	}
	return val, nil
}


func (self *SEROLight) CommitTx(from, to, currency, amountStr, gasPriceStr string) (hash string,err error) {

	amount, err := NewBigIntFromString(amountStr, 10)
	if err != nil {
		return hash,err
	}
	if amount.Sign() <0 {
		return hash,fmt.Errorf("amount < 0 ")
	}

	gasprice, err := NewBigIntFromString(gasPriceStr, 10)
	if err != nil {
		return hash, err
	}
	if gasprice.Sign() <0 {
		return hash,fmt.Errorf("gasprice < 0 ")
	}

	fee := new(big.Int).Mul(big.NewInt(25000), gasprice)

	fromPk := address.Base58ToAccount(from).ToUint512()
	txParam := light_types.GenTxParam{}
	txParam.Gas = fee.Uint64()
	txParam.GasPrice = *gasprice

	skr := keys.PKr{}
	if ac := self.getAccountByPk(*fromPk); ac != nil {
		skr = ac.skr
		txParam.From = light_types.Kr{SKr: ac.skr, PKr: ac.mainPkr}
	}
	utxoUpdates := []Utxo{}
	pageNo := uint64(1)
	pageSize := uint64(100)
	Ins := [] light_types.GIn{}
	sumOut := big.NewInt(0)
	total := big.NewInt(0).Add(amount, fee)
Loop:
	for {
		start := (pageNo - 1) * pageSize
		end := start + pageSize
		if utxos, err := self.GetRecordsByPk(fromPk, start, end); err == nil {
			if len(utxos) == 0 {
				break Loop
			}
			for _, utxo := range utxos {
				currencyUtxo := common.BytesToString(utxo.Asset.Tkn.Currency[:])
				if currencyUtxo != currency {
					continue
				}
				// has used
				if utxo.Flag !=0 {
					continue
				}


				sumOut = big.NewInt(0).Add(sumOut, utxo.Asset.Tkn.Value.ToIntRef())
				In := light_types.GIn{}
				In.Out = utxo.Out
				In.SKr = skr
				Ins = append(Ins, In)
				utxoUpdates = append(utxoUpdates,utxo)

				if sumOut.Cmp(total) >= 0 {
					break Loop
				}
			}
			pageNo++
		}

		break Loop
	}
	if sumOut.Cmp(total) < 0 {
		return hash,fmt.Errorf("not enough outs")
	}

	//交易生成
	roots := []keys.Uint256{}
	for _, in := range Ins {
		roots = append(roots, in.Out.Root)
	}

	witnesses, err := self.GetAnchor(roots)
	if err != nil {
		logex.Infof("Withdraw: GetAnchor error : %s", err)
		return hash,err
	}

	for index, witness := range witnesses {
		Ins[index].Witness = witness
	}
	txParam.Ins = Ins

	if sumOut.Cmp(total) == 0 {
		out := self.buildOut(to, currency, amount)
		txParam.Outs = []light_types.GOut{out}
	} else if (sumOut.Cmp(total) > 0) {
		amountMy := big.NewInt(0).Sub(sumOut, total)
		outTo := self.buildOut(to, currency, amount)
		outMy := self.buildOut(from, currency, amountMy)
		txParam.Outs = []light_types.GOut{outTo, outMy}
	}

	gtx, err := self.sli.GenTx(&txParam)
	if err != nil {
		return hash,err
	}

	sync := Sync{RpcHost: "http://127.0.0.1:8545", Method: "sero_commitTx", Params: []interface{}{gtx}}
	if _, err := sync.Do(); err != nil {
		return hash,err
	}

	// update utxoflg
	batch := self.db.NewBatch()
	for _,utxo :=range utxoUpdates{
		utxoNew := utxo
		utxoNew.Flag = 1
		data, err := rlp.EncodeToBytes(utxoNew)
		if err != nil {
			return hash,err
		}
		// "ROOT" + root
		batch.Put(rootKey(utxo.Root), data)
	}
	batch.Write()

	return hash,nil
}

func (self *SEROLight) buildOut(to string, currency string, amount *big.Int) light_types.GOut {
	var pkr keys.PKr
	copy(pkr[:], util.Base58ToBytes(to))
	asset := assets.Asset{Tkn: &assets.Token{
		Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
		Value:    utils.U256(*amount),
	},
	}
	return light_types.GOut{PKr: pkr, Asset: asset}
}

func (self *SEROLight) GetAnchor(roots []keys.Uint256) ([]light_types.Witness, error) {
	params := []string{}
	for _, each := range roots {
		params = append(params, hexutil.Encode(each[:]))
	}
	sync := Sync{RpcHost: "http://127.0.0.1:8545", Method: "sero_getAnchor", Params: []interface{}{params}}
	rpcResp, err := sync.Do()
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var witnesses []light_types.Witness
		err = json.Unmarshal(*rpcResp.Result, &witnesses)
		return witnesses, err
	}
	return nil, nil
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
