package ethapi

import (
	"context"
	"fmt"
	"math/big"

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

func (s *PublicExchangeAPI) GetPkSynced(ctx context.Context, pk *keys.Uint512) (map[string]uint64, error) {
	currentPKBlock, err := s.b.GetPkNumber(*pk)
	if err != nil {
		return nil, err
	}
	progress := s.b.Downloader().Progress()
	if progress.CurrentBlock >= progress.HighestBlock {
		progress.HighestBlock = progress.CurrentBlock
	}

	// Otherwise gather the block sync stats
	return map[string]uint64{
		"currentPKBlock": currentPKBlock,
		"confirmedBlock": seroparam.DefaultConfirmedBlock(),
		"currentBlock":   progress.CurrentBlock,
		"highestBlock":   progress.HighestBlock,
	}, nil

}

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, address *keys.Uint512, index *keys.Uint256) (pkr keys.PKr, e error) {
	return s.b.GetPkr(address, index)
}

func (s *PublicExchangeAPI) GetBalances(ctx context.Context, address keys.Uint512) (balances map[string]*big.Int) {
	return s.b.GetBalances(address)
}

type Big big.Int

func (b Big) MarshalJSON() ([]byte, error) {
	i := big.Int(b)
	return i.MarshalText()
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
	Addr     keys.PKr
	Currency string
	Value    *Big
}

type GenTxArgs struct {
	From       keys.Uint512
	Receptions []ReceptionArgs
	Gas        uint64
	GasPrice   *Big
	Roots      []keys.Uint256
}

func (args GenTxArgs) toTxParam() exchange.TxParam {
	gasPrice := args.GasPrice.ToInt()

	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}
	receptions := []exchange.Reception{}
	for _, rec := range args.Receptions {
		receptions = append(receptions, exchange.Reception{
			rec.Addr,
			rec.Currency,
			(*big.Int)(rec.Value),
		})
	}
	return exchange.TxParam{args.From, receptions, args.Gas, gasPrice, args.Roots}
}

func (s *PublicExchangeAPI) GenTx(ctx context.Context, param GenTxArgs) (*light_types.GenTxParam, error) {
	if param.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	return s.b.GenTx(param.toTxParam())
}

func (s *PublicExchangeAPI) GenTxWithSign(ctx context.Context, param GenTxArgs) (*light_types.GTx, error) {
	if param.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	tx, e := s.b.GenTxWithSign(param.toTxParam())
	return tx, e
}

type Record struct {
	Pkr      keys.PKr
	Root     keys.Uint256
	TxHash   keys.Uint256
	Nil      keys.Uint256
	Num      uint64
	Currency string
	Value    *big.Int
}

func (s *PublicExchangeAPI) GetRecords(ctx context.Context, address hexutil.Bytes, begin, end uint64) (records []Record, err error) {

	utxos, err := s.b.GetRecords(address, begin, end)
	if err != nil {
		return
	}
	for _, utxo := range utxos {
		if utxo.Asset.Tkn != nil {
			records = append(records, Record{Pkr: utxo.Pkr, Root: utxo.Root, TxHash: utxo.TxHash, Nil: utxo.Nil, Num: utxo.Num, Currency: common.BytesToString(utxo.Asset.Tkn.Currency[:]), Value: utxo.Asset.Tkn.Value.ToIntRef()})
		}
	}
	return
}

func (s *PublicExchangeAPI) CommitTx(ctx context.Context, args *light_types.GTx) error {
	return s.b.CommitTx(args)
}
