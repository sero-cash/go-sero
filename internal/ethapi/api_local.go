package ethapi

import (
	"context"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v0"

	"github.com/tyler-smith/go-bip39"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
)

type PublicLocalAPI struct {
}

func (s *PublicLocalAPI) DecOut(ctx context.Context, outs []txtool.Out, tk address.TKAddress) (douts []txtool.TDOut, e error) {
	tk_u64 := tk.ToTk()
	douts = flight.DecOut(&tk_u64, outs)
	return
}

func (s *PublicLocalAPI) ConfirmOutZ(ctx context.Context, key c_type.Uint256, outz stx_v0.Out_Z) (dout txtool.TDOut, e error) {
	return flight.ConfirmOutZ(&key, &outz)
}

func (s *PublicLocalAPI) ConfirmOutC(ctx context.Context, key c_type.Uint256, outc stx_v1.Out_C) (dout txtool.TDOut, e error) {
	return flight.ConfirmOutC(&key, &outc)
}

func (s *PublicLocalAPI) IsPkrValid(ctx context.Context, tk PKrAddress) error {
	return nil
}

func (s *PublicLocalAPI) IsPkValid(ctx context.Context, tk address.PKAddress) error {
	return nil
}

func (s *PublicLocalAPI) GenSeed(ctx context.Context) (hexutil.Bytes, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(entropy), nil
}

func (s *PublicLocalAPI) Seed2Mnemonic(ctx context.Context, seed hexutil.Bytes) (string, error) {
	return bip39.NewMnemonic(seed[:])
}

func (s *PublicLocalAPI) CurrencyToId(ctx context.Context, currency string) (ret c_type.Uint256, e error) {
	ret = flight.CurrencyToId(currency)
	return
}

func (s *PublicLocalAPI) IdToCurrency(ctx context.Context, hex c_type.Uint256) (ret string, e error) {
	ret = flight.IdToCurrency(&hex)
	return
}

func (s *PublicLocalAPI) IsMyPkr(ctx context.Context, tk address.TKAddress, pkr PKrAddress) (ret bool, e error) {
	ret = superzk.IsMyPKr(tk.ToTk().NewRef(), pkr.ToPKr().NewRef())
	return
}

func (s *PublicLocalAPI) Seed2Sk(ctx context.Context, seed hexutil.Bytes, v *int) (c_type.Uint512, error) {
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

func (s *PublicLocalAPI) Sk2Tk(ctx context.Context, sk c_type.Uint512) (ret address.TKAddress, err error) {
	tk, err := superzk.Sk2Tk(&sk)
	if err != nil {
		return
	}
	copy(ret[:], tk[:])
	return
}

func (s *PublicLocalAPI) Tk2Pk(ctx context.Context, tk address.TKAddress) (ret address.PKAddress, err error) {
	var pk c_type.Uint512
	pk, err = superzk.Tk2Pk(tk.ToTk().NewRef())
	copy(ret[:], pk[:])
	return
}

func (s *PublicLocalAPI) Pk2Pkr(ctx context.Context, pk address.PKAddress, index *c_type.Uint256) (PKrAddress, error) {
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

func (s *PublicLocalAPI) SignTxWithSk(param txtool.GTxParam, SK c_type.Uint512) (txtool.GTx, error) {
	return flight.SignTx(&SK, &param)
}
