package ethapi

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/core/types"

	"github.com/sero-cash/go-sero/core/rawdb"

	"github.com/sero-cash/go-sero/rpc"

	"math/big"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/zero/wallet/exchange"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/zero/light"
)

type PublicExchangeAPI struct {
	b Backend
}

func (s *PublicExchangeAPI) GetPkSynced(ctx context.Context, pk *PKAddress) (map[string]interface{}, error) {
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

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, address PKAddress, index *keys.Uint256) (pkrAdd PKrAddress, e error) {
	pk := address.ToUint512()
	pkr, err := s.b.GetPkr(&pk, index)
	if err != nil {
		e = err
		return
	}
	copy(pkrAdd[:], pkr[:])
	return
}

func (s *PublicExchangeAPI) GetLockedBalances(address PKAddress) map[string]*Big {
	result := map[string]*Big{}

	balances := s.b.GetLockedBalances(address.ToUint512())
	for k, v := range balances {
		result[k] = (*Big)(v)
	}
	return result
}

func (s *PublicExchangeAPI) GetMaxAvailable(address PKAddress, currency Smbol) (amount *Big) {
	return (*Big)(s.b.GetMaxAvailable(address.ToUint512(), string(currency)))
}

func (s *PublicExchangeAPI) GetBalances(ctx context.Context, address PKAddress) map[string]*Big {
	result := map[string]*Big{}
	balances := s.b.GetBalances(address.ToUint512())
	for k, v := range balances {
		result[k] = (*Big)(v)
	}
	return result
}

type ReceptionArgs struct {
	Addr     MixAdrress
	Currency Smbol
	Value    *Big
}

type GenTxArgs struct {
	From       PKAddress
	RefundTo   *PKrAddress
	Receptions []ReceptionArgs
	Gas        uint64
	GasPrice   *Big
	Roots      []keys.Uint256
}

func (args GenTxArgs) check() error {
	if len(args.Receptions) == 0 {
		return errors.New("have no receptions")
	}
	if args.GasPrice == nil {
		return fmt.Errorf("gasPrice not specified")
	}

	if args.RefundTo != nil {
		if !keys.PKrValid(args.RefundTo.ToPKr()) {
			return errors.New("RefundTo is not a valid pkr")
		}
	}

	for _, rec := range args.Receptions {
		_, err := validAddress(rec.Addr)
		if err != nil {
			return err
		}
		if rec.Currency.IsEmpty() {
			return errors.Errorf("%v reception currency is nil", hexutil.Encode(rec.Addr[:]))
		}
		if rec.Value == nil {
			return errors.Errorf("%v reception value is nil", hexutil.Encode(rec.Addr[:]))
		}
	}
	return nil

}

func MixAdrressToPkr(addr MixAdrress) keys.PKr {
	pkr := keys.PKr{}
	if len(addr) == 64 {
		pk := keys.Uint512{}
		copy(pk[:], addr[:])
		pkr = keys.Addr2PKr(&pk, nil)
	} else {
		copy(pkr[:], addr[:])
	}
	return pkr
}

func (args GenTxArgs) toTxParam() prepare.PreTxParam {
	gasPrice := args.GasPrice.ToInt()

	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}
	receptions := []prepare.Reception{}
	for _, rec := range args.Receptions {
		pkr := MixAdrressToPkr(rec.Addr)
		var currency keys.Uint256
		bytes := common.LeftPadBytes([]byte(string(rec.Currency)), 32)
		copy(currency[:], bytes)
		receptions = append(receptions, prepare.Reception{
			pkr,
			assets.Asset{Tkn: &assets.Token{
				Currency: currency,
				Value:    utils.U256(*rec.Value.ToInt())},
			},
		})
	}
	var refundPkr *keys.PKr
	if args.RefundTo != nil {
		refundPkr = args.RefundTo.ToPKr()
	}
	return prepare.PreTxParam{
		args.From.ToUint512(),
		refundPkr,
		receptions,
		prepare.Cmds{},
		assets.Token{
			utils.CurrencyToUint256("SERO"),
			utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(args.Gas)), args.GasPrice.ToInt())),
		},
		gasPrice,
		args.Roots,
	}
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
	_, tx, e := exchange.CurrentExchange().GenTxWithSign(param.toTxParam())
	return tx, e
}

func pkrToPKrAddress(pkr keys.PKr) PKrAddress {
	pkrAddress := PKrAddress{}
	copy(pkrAddress[:], pkr[:])
	return pkrAddress
}

type Record struct {
	Pkr      PKrAddress
	Root     keys.Uint256
	TxHash   keys.Uint256
	Nil      keys.Uint256
	Num      uint64
	Currency string
	Value    *Big
}

func (s *PublicExchangeAPI) GetTx(ctx context.Context, txHash keys.Uint256) (map[string]interface{}, error) {

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
		outs = append(outs, r)
	}
	fields["Outs"] = outs

	ins := []keys.Uint256{}
	for _, in := range tx.Stxt().Desc_O.Ins {
		ins = append(ins, in.Root)
	}
	for _, in := range tx.Stxt().Desc_Z.Ins {
		if root := exchange.CurrentExchange().GetRootByNil(in.Trace); root != nil {
			ins = append(ins, *root)
		}
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
			var pk keys.Uint512
			copy(pk[:], addr[:])
			utxos, err = s.b.GetRecordsByPk(&pk, begin, end)
		} else if len(addr) == 96 {
			var pkr keys.PKr
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
	From     PKAddress
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
		if !keys.PKrValid(args.To.ToPKr()) {
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

func (s *PublicExchangeAPI) Merge(ctx context.Context, address *PKAddress, cy Smbol) (map[string]interface{}, error) {
	if address == nil {
		return nil, errors.New("pk can not be nil")
	}
	if cy == "" {
		return nil, errors.New("cy can not be nil")

	}
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return nil, errors.New("exchange mode no start")
	}
	count, hash, err := exchangeInstance.Merge(address.ToUint512().NewRef(), string(cy), true)
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
		pk := keys.Uint512{}
		copy(pk[:], addr[:])
		if !keys.IsPKValid(&pk) {
			return false, errors.Errorf("invalid pk %v", hexutil.Encode(addr[:]))
		}
	}
	if len(addr) == 96 {
		pkr := keys.PKr{}
		copy(pkr[:], addr[:])
		if !keys.PKrValid(&pkr) {
			return false, errors.Errorf("invalid  pkr %v", hexutil.Encode(addr[:]))
		}
	}
	return true, nil
}

func (s *PublicExchangeAPI) ValidAddress(ctx context.Context, addr MixBase58Adrress) (bool, error) {
	if len(addr) != 64 && len(addr) != 96 {
		return false, errors.Errorf("invalid addr %v", base58.Encode(addr[:]))
	}

	if len(addr) == 64 {
		pk := keys.Uint512{}
		copy(pk[:], addr[:])
		if !keys.IsPKValid(&pk) {
			return false, errors.Errorf("invalid pk %v", base58.Encode(addr[:]))
		}
	}
	if len(addr) == 96 {
		pkr := keys.PKr{}
		copy(pkr[:], addr[:])
		if !keys.PKrValid(&pkr) {
			return false, errors.Errorf("invalid  pkr %v", base58.Encode(addr[:]))
		}
	}
	return true, nil

}

func (s *PublicExchangeAPI) CommitTx(ctx context.Context, args *txtool.GTx) error {
	return s.b.CommitTx(args)
}

func (s *PublicExchangeAPI) ClearUsedFlag(ctx context.Context, pkaddress PKAddress) (count int, e error) {
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return 0, errors.New("exchange mode no start")
	}
	address := pkaddress.ToUint512()
	count = exchangeInstance.ClearUsedFlagForPK(&address)
	return
}

func (s *PublicExchangeAPI) ClearUsedFlagForRoot(ctx context.Context, roots []keys.Uint256) (count int, e error) {
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
	BlockHash   keys.Uint256
	Ins         []keys.Uint256
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
func (s *PublicExchangeAPI) GetPkByPkr(ctx context.Context, pkr PKrAddress) (*address.AccountAddress, error) {
	wallets := s.b.AccountManager().Wallets()
	if len(wallets) == 0 {
		return nil, nil
	}
	for _, wallet := range wallets {
		if keys.IsMyPKr(wallet.Accounts()[0].Tk.ToUint512(), pkr.ToPKr()) {
			return &wallet.Accounts()[0].Address, nil
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
		"Timestamp":   block.Header().Time.Uint64(),
		"TxHashes":    transactions,
	}
	return fields, nil
}


func (s *PublicExchangeAPI) Seed2Sk(ctx context.Context,seed hexutil.Bytes) (keys.Uint512,error) {
	if len(seed) !=32{
      return keys.Uint512{},errors.New("seed len must be 32")
	}
	var sd keys.Uint256
	copy(sd[:],seed[:])
	return keys.Seed2Sk(&sd),nil
}

func (s *PublicExchangeAPI) SignTxWithSk(param light_types.GenTxParam,SK keys.Uint512) (light_types.GTx, error) {
	return light.SignTx(&SK,&param)
}

func (s  *PublicExchangeAPI) Sk2Tk(ctx context.Context,sk keys.Uint512) (address.AccountAddress,error) {
	tk:=keys.Sk2Tk(&sk)
	return address.BytesToAccount(tk[:]),nil
}

func (s  *PublicExchangeAPI) Tk2Pk(ctx context.Context,tk TKAddress) (address.AccountAddress,error) {
	 pk:= keys.Tk2Pk(tk.ToUint512().NewRef())
	return address.BytesToAccount(pk[:]),nil
}
