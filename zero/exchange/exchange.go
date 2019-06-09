package exchange

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/robfig/cron"
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/light"
	"github.com/sero-cash/go-sero/zero/light/light_ref"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type ExchangeConfig struct {
	autoMerge bool
}

type Account struct {
	wallet  accounts.Wallet
	pk      *keys.Uint512
	tk      *keys.Uint512
	sk      *keys.PKr
	skr     keys.PKr
	mainPkr keys.PKr
}

type PkrAccount struct {
	Pkr      keys.PKr
	balances map[string]*big.Int
	num      uint64
}

type Uxto struct {
	Root   keys.Uint256
	TxHash keys.Uint256
	Nil    keys.Uint256
	Num    uint64
	Asset  assets.Asset
	flag   int
}

type UxtoList []Uxto

func (list UxtoList) Len() int {
	return len(list)
}

func (list UxtoList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list UxtoList) Less(i, j int) bool {
	if list[i].flag == list[j].flag {
		return list[i].Asset.Tkn.Value.ToIntRef().Cmp(list[j].Asset.Tkn.Value.ToIntRef()) < 0
	} else {
		return list[i].flag < list[j].flag
	}
}

type Reception struct {
	Pkr      common.Address
	Currency string
	Value    *big.Int
}

type TxParam struct {
	Receptions []Reception
	Gas        uint64
	GasPrice   uint64
	Roots      []keys.Uint256
}

type (
	HandleUxtoFunc func(uxto Uxto)
)

type PkKey struct {
	Pkr keys.PKr
	Num uint64
}

type PkrKey struct {
	Pkr keys.PKr
	Num uint64
}

type Exchange struct {
	db     *serodb.LDBDatabase
	txPool *core.TxPool
	config ExchangeConfig
	//pkAccounts  map[keys.Uint512]*PkAccount
	accountManager *accounts.Manager

	account     Account
	pkrAccounts sync.Map

	sri light.SRI
	sli light.SLI

	usedFlag sync.Map
	inits    sync.Map

	lastBlockNumber uint64
}

func NewExchange(db *serodb.LDBDatabase, txPool *core.TxPool, accountManager *accounts.Manager) (exchange *Exchange) {
	exchange = &Exchange{
		db:     db,
		txPool: txPool,
		sri:    light.SRI_Inst,
		sli:    light.SLI_Inst,
	}

	exchange.account = Account{}
	for _, w := range accountManager.Wallets() {
		exchange.account.wallet = w
		exchange.account.pk = w.Accounts()[0].Address.ToUint512()
		exchange.account.tk = w.Accounts()[0].Tk.ToUint512()
		copy(exchange.account.skr[:], exchange.account.tk[:])
		exchange.account.mainPkr = exchange.createPkr(exchange.account.pk, 1)
		log.Info("PK", "address", w.Accounts()[0].Address)
		break
	}

	exchange.pkrAccounts = sync.Map{}

	exchange.usedFlag = sync.Map{}
	exchange.inits = sync.Map{}

	data, err := db.Get(lastBlockNumberKey)
	if err != nil {
		log.Warn("Exchange init lastBlockNumber", "error", err)
	} else {
		exchange.lastBlockNumber = decodeNumber(data)
	}

	AddJob("0/10 * * * * ?", exchange.fetchAndIndexUxto)

	//AddJob("0 0/1 * * * ?", exchange.merge)

	log.Info("Init NewExchange success")
	return
}

func (self *Exchange) CreatePkr(index uint64) (pkr keys.PKr) {
	if index < 100 {
		return
	}
	return self.createPkr(self.account.pk, index)
}

func (self *Exchange) GetBalances(address common.Address) (balances map[string]*big.Int) {
	if bytes.Equal(self.account.pk[:], address[0:64]) {
		prefix := append(pkPrefix, self.account.pk[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)

		balances = map[string]*big.Int{}
		for iterator.Next() {
			key := iterator.Key()
			var root keys.Uint256
			copy(root[:], key[98:130])

			if uxto, err := self.getUxto(root); err == nil {
				if uxto.Asset.Tkn != nil {
					currency := common.BytesToString(uxto.Asset.Tkn.Currency[:])
					if amount, ok := balances[currency]; ok {
						amount.Add(amount, uxto.Asset.Tkn.Value.ToIntRef())
					} else {
						balances[currency] = new(big.Int).Set(uxto.Asset.Tkn.Value.ToIntRef())
					}
				}
			}
		}
		return
	} else {

		pkr := *address.ToPKr()
		if _, ok := self.inits.LoadOrStore(pkr, 1); !ok {
			self.initAccount(pkr)
			self.inits.Delete(pkr)
		}
		if value, ok := self.pkrAccounts.Load(pkr); ok {
			return value.(*PkrAccount).balances
		}
	}
	return map[string]*big.Int{}
}

func (self *Exchange) GetRecords(pkr keys.PKr, begin, end uint64) (records []Uxto, err error) {
	if self.isMyPkr(pkr) {
		err = self.iteratorUxto(pkr, begin, end, func(uxto Uxto) {
			records = append(records, uxto)
		})
	}
	return
}

func (self *Exchange) GenTx(param TxParam) (txParam *light_types.GenTxParam, e error) {
	var roots []keys.Uint256
	uxtos := []Uxto{}
	if len(param.Roots) > 0 {
		roots = param.Roots
		for _, root := range roots {

			uxto, err := self.getUxto(root)
			if err != nil {
				e = err
				return
			}
			uxtos = append(uxtos, uxto)
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
		for currency, amount := range amounts {
			if list, err := self.findUxtos(self.account.pk, currency, amount); err != nil {
				e = err
				return
			} else {
				uxtos = append(uxtos, list...)
			}
		}
	}
	txParam, e = self.buildTxParam(uxtos, param.Receptions, param.Gas, param.GasPrice)
	return
}
func (self *Exchange) GenTxWithSign(param TxParam) (*light_types.GTx, error) {
	var roots []keys.Uint256
	uxtos := []Uxto{}
	if len(param.Roots) > 0 {
		roots = param.Roots
		for _, root := range roots {

			uxto, err := self.getUxto(root)
			if err != nil {
				return nil, err
			}
			uxtos = append(uxtos, uxto)
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
			amount.Add(amount, new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), new(big.Int).SetUint64(param.GasPrice)))
		} else {
			amount = new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), new(big.Int).SetUint64(param.GasPrice))
		}
		for currency, amount := range amounts {
			if list, err := self.findUxtos(self.account.pk, currency, amount); err != nil {
				return nil, err
			} else {
				uxtos = append(uxtos, list...)
			}
		}
	}

	gtx, err := self.genTx(uxtos, param.Receptions, param.Gas, param.GasPrice)
	if err != nil {
		log.Error("Exchange genTx", "error", err)
		return nil, err
	}
	gtx.Hash = gtx.Tx.ToHash()
	log.Info("Exchange genTx success", "roots", roots)
	return gtx, nil
}

//func (self *Exchange) CommitTx(gtx light_types.GTx) (err error) {
//	return self.commitTx(&gtx)
//}

func (self *Exchange) createPkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(encodeNumber(index), 32))
	return keys.Addr2PKr(pk, &r)
}

func (self *Exchange) genTx(uxtos []Uxto, receptions []Reception, gas, gasPrice uint64) (*light_types.GTx, error) {
	txParam, err := self.buildTxParam(uxtos, receptions, gas, gasPrice)
	if err != nil {
		return nil, err
	}

	if self.account.sk == nil {
		seed, err := self.account.wallet.GetSeed()
		if err != nil {
			return nil, err
		}
		sk := keys.Seed2Sk(seed.SeedToUint256())
		self.account.sk = new(keys.PKr)
		copy(self.account.sk[:], sk[:])
	}
	txParam.From.SKr = *self.account.sk
	for index, _ := range txParam.Ins {
		txParam.Ins[index].SKr = *self.account.sk
	}

	gtx, err := self.sli.GenTx(txParam)
	if err != nil {
		return nil, err
	}
	return &gtx, nil
}

func (self *Exchange) buildTxParam(uxtos []Uxto, receptions []Reception, gas, gasPrice uint64) (txParam *light_types.GenTxParam, e error) {
	txParam = new(light_types.GenTxParam)
	txParam.Gas = gas
	txParam.GasPrice = *big.NewInt(int64(gasPrice))

	txParam.From = light_types.Kr{PKr: self.account.mainPkr}

	roots := []keys.Uint256{}
	for _, uxtos := range uxtos {
		roots = append(roots, uxtos.Root)
	}
	Ins := []light_types.GIn{}
	wits, err := self.sri.GetAnchor(roots)
	if err != nil {
		e = err
		return
	}

	amounts := make(map[string]*big.Int)
	ticekts := make(map[keys.Uint256]keys.Uint256)
	for index, uxto := range uxtos {
		if out := localdb.GetRoot(light_ref.Ref_inst.Bc.GetDB(), &uxto.Root); out != nil {
			Ins = append(Ins, light_types.GIn{Out: light_types.Out{Root: uxto.Root, State: *out}, Witness: wits[index]})

			if uxto.Asset.Tkn != nil {
				currency := strings.Trim(string(uxto.Asset.Tkn.Currency[:]), string([]byte{0}))
				if amount, ok := amounts[currency]; ok {
					amount.Add(amount, uxto.Asset.Tkn.Value.ToIntRef())
				} else {
					amounts[currency] = new(big.Int).Set(uxto.Asset.Tkn.Value.ToIntRef())
				}

			}
			if uxto.Asset.Tkt != nil {
				ticekts[uxto.Asset.Tkt.Value] = uxto.Asset.Tkt.Category
			}
		}
	}

	Outs := []light_types.GOut{}
	for _, reception := range receptions {
		currency := strings.ToUpper(reception.Currency)
		if amount, ok := amounts[currency]; ok && amount.Cmp(reception.Value) >= 0 {

			Outs = append(Outs, light_types.GOut{PKr: *reception.Pkr.ToPKr(), Asset: assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*reception.Value),
			}}})

			amount.Sub(amount, reception.Value)
			if amount.Sign() == 0 {
				delete(amounts, currency)
			}
		}

	}

	fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), new(big.Int).SetUint64(gasPrice))
	if amount, ok := amounts["SERO"]; !ok || amount.Cmp(fee) < 0 {
		e = fmt.Errorf("SSI GenTx Error: not enough token")
		return
	} else {
		amount.Sub(amount, fee)
		if amount.Sign() == 0 {
			delete(amounts, "SERO")
		}
	}

	if len(amounts) > 0 {
		for currency, value := range amounts {
			Outs = append(Outs, light_types.GOut{PKr: self.account.mainPkr, Asset: assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*value),
			}}})
		}
	}
	if len(ticekts) > 0 {
		for value, category := range ticekts {
			Outs = append(Outs, light_types.GOut{PKr: self.account.mainPkr, Asset: assets.Asset{Tkt: &assets.Ticket{
				Category: category,
				Value:    value,
			}}})
		}
	}

	txParam.Ins = Ins
	txParam.Outs = Outs

	for _, uxto := range uxtos {
		self.usedFlag.Store(uxto.Nil, 1)
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
	if !self.isMyPkr(pkr) {
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

	err = self.iteratorUxto(pkr, account.num+1, math.MaxUint64, func(uxto Uxto) {
		if uxto.Asset.Tkn != nil {
			curency := strings.ToUpper(common.BytesToString(uxto.Asset.Tkn.Currency[:]))
			if balance, ok := account.balances[curency]; ok {
				balance.Add(balance, uxto.Asset.Tkn.Value.ToIntRef())
			} else {
				account.balances[curency] = new(big.Int).Set(uxto.Asset.Tkn.Value.ToIntRef())
			}
			account.num = uxto.Num
		}
	})

	return
}

func (self *Exchange) iteratorUxto(pkr keys.PKr, begin, end uint64, handler HandleUxtoFunc) (e error) {
	iterator := self.db.NewIteratorWithPrefix(append(pkrPrefix, pkr[:]...))
	for ok := iterator.Seek(uxtoPkrKey(pkr, begin)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := decodeNumber(key[99:107])
		if num > end {
			break
		}

		//var p keys.PKr
		//copy(p[:], key[3:99])
		//if p != pkr {
		//	break
		//}

		value := iterator.Value()
		roots := []keys.Uint256{}
		if err := rlp.Decode(bytes.NewReader(value), &roots); err != nil {
			log.Error("Invalid roots RLP", "pkr", common.Bytes2Hex(pkr[:]), "blockNumber", num, "err", err)
			e = err
			return
		}
		for _, root := range roots {
			if uxto, err := self.getUxto(root); err != nil {
				return
			} else {
				handler(uxto)
			}
		}
	}

	return
}

func (self *Exchange) getUxto(root keys.Uint256) (uxto Uxto, e error) {
	data, err := self.db.Get(rootKey(root))
	if err != nil {
		return
	}
	if err := rlp.Decode(bytes.NewReader(data), &uxto); err != nil {
		log.Error("Exchange Invalid uxto RLP", "root", common.Bytes2Hex(root[:]), "err", err)
		e = err
		return
	}

	if value, ok := self.usedFlag.Load(uxto.Nil); ok {
		uxto.flag = value.(int)
	}
	return
}

func (self *Exchange) findUxtos(pk *keys.Uint512, currency string, amount *big.Int) (uxtos []Uxto, e error) {
	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	list := UxtoList{}
	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if uxto, err := self.getUxto(root); err == nil {
			if _, ok := self.usedFlag.Load(uxto.Nil); !ok {
				uxtos = append(uxtos, uxto)
				amount.Sub(amount, uxto.Asset.Tkn.Value.ToIntRef())
			} else {
				list = append(list, uxto)
			}
		}
		if amount.Sign() <= 0 {
			break
		}
	}

	if amount.Sign() > 0 {
		if list.Len() > 0 {
			sort.Sort(list)
			for _, uxto := range list {
				uxtos = append(uxtos, uxto)
				amount.Sub(amount, uxto.Asset.Tkn.Value.ToIntRef())
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
			dout.Nil = out.Root
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

func (self *Exchange) fetchAndIndexUxto() {
	blocks, err := self.sri.GetBlocksInfo(self.lastBlockNumber+1, 100)
	if err != nil {
		log.Info("Exchange GetBlocksInfo", "error", err)
		return
	}

	log.Info("Exchange getBlocksInfo", "blocks", len(blocks), "lastBlockNumber", self.lastBlockNumber)
	if len(blocks) == 0 {
		return
	}

	outs := map[PkrKey][]light_types.Out{}
	nils := []keys.Uint256{}
	for _, block := range blocks {
		for _, out := range block.Outs {
			var pkr keys.PKr

			if out.State.OS.Out_Z != nil {
				pkr = out.State.OS.Out_Z.PKr
			}
			if out.State.OS.Out_O != nil {
				pkr = out.State.OS.Out_O.Addr
			}

			key := PkrKey{pkr, out.State.Num}

			if self.isMyPkr(pkr) {
				if _, ok := outs[key]; ok {
					outs[key] = append(outs[key], out)
				} else {
					outs[key] = []light_types.Out{out}
				}
			}

		}
		if len(block.Nils) > 0 {
			nils = append(nils, block.Nils...)
		}

	}

	uxtos := map[PkrKey][]Uxto{}
	for key, outs := range outs {
		//account := self.pkrAccounts[key.Pkr]
		douts := DecOuts(outs, &self.account.skr)
		list := []Uxto{}
		for index, out := range douts {
			dout := outs[index]
			list = append(list, Uxto{Root: dout.Root, Nil: out.Nil, TxHash: dout.State.TxHash, Num: dout.State.Num, Asset: out.Asset})
		}
		uxtos[key] = list
	}

	if len(uxtos) > 0 || len(nils) > 0 {
		if err := self.indexBlocks(uxtos, nils); err != nil {
			log.Error("indexBlocks ", "error", err)
		}
	}

}

func (self *Exchange) indexBlocks(uxtos map[PkrKey][]Uxto, nils []keys.Uint256) (err error) {
	batch := self.db.NewBatch()
	lastBlockNumber := self.lastBlockNumber
	for key, list := range uxtos {
		roots := []keys.Uint256{}
		for _, uxto := range list {
			data, err := rlp.EncodeToBytes(uxto)
			if err != nil {
				return err
			}

			// "ROOT" + root
			batch.Put(rootKey(uxto.Root), data)

			var pkKey []byte
			if uxto.Asset.Tkn != nil {
				// "PK" + pk + currency + root
				pkKey = uxtoPkKey(*self.account.pk, uxto.Asset.Tkn.Currency[:], &uxto.Root)

			} else if uxto.Asset.Tkt != nil {
				// "PK" + pk + tkt + root
				pkKey = uxtoPkKey(*self.account.pk, uxto.Asset.Tkt.Value[:], &uxto.Root)
			}
			// "PK" + pk + currency + root => 0
			batch.Put(pkKey, []byte{0})
			// "NIL" + pk + tkt + root => "PK" + pk + currency + root
			batch.Put(nilKey(uxto.Nil), pkKey)
			roots = append(roots, uxto.Root)
			log.Info("Index add", "Nil", common.Bytes2Hex(uxto.Nil[:]), "Key", common.Bytes2Hex(pkKey[:]))
		}

		data, err := rlp.EncodeToBytes(roots)
		if err != nil {
			return err
		}
		// "PKR" + prk + blockNumber => [roots]
		batch.Put(uxtoPkrKey(key.Pkr, key.Num), data)
		if lastBlockNumber < key.Num {
			lastBlockNumber = key.Num
		}
	}

	if err := batch.Write(); err != nil {
		return err
	}

	batch = self.db.NewBatch()
	for _, Nil := range nils {
		data, _ := self.db.Get(nilKey(Nil))
		log.Info("Index del", "Nil", common.Bytes2Hex(Nil[:]), "Key", common.Bytes2Hex(data[:]))
		if data != nil {
			batch.Delete(data)
			batch.Delete(nilKey(Nil))

			self.usedFlag.Delete(common.Bytes2Hex(Nil[:]))
		}
	}
	batch.Put(lastBlockNumberKey, encodeNumber(lastBlockNumber))

	if err := batch.Write(); err != nil {
		return err
	}
	self.lastBlockNumber = lastBlockNumber
	return nil
}

func (self *Exchange) isMyPkr(pkr keys.PKr) (ok bool) {
	return keys.IsMyPKr(self.account.tk, &pkr)
}

func (self *Exchange) merge() {
	prefix := uxtoPkKey(*self.account.pk, common.LeftPadBytes([]byte("SERO"), 32), nil)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	uxtos := UxtoList{}
	for iterator.Next() {
		key := iterator.Key()
		var root keys.Uint256
		copy(root[:], key[98:130])

		if uxto, err := self.getUxto(root); err == nil {
			uxtos = append(uxtos, uxto)
		}

		if uxtos.Len() > 150 {
			break
		}
	}
	if uxtos.Len() < 10 {
		return
	}

	sort.Sort(uxtos)

	uxtos = uxtos[0 : uxtos.Len()-10]

	if uxtos.Len() > 1 {
		amount := new(big.Int)
		for _, uxto := range uxtos {
			amount.Add(amount, uxto.Asset.Tkn.Value.ToIntRef())
		}
		amount.Sub(amount, new(big.Int).Mul(big.NewInt(25000), big.NewInt(1000000000)))
		var pkr common.Address
		copy(pkr[:], self.account.mainPkr[:])
		gtx, err := self.genTx(uxtos, []Reception{{Value: amount, Currency: "SERO", Pkr: pkr}}, 25000, 1000000000)
		if err != nil {
			log.Error("Exchange merge uxto", "error", err)
			return
		}
		self.commitTx(gtx)
	}
}

var (
	pkPrefix   = []byte("PK")
	pkrPrefix  = []byte("PKR")
	rootPrefix = []byte("ROOT")
	nilPrefix  = []byte("NIL")

	lastBlockNumberKey = []byte("LastBlockNumberKey")

	Prefix = []byte("Out")
)

func nilKey(nil keys.Uint256) []byte {
	return append(nilPrefix, nil[:]...)
}

func rootKey(root keys.Uint256) []byte {
	return append(rootPrefix, root[:]...)
}

// uxtoKey = pk + currency +root
func uxtoPkKey(pk keys.Uint512, currency []byte, root *keys.Uint256) []byte {
	key := append(pkPrefix, pk[:]...)
	if len(currency) > 0 {
		key = append(key, currency...)
	}
	if root != nil {
		key = append(key, root[:]...)
	}
	return key
}

func uxtoPkrKey(pkr keys.PKr, number uint64) []byte {
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

func AddJob(spec string, run RunFunc) (*cron.Cron) {
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
