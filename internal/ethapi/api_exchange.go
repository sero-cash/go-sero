package ethapi

import (
	"context"
	"fmt"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/core/types"

	"github.com/sero-cash/go-sero/core/rawdb"

	"github.com/sero-cash/go-sero/rpc"

	"math/big"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/zero/exchange"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/light/light_types"
)

type PublicExchangeAPI struct {
	b Backend
}

func (s *PublicExchangeAPI) GetPkSynced(ctx context.Context, pk *keys.Uint512) (map[string]interface{}, error) {
	currentPKBlock, err := s.b.GetPkNumber(*pk)
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
	numbers := exchangeInstance.GetUtxoNum(*pk)

	// Otherwise gather the block sync stats
	return map[string]interface{}{
		"currentPKBlock": currentPKBlock,
		"confirmedBlock": seroparam.DefaultConfirmedBlock(),
		"currentBlock":   progress.CurrentBlock,
		"highestBlock":   progress.HighestBlock,
		"utxoCount":      numbers,
	}, nil

}

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, address *keys.Uint512, index *keys.Uint256) (pkr keys.PKr, e error) {
	return s.b.GetPkr(address, index)
}

func (s *PublicExchangeAPI) GetLockedBalances(address keys.Uint512) map[string]*Big {
	result := map[string]*Big{}

	balances := s.b.GetLockedBalances(address)
	for k, v := range balances {
		result[k] = (*Big)(v)
	}
	return result
}

func (s *PublicExchangeAPI) GetMaxAvailable(address keys.Uint512, currency Smbol) (amount *Big) {
	return (*Big)(s.b.GetMaxAvailable(address, string(currency)))
}

func (s *PublicExchangeAPI) GetBalances(ctx context.Context, address keys.Uint512) map[string]*Big {
	result := map[string]*Big{}
	balances := s.b.GetBalances(address)
	for k, v := range balances {
		result[k] = (*Big)(v)
	}
	return result
}

type Big big.Int

func (b Big) MarshalJSON() ([]byte, error) {
	i := big.Int(b)
	by, err := i.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if seroparam.IsExchangeValueStr() {
		bytes := make([]byte, len(by)+2)
		copy(bytes[1:len(bytes)-1], by[:])
		bytes[0] = '"'
		bytes[len(bytes)-1] = '"'
		return bytes, nil
	} else {
		return by, err
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Big) UnmarshalJSON(input []byte) error {
	if isString(input) {
		input = input[1 : len(input)-1]
	}
	i := big.Int{}
	if e := i.UnmarshalText(input); e != nil {
		return e
	} else {
		*b = Big(i)
		return nil
	}
}

func (b *Big) ToInt() *big.Int {
	return (*big.Int)(b)
}

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

type ReceptionArgs struct {
	Addr     hexutil.Bytes
	Currency Smbol
	Value    *Big
}

type GenTxArgs struct {
	From       keys.Uint512
	RefundTo   *keys.PKr
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
		if !keys.PKrValid(args.RefundTo) {
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

func byteToPkr(addr hexutil.Bytes) keys.PKr {
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

func (args GenTxArgs) toTxParam() exchange.TxParam {
	gasPrice := args.GasPrice.ToInt()

	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}
	receptions := []exchange.Reception{}
	for _, rec := range args.Receptions {
		pkr := byteToPkr(rec.Addr)
		receptions = append(receptions, exchange.Reception{
			pkr,
			string(rec.Currency),
			(*big.Int)(rec.Value),
		})
	}
	return exchange.TxParam{args.From, args.RefundTo, receptions, args.Gas, gasPrice, args.Roots}
}

func (s *PublicExchangeAPI) GenTx(ctx context.Context, param GenTxArgs) (*light_types.GenTxParam, error) {
	if err := param.check(); err != nil {
		return nil, err
	}

	return s.b.GenTx(param.toTxParam())
}

func (s *PublicExchangeAPI) GenTxWithSign(ctx context.Context, param GenTxArgs) (*light_types.GTx, error) {
	if err := param.check(); err != nil {
		return nil, err
	}
	_, tx, e := exchange.CurrentExchange().GenTxWithSign(param.toTxParam())
	return tx, e
}

type Record struct {
	Pkr      keys.PKr
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
			records = append(records, Record{Pkr: utxo.Pkr, Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: (*Big)(utxo.Asset.Tkn.Value.ToIntRef())})
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
func (s *PublicExchangeAPI) GetRecords(ctx context.Context, begin, end uint64, address *hexutil.Bytes) (records []Record, err error) {

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
			records = append(records, Record{Pkr: utxo.Pkr, Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: (*Big)(utxo.Asset.Tkn.Value.ToIntRef())})
		}
	}

	return
}

func (s *PublicExchangeAPI) GenMergeTx(ctx context.Context, param exchange.MergeParam) (txParam *light_types.GenTxParam, e error) {
	if param.To != nil {
		if !keys.PKrValid(param.To) {
			return nil, errors.New("param.to must valid pkr or nil")
		}
	}
	if param.Currency == "" {
		return nil, errors.New("cy can not be nil")

	}
	exchangeInstance := exchange.CurrentExchange()
	if exchangeInstance == nil {
		return nil, errors.New("exchange mode no start")
	}
	return exchangeInstance.GenMergeTx(&param)
}

func (s *PublicExchangeAPI) Merge(ctx context.Context, address *keys.Uint512, cy Smbol) (map[string]interface{}, error) {
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
	count, hash, err := exchangeInstance.Merge(address, string(cy), true)
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

func validAddress(addr hexutil.Bytes) (bool, error) {
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

func (s *PublicExchangeAPI) ValidAddress(ctx context.Context, addr hexutil.Bytes) (bool, error) {
	return validAddress(addr)

}

func (s *PublicExchangeAPI) CommitTx(ctx context.Context, args *light_types.GTx) error {
	return s.b.CommitTx(args)
}

func (s *PublicExchangeAPI) ClearUsedFlag(ctx context.Context, address keys.Uint512) (count int, e error) {
	count = exchange.CurrentExchange().ClearUsedFlagForPK(&address)
	return
}

func (s *PublicExchangeAPI) ClearUsedFlagForRoot(ctx context.Context, roots []keys.Uint256) (count int, e error) {
	if len(roots) > 0 {
		for _, root := range roots {
			count += exchange.CurrentExchange().ClearUsedFlagForRoot(root)
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

	infos, err := exchange.CurrentExchange().GetBlocksInfo(start, end)
	if err != nil {
		return
	}
	for _, block := range infos {

		outs := []Record{}
		for _, utxo := range block.Outs {
			record := Record{Pkr: utxo.Pkr, Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: (*Big)(utxo.Asset.Tkn.Value.ToIntRef())}
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
func (s *PublicExchangeAPI) GetPkByPkr(ctx context.Context, pkr keys.PKr) (*keys.Uint512, error) {
	wallets := s.b.AccountManager().Wallets()
	if len(wallets) == 0 {
		return nil, nil
	}
	for _, wallet := range wallets {
		if keys.IsMyPKr(wallet.Accounts()[0].Tk.ToUint512(), &pkr) {
			return wallet.Accounts()[0].Address.ToUint512(), nil
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
