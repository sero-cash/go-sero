package ethapi

import (
	"context"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/tyler-smith/go-bip39"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
)

type PublicLocalAPI struct {
}

func (s *PublicLocalAPI) DecOut(ctx context.Context, outs []txtool.Out, tk PKAddress) (douts []txtool.TDOut, e error) {
	tk_u64 := tk.ToUint512()
	tk_u96 := keys.PKr{}
	copy(tk_u96[:], tk_u64[:])
	douts = flight.DecTraceOuts(outs, &tk_u96)
	return
}

func (s *PublicLocalAPI) IsPkrValid(ctx context.Context, tk PKrAddress) error {
	return nil
}

func (s *PublicLocalAPI) IsPkValid(ctx context.Context, tk PKAddress) error {
	return nil
}

func (s *PublicLocalAPI) GenSeed(ctx context.Context) (hexutil.Bytes, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(entropy), nil
}

func (s *PublicLocalAPI) CurrencyToId(ctx context.Context, currency string) (ret keys.Uint256, e error) {
	bs := utils.CurrencyToBytes(currency)
	copy(ret[:], bs[:])
	return
}

func (s *PublicLocalAPI) IdToCurrency(ctx context.Context, hex keys.Uint256) (ret string, e error) {
	ret = utils.Uint256ToCurrency(&hex)
	return
}

func (s *PublicLocalAPI) Seed2Sk(ctx context.Context, seed hexutil.Bytes) (keys.Uint512, error) {
	if len(seed) != 32 {
		return keys.Uint512{}, errors.New("seed len must be 32")
	}
	var sd keys.Uint256
	copy(sd[:], seed[:])
	return keys.Seed2Sk(&sd), nil
}

func (s *PublicLocalAPI) Sk2Tk(ctx context.Context, sk keys.Uint512) (address.AccountAddress, error) {
	tk := keys.Sk2Tk(&sk)
	return address.BytesToAccount(tk[:]), nil
}

func (s *PublicLocalAPI) Tk2Pk(ctx context.Context, tk TKAddress) (address.AccountAddress, error) {
	pk := keys.Tk2Pk(tk.ToUint512().NewRef())
	return address.BytesToAccount(pk[:]), nil
}
func (s *PublicLocalAPI) Pk2Pkr(ctx context.Context, pk PKAddress, index *keys.Uint256) (PKrAddress, error) {
	empty := keys.Uint256{}
	if index != nil {
		if (*index) == empty {
			*index = keys.RandUint256()
		}
	}
	pkr := keys.Addr2PKr(pk.ToUint512().NewRef(), index)
	var pkrAddress PKrAddress
	copy(pkrAddress[:], pkr[:])
	return pkrAddress, nil
}

func (s *PublicLocalAPI) SignTxWithSk(param txtool.GTxParam, SK keys.Uint512) (txtool.GTx, error) {
	return flight.SignTx(&SK, &param)
}
