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

func (s *PublicExchangeAPI) GetPkr(ctx context.Context, address keys.Uint512, index hexutil.Uint64) (pkr keys.PKr, e error) {
	return s.b.GetPkr(address, uint64(index))
}

func (s *PublicExchangeAPI) GetBalances(ctx context.Context, address hexutil.Bytes) (balances map[string]*big.Int) {
	var pkr keys.PKr
	copy(pkr[:], address[:])
	return s.b.GetBalances(pkr)
}

func (s *PublicExchangeAPI) GenTx(ctx context.Context, param exchange.TxParam) (*light_types.GenTxParam, error) {
	return s.b.GenTx(param)
}

func (s *PublicExchangeAPI) GenTxWithSign(ctx context.Context, param exchange.TxParam) (*light_types.GTx, error) {
	tx, e := s.b.GenTxWithSign(param)
	return tx, e
}

func (s *PublicExchangeAPI) GetRecords(ctx context.Context, pkr keys.PKr, begin, end hexutil.Uint64) (records []exchange.Utxo, err error) {
	return s.b.GetRecords(pkr, uint64(begin), uint64(end))
}
