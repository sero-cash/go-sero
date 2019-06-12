package ethapi

import (
	"context"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/exchange"
	"github.com/sero-cash/go-sero/zero/light/light_types"
)

type PublicExchangeAPI struct {
	b Backend
}

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, address keys.Uint512, index uint64) (pkr keys.PKr, e error) {
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

func (s *PublicExchangeAPI) GetRecords(ctx context.Context, address hexutil.Bytes, begin, end uint64) (records []exchange.Utxo, err error) {
	return s.b.GetRecords(address, begin, end)
}
