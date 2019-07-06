package ethapi

import (
	"context"
	"math/big"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/wallet/exchange"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/stake"
)

type PublicShareApI struct {
	b         Backend
	nonceLock *AddrLocker
}

func NewPublicShareApI(b Backend, nonceLock *AddrLocker) *PublicShareApI {
	return &PublicShareApI{
		nonceLock: nonceLock,
		b:         b,
	}
}

type BuyShareTxArg struct {
	From     address.AccountAddress `json:"from"`
	Vote     *common.Address        `json:"vote"`
	Pool     *common.Hash           `json:"pool"`
	Gas      *hexutil.Uint64        `json:"gas"`
	GasPrice *hexutil.Big           `json:"gasPrice"`
	Value    *hexutil.Big           `json:"value"`
}

func (args *BuyShareTxArg) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 25000
	}

	if args.Vote == nil {
		return errors.New("vote address cannot be nil")
	}

	if args.Pool != nil {
		state, _, err := b.StateAndHeaderByNumber(ctx, -1)
		if err != nil {
			return err
		}

		pool := stake.NewStakeState(state).GetStakePool(*args.Pool)

		if pool == nil {
			return errors.New("stake pool not exists")
		}

		if pool.Closed {
			return errors.New("stake pool has closed")
		}
	}

	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	} else {
		if args.GasPrice.ToInt().Sign() == 0 {
			return errors.New(`gasPrice can not be zero`)
		}
	}

	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	return nil
}

func (args *BuyShareTxArg) toPreTxParam() txtool.PreTxParam {
	preTx := txtool.PreTxParam{}
	preTx.From = *args.From.ToUint512()
	preTx.RefundTo = keys.Addr2PKr(args.From.ToUint512(), &shareFromRand).NewRef()
	preTx.Gas = uint64(*args.Gas)
	preTx.GasPrice = (*big.Int)(args.GasPrice)
	preTx.Cmds = txtool.Cmds{}

	buyShareCmd := stx.BuyShareCmd{}
	buyShareCmd.Value = utils.U256(*args.Value.ToInt())
	buyShareCmd.Vote = common.AddrToPKr(*args.Vote)
	if args.Pool != nil {
		var pool keys.Uint256
		copy(pool[:], args.Pool[:])
		buyShareCmd.Pool = &pool
	}
	preTx.Cmds.BuyShare = &buyShareCmd
	return preTx

}

var shareFromRand = keys.RandUint256()

func (s *PublicShareApI) BuyShare(ctx context.Context, args BuyShareTxArg) (common.Hash, error) {
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	preTx := args.toPreTxParam()
	pretx, gtx, err := exchange.CurrentExchange().GenTxWithSign(preTx)
	if err != nil {
		return common.Hash{}, err
	}
	err = s.b.CommitTx(gtx)
	if err != nil {
		exchange.CurrentExchange().ClearTxParam(pretx)
		return common.Hash{}, err
	}
	return common.BytesToHash(gtx.Hash[:]), nil
}

type RegistStakePoolTxArg struct {
	From     address.AccountAddress `json:"from"`
	Vote     *common.Address        `json:"vote"`
	Gas      *hexutil.Uint64        `json:"gas"`
	GasPrice *hexutil.Big           `json:"gasPrice"`
	Value    *hexutil.Big           `json:"value"`
	Fee      *hexutil.Uint          `json:"fee"`
}

func (args *RegistStakePoolTxArg) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 25000
	}

	if args.Vote == nil {
		return errors.New("vote address cannot be nil")
	}
	if args.Fee == nil {
		return errors.New("pool fee cannot be nil")
	}

	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	} else {
		if args.GasPrice.ToInt().Sign() == 0 {
			return errors.New(`gasPrice can not be zero`)
		}
	}
	if uint32(*args.Fee) > 10000 {
		return errors.New("fee rate can not large then %100")
	}

	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	return nil
}

func (args *RegistStakePoolTxArg) toPreTxParam() txtool.PreTxParam {
	preTx := txtool.PreTxParam{}
	preTx.From = *args.From.ToUint512()
	preTx.RefundTo = keys.Addr2PKr(args.From.ToUint512(), &shareFromRand).NewRef()
	preTx.Gas = uint64(*args.Gas)
	preTx.GasPrice = (*big.Int)(args.GasPrice)
	preTx.Cmds = txtool.Cmds{}
	registPoolCmd := stx.RegistPoolCmd{}
	registPoolCmd.Value = utils.U256(*args.Value.ToInt())
	registPoolCmd.Vote = common.AddrToPKr(*args.Vote)
	registPoolCmd.FeeRate = uint32(*args.Fee) * 2 / 3
	preTx.Cmds.RegistPool = &registPoolCmd
	return preTx

}

func (s *PublicShareApI) RegistStakePool(ctx context.Context, args RegistStakePoolTxArg) (common.Hash, error) {
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	preTx := args.toPreTxParam()
	pretx, gtx, err := exchange.CurrentExchange().GenTxWithSign(preTx)
	if err != nil {
		return common.Hash{}, err
	}
	err = s.b.CommitTx(gtx)
	if err != nil {
		exchange.CurrentExchange().ClearTxParam(pretx)
		return common.Hash{}, err
	}
	return common.BytesToHash(gtx.Hash[:]), nil
}

func (s *PublicShareApI) PoolState(ctx context.Context, pool common.Hash) (map[string]interface{}, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	poolState := stake.NewStakeState(state).GetStakePool(pool)

	if poolState == nil {
		return nil, errors.New("stake pool not exists")
	}

	return nil, nil
}
