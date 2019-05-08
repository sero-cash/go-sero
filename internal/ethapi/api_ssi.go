package ethapi

import (
	"context"
	"math/big"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/light/light_ref"
	"github.com/sero-cash/go-sero/zero/light/light_types"

	"github.com/sero-cash/go-sero/zero/light"

	"github.com/sero-cash/go-sero/zero/light/light_issi"

	"github.com/sero-cash/go-sero/common/hexutil"
)

type PublicSSIAPI struct {
}

func (s *PublicSSIAPI) CreateKr() (kr light_types.Kr) {
	return light.SLI_Inst.CreateKr()
}

func (s *PublicSSIAPI) GetBlocksInfo(ctx context.Context, start hexutil.Uint64, count hexutil.Uint64) ([]light_issi.Block, error) {
	return light.SSI_Inst.GetBlocksInfo(uint64(start), uint64(count))
}

func (s *PublicSSIAPI) Detail(ctx context.Context, roots []keys.Uint256, skr *keys.PKr) (douts []light_types.DOut, e error) {
	return light.SSI_Inst.Detail(roots, skr)
}

func (s *PublicSSIAPI) GenTx(ctx context.Context, param *light_issi.GenTxParam) (hash keys.Uint256, e error) {
	return light.SSI_Inst.GenTx(param)
}

func (s *PublicSSIAPI) GetTx(ctx context.Context, txhash keys.Uint256) (tx *light_types.GTx, e error) {
	return light.SSI_Inst.GetTx(txhash)
}

func (s *PublicSSIAPI) CommitTx(ctx context.Context, txhash keys.Uint256) (e error) {
	if tx, err := light.SSI_Inst.GetTx(txhash); err != nil {
		e = err
		return
	} else {
		gasPrice := big.Int(tx.GasPrice)
		gas := uint64(tx.Gas)
		if signedTx := types.NewTxWithGTx(gas, &gasPrice, &tx.Tx); signedTx == nil {
			e = errors.New("CommitTx Failed: signedTx is nil")
			return
		} else {
			log.Info("commitTx", "txhash", signedTx.Hash().String())
			if err := light_ref.Ref_inst.Se.TxPool().AddLocal(signedTx); err != nil {
				e = err
				return
			} else {
				return
			}
		}
	}
}
