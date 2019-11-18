package ethapi

import (
	"context"

	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/zero/txtool/flight"

	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/core/types"

	"github.com/sero-cash/go-sero/core/rawdb"

	"github.com/sero-cash/go-sero/rpc"

	"math/big"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/zero/wallet/exchange"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/common/hexutil"
)

type PublicExchangeAPI struct {
	b Backend
}

func (s *PublicExchangeAPI) GetPkSynced(ctx context.Context, pk *address.PKAddress) (map[string]interface{}, error) {
	if pk == nil {
		return nil, errors.New("pk can not be nil")
	}
	currentPKBlock, err := s.b.GetPkNumber(pk.ToUint512())
	if err != nil {
		return nil, err
	}
	progress := s.b.Downloader().Progress()
	if progress.CurrentBlock >= progress.HighestBlock {
		progress.HighestBlock = progress.CurrentBlock
	}
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return nil, errors.New("exchange mode no start")
	}

	numbers := exchangeInstance.GetUtxoNum(pk.ToUint512())

	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"currentPKBlock": currentPKBlock,
		"confirmedBlock": seroparam.DefaultConfirmedBlock(),
		"currentBlock":   progress.CurrentBlock,
		"highestBlock":   progress.HighestBlock,
		"utxoCount":      numbers,
	}, nil

}

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, pk address.PKAddress, index *c_type.Uint256) (pkrAdd PKrAddress, e error) {

	pkr, err := s.b.GetPkr(pk.ToUint512().NewRef(), index)
	if err != nil {
		e = err
		return
	}
	copy(pkrAdd[:], pkr[:])
	return
}

func (s *PublicExchangeAPI) GetLockedBalances(pk address.PKAddress) map[string]*Big {
	result := map[string]*Big{}
	balances := s.b.GetLockedBalances(pk.ToUint512())
	for k, v := range balances {
		result[k] = (*Big)(v)
	}
	return result
}

func (s *PublicExchangeAPI) GetMaxAvailable(pk address.PKAddress, currency Smbol) (amount *Big) {
	return (*Big)(s.b.GetMaxAvailable(pk.ToUint512(), string(currency)))
}

func (s *PublicExchangeAPI) GetBalances(ctx context.Context, pk address.PKAddress) map[string]interface{} {
	result := map[string]*Big{}
	ret := make(map[string]interface{})
	balances, tickets := s.b.GetBalances(pk.ToUint512())
	for k, v := range balances {
		result[k] = (*Big)(v)
		ret[k] = (*Big)(v)
	}
	ret["tkn"] = result
	ret["tkt"] = tickets
	return ret
}

type ReceptionArgs struct {
	Addr     MixAdrress
	Currency Smbol
	Value    *Big
}

func MixAdrressToPkr(addr MixAdrress) c_type.PKr {
	pkr := c_type.PKr{}
	if len(addr) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], addr[:])
		pkr = superzk.Pk2PKr(&pk, nil)
	} else {
		copy(pkr[:], addr[:])
	}
	return pkr
}

func (s *PublicExchangeAPI) GenTx(ctx context.Context, param GenTxArgs) (*txtool.GTxParam, error) {
	if err := param.check(); err != nil {
		return nil, err
	}

	return s.b.GenTx(param.toTxParam())
}

func (s *PublicExchangeAPI) GenTxWithSign(ctx context.Context, param GenTxArgs) (*txtool.GTx, error) {
	if err := param.check(); err != nil {
		return nil, err
	}
	txParam, tx, e := exchange.CurrentExchange().GenTxWithSign(param.toTxParam())
	if tx != nil {
		for _, in := range txParam.Ins {
			tx.Roots = append(tx.Roots, in.Out.Root)
		}
	}
	return tx, e
}

func pkrToPKrAddress(pkr c_type.PKr) PKrAddress {
	pkrAddress := PKrAddress{}
	copy(pkrAddress[:], pkr[:])
	return pkrAddress
}

type Record struct {
	Pkr      PKrAddress
	Root     c_type.Uint256
	TxHash   c_type.Uint256
	Nil      c_type.Uint256
	Num      uint64
	Currency string
	Value    *Big
}

func (s *PublicExchangeAPI) GetTx(ctx context.Context, txHash c_type.Uint256) (map[string]interface{}, error) {

	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), common.BytesToHash(txHash[:]))
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}

	utxos, err := s.b.GetRecordsByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	receipt := receipts[index]
	gasUsed := receipt.GasUsed
	fee := new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(gasUsed)))
	fields := map[string]interface{}{
		"BlockNumber": blockNumber,
		"BlockHash":   blockHash,
		"TxHash":      common.BytesToHash(txHash[:]),
		"Fee":         (*utils.U256)(fee),
		"GasPrice":    (*utils.U256)(tx.GasPrice()),
		"GasUsed":     gasUsed,
	}

	block, _ := s.b.BlockByNumber(ctx, rpc.BlockNumber(blockNumber))
	if block != nil {
		fields["Timestamp"] = block.Header().Time.Uint64()
	} else {
		fields["Timestamp"] = 0
	}
	records := []Record{}
	for _, utxo := range utxos {
		if utxo.Asset.Tkn != nil {
			records = append(records, Record{Pkr: pkrToPKrAddress(utxo.Pkr), Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: (*Big)(utxo.Asset.Tkn.Value.ToIntRef())})
		}
	}
	outs := []map[string]interface{}{}
	for _, record := range records {
		r := map[string]interface{}{}
		r["Pkr"] = record.Pkr
		r["Currency"] = record.Currency
		r["Value"] = record.Value
		r["Root"] = record.Root
		outs = append(outs, r)
	}
	fields["Outs"] = outs

	ins := []c_type.Uint256{}
	if tx.Stxt().Tx0() != nil {
		for _, in := range tx.Stxt().Tx0().Desc_O.Ins {
			ins = append(ins, in.Root)
		}
		for _, in := range tx.Stxt().Tx0().Desc_Z.Ins {
			if root := exchange.CurrentExchange().GetRootByNil(in.Trace); root != nil {
				ins = append(ins, *root)
			}
		}
	}
	for _, in := range tx.Stxt().Tx1.Ins_C {
		if root := exchange.CurrentExchange().GetRootByNil(in.Nil); root != nil {
			ins = append(ins, *root)
		}
	}
	for _, in := range tx.Stxt().Tx1.Ins_P {
		ins = append(ins, in.Root)
	}
	for _, in := range tx.Stxt().Tx1.Ins_P0 {
		ins = append(ins, in.Root)
	}
	fields["Ins"] = ins

	return fields, nil

}
func (s *PublicExchangeAPI) GetRecords(ctx context.Context, begin, end uint64, address *MixAdrress) (records []Record, err error) {

	var utxos []exchange.Utxo
	if address == nil || len(*address) == 0 {
		utxos, err = s.b.GetRecordsByPk(nil, begin, end)
	} else {
		addr := *address
		if len(addr) == 64 {
			var pk c_type.Uint512
			copy(pk[:], addr[:])

			utxos, err = s.b.GetRecordsByPk(&pk, begin, end)
		} else if len(addr) == 96 {
			var pkr c_type.PKr
			copy(pkr[:], addr[:])
			utxos, err = s.b.GetRecordsByPkr(pkr, begin, end)
		} else {
			return records, errors.New("address is error")
		}
	}

	if err != nil {
		return
	}

	for _, utxo := range utxos {
		if utxo.Asset.Tkn != nil {
			records = append(records, Record{Pkr: pkrToPKrAddress(utxo.Pkr), Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: (*Big)(utxo.Asset.Tkn.Value.ToIntRef())})
		}
	}

	return
}

type MergeArgs struct {
	From     address.PKAddress
	To       *PKrAddress
	Currency Smbol
	Zcount   uint64
	Left     uint64
}

func (args MergeArgs) ToMergParam() *exchange.MergeParam {
	mergeParam := exchange.MergeParam{}
	if args.To != nil {
		mergeParam.To = args.To.ToPKr()
	}
	mergeParam.From = args.From.ToUint512()
	mergeParam.Currency = string(args.Currency)
	mergeParam.Zcount = args.Zcount
	mergeParam.Left = args.Left
	return &mergeParam

}
func (args MergeArgs) Check() error {
	if args.Currency == "" {
		return errors.New("cy can not be nil")
	}
	if args.To != nil {
		if !superzk.IsPKrValid(args.To.ToPKr()) {
			return errors.New("To is not a valid pkr")
		}
	}
	return nil
}

func (s *PublicExchangeAPI) GenMergeTx(ctx context.Context, args MergeArgs) (txParam *txtool.GTxParam, e error) {
	if e = args.Check(); e != nil {
		return
	}

	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return nil, errors.New("exchange mode no start")
	}
	return exchangeInstance.GenMergeTx(args.ToMergParam())
}

func (s *PublicExchangeAPI) Merge(ctx context.Context, pk *address.PKAddress, cy Smbol) (map[string]interface{}, error) {
	if pk == nil {
		return nil, errors.New("pk can not be nil")
	}
	if cy == "" {
		return nil, errors.New("cy can not be nil")

	}
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return nil, errors.New("exchange mode no start")
	}
	count, hash, err := exchangeInstance.Merge(pk.ToUint512().NewRef(), string(cy), true)
	log.Info("merge query utxo", "cy=", cy, "count=", count)
	if err != nil {
		return nil, err
	}
	log.Info("merge query utxo", "count=", count)
	txhash := common.Hash{}
	copy(txhash[:], hash[:])
	return map[string]interface{}{
		"utxoCount": count,
		"txhash":    txhash,
	}, nil
}

func validAddress(addr MixAdrress) (bool, error) {
	if len(addr) != 64 && len(addr) != 96 {
		return false, errors.Errorf("invalid addr %v", hexutil.Encode(addr[:]))
	}

	if len(addr) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], addr[:])
		if !superzk.IsPKValid(&pk) {
			return false, errors.Errorf("invalid pk %v", hexutil.Encode(addr[:]))
		}
	}
	if len(addr) == 96 {
		pkr := c_type.PKr{}
		copy(pkr[:], addr[:])
		if !superzk.IsPKrValid(&pkr) {
			return false, errors.Errorf("invalid  pkr %v", hexutil.Encode(addr[:]))
		}
	}
	return true, nil
}

func (s *PublicExchangeAPI) ValidAddress(ctx context.Context, addr address.MixBase58Adrress) (bool, error) {
	if len(addr) != 64 && len(addr) != 96 {
		return false, errors.Errorf("invalid addr %v", base58.Encode(addr[:]))
	}

	if len(addr) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], addr[:])
		if !superzk.IsPKValid(&pk) {
			return false, errors.Errorf("invalid pk %v", base58.Encode(addr[:]))
		}
	}
	if len(addr) == 96 {
		pkr := c_type.PKr{}
		copy(pkr[:], addr[:])
		if !superzk.IsPKrValid(&pkr) {
			return false, errors.Errorf("invalid  pkr %v", base58.Encode(addr[:]))
		}
	}
	return true, nil

}

func (s *PublicExchangeAPI) CommitTx(ctx context.Context, args *txtool.GTx) error {
	return s.b.CommitTx(args)
}

func (s *PublicExchangeAPI) ClearUsedFlag(ctx context.Context, pk address.PKAddress) (count int, e error) {
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return 0, errors.New("exchange mode no start")
	}
	count = exchangeInstance.ClearUsedFlagForPK(pk.ToUint512().NewRef())
	return
}

func (s *PublicExchangeAPI) ClearUsedFlagForRoot(ctx context.Context, roots []c_type.Uint256) (count int, e error) {
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return 0, errors.New("exchange mode no start")
	}
	if len(roots) > 0 {
		for _, root := range roots {
			count += exchangeInstance.ClearUsedFlagForRoot(root)
		}
	}
	return
}

type Block struct {
	BlockNumber uint64
	BlockHash   c_type.Uint256
	Ins         []c_type.Uint256
	Outs        []Record
	TxHashes    []common.Hash
	Timestamp   uint64
}

func (s *PublicExchangeAPI) GetBlocksInfo(ctx context.Context, start, end uint64) (blocks []Block, err error) {
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return nil, errors.New("exchange mode no start")
	}
	infos, err := exchangeInstance.GetBlocksInfo(start, end)
	if err != nil {
		return
	}
	for _, block := range infos {

		outs := []Record{}
		for _, utxo := range block.Outs {
			record := Record{Pkr: pkrToPKrAddress(utxo.Pkr), Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: (*Big)(utxo.Asset.Tkn.Value.ToIntRef())}
			outs = append(outs, record)
		}

		b, _ := s.b.BlockByNumber(ctx, rpc.BlockNumber(block.Num))
		if b == nil {
			err = errors.New("getBlockByNumber is nil")
			return
		}

		formatTx := func(tx *types.Transaction) (common.Hash, error) {
			return tx.Hash(), nil
		}
		txs := b.Transactions()
		transactions := make([]common.Hash, len(txs))
		var err error
		for i, tx := range txs {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}

		blocks = append(blocks, Block{BlockNumber: block.Num, BlockHash: block.Hash, TxHashes: transactions, Ins: block.Ins, Outs: outs, Timestamp: b.Header().Time.Uint64()})

	}
	return
}
func (s *PublicExchangeAPI) GetPkByPkr(ctx context.Context, pkr PKrAddress) (*address.PKAddress, error) {
	wallets := s.b.AccountManager().Wallets()
	if len(wallets) == 0 {
		return nil, nil
	}
	for _, wallet := range wallets {
		if superzk.IsMyPKr(wallet.Accounts()[0].Tk.ToTk().NewRef(), pkr.ToPKr()) {
			pkAddr := wallet.Accounts()[0].Address
			return &pkAddr, nil
		}
	}
	return nil, nil
}

func (s *PublicExchangeAPI) GetBlockByNumber(ctx context.Context, blockNum *int64) (map[string]interface{}, error) {
	blockNr := rpc.LatestBlockNumber
	if blockNum != nil {
		blockNr = rpc.BlockNumber(*blockNum)
	}
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}

	formatTx := func(tx *types.Transaction) (common.Hash, error) {
		return tx.Hash(), nil
	}
	txs := block.Transactions()
	transactions := make([]common.Hash, len(txs))
	for i, tx := range txs {
		if transactions[i], err = formatTx(tx); err != nil {
			return nil, err
		}
	}

	fields := map[string]interface{}{
		"BlockNumber": block.Header().Number.Uint64(),
		"BlockHash":   block.Hash(),
		"ParentHash":  block.ParentHash(),
		"Timestamp":   block.Header().Time.Uint64(),
		"TxHashes":    transactions,
	}
	return fields, nil
}

func (s *PublicExchangeAPI) Seed2Sk(ctx context.Context, seed hexutil.Bytes, v *int) (c_type.Uint512, error) {
	if len(seed) != 32 {
		return c_type.Uint512{}, errors.New("seed len must be 32")
	}

	version := 1
	if v != nil {
		version = *v
	}
	if version > 2 || version < 1 {
		return c_type.Uint512{}, errors.New("version must 1 or 2")
	}

	var sd c_type.Uint256
	copy(sd[:], seed[:])
	return superzk.Seed2Sk(&sd, version), nil
}

func (s *PublicExchangeAPI) SignTxWithSk(param txtool.GTxParam, SK c_type.Uint512) (txtool.GTx, error) {
	return flight.SignTx(&SK, &param)
}

func (s *PublicExchangeAPI) Sk2Tk(ctx context.Context, sk c_type.Uint512) (ret address.TKAddress, err error) {
	tk, err := superzk.Sk2Tk(&sk)
	if err != nil {
		return
	}
	copy(ret[:], tk[:])
	return
}

func (s *PublicExchangeAPI) Tk2Pk(ctx context.Context, tk address.TKAddress) (ret address.PKAddress, err error) {
	var pk c_type.Uint512
	pk, err = superzk.Tk2Pk(tk.ToTk().NewRef())
	copy(ret[:], pk[:])
	return
}
func (s *PublicExchangeAPI) Pk2Pkr(ctx context.Context, pk address.PKAddress, index *c_type.Uint256) (PKrAddress, error) {
	empty := c_type.Uint256{}
	if index != nil {
		if (*index) == empty {
			*index = c_type.RandUint256()
		}
	}
	pkr := superzk.Pk2PKr(pk.ToUint512().NewRef(), index)
	var pkrAddress PKrAddress
	copy(pkrAddress[:], pkr[:])
	return pkrAddress, nil
}

func (s *PublicExchangeAPI) FindRoots(pk address.PKAddress, cy Smbol, amount Big) (map[string]interface{}, error) {
	utxos, remaining := exchange.CurrentExchange().FindRoots(pk.ToUint512().NewRef(), string(cy), amount.ToInt())
	result := map[string]interface{}{}
	result["utxos"] = utxos
	result["remaining"] = Big(remaining)
	return result, nil
}

func (s *PublicExchangeAPI) GetOut(ctx context.Context, root c_type.Uint256) *prepare.Utxo {
	return exchange.CurrentExchange().GetRoot(&root)
}

func (s *PublicExchangeAPI) SetBalancePkr(ctx context.Context, pkr PKrAddress) error {
	account, err := s.b.AccountManager().FindAccountByPkr(*pkr.ToPKr())
	if err != nil {
		return err
	}
	return exchange.CurrentExchange().SetBalancePkr(account.Address.ToUint512().NewRef(), *pkr.ToPKr())
}

func (s *PublicExchangeAPI) IgnorePkrUtxos(ctx context.Context, pkr PKrAddress, ignore bool) (utxos []exchange.Utxo, e error) {
	return exchange.CurrentExchange().IgnorePkrUtxos(*pkr.ToPKr(), ignore)
}
