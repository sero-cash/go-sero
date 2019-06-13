package ethapi

import (
	"context"
	"github.com/sero-cash/go-sero/common"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/exchange"
	"github.com/sero-cash/go-sero/zero/light/light_types"
)

type PublicExchangeAPI struct {
	b Backend
}

func (s *PublicExchangeAPI) GetPkNumber(ctx context.Context, pk *keys.Uint512) (uint64, error) {
	return s.b.GetPkNumber(*pk)
}

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, address *keys.Uint512, index *keys.Uint256) (pkr keys.PKr, e error) {
	return s.b.GetPkr(address, index)
}

func (s *PublicExchangeAPI) GetBalances(ctx context.Context, address keys.Uint512) (balances map[string]*big.Int) {
	return s.b.GetBalances(address)
}

func (s *PublicExchangeAPI) GenTx(ctx context.Context, param exchange.TxParam) (*light_types.GenTxParam, error) {
	return s.b.GenTx(param)
}

func (s *PublicExchangeAPI) GenTxWithSign(ctx context.Context, param exchange.TxParam) (*light_types.GTx, error) {
	tx, e := s.b.GenTxWithSign(param)
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
