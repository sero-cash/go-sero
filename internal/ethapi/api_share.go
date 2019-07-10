package ethapi

import (
	"context"
	"math/big"

	"github.com/sero-cash/go-sero/accounts"

	"github.com/sero-cash/go-sero/zero/wallet/stakeservice"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-sero/zero/txtool/prepare"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/wallet/exchange"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/stake"
)

type PublicStakeApI struct {
	b         Backend
	nonceLock *AddrLocker
}

func NewPublicStakeApI(b Backend, nonceLock *AddrLocker) *PublicStakeApI {
	return &PublicStakeApI{
		nonceLock: nonceLock,
		b:         b,
	}
}

type BuyShareTxArg struct {
	From     address.AccountAddress `json:"from"`
	Vote     *common.Address        `json:"vote"`
	Pool     *hexutil.Bytes         `json:"pool"`
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

		pool := stake.NewStakeState(state).GetStakePool(common.BytesToHash((*args.Pool)[:]))

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

func (args *BuyShareTxArg) toPreTxParam() prepare.PreTxParam {
	preTx := prepare.PreTxParam{}
	preTx.From = *args.From.ToUint512()
	preTx.RefundTo = keys.Addr2PKr(args.From.ToUint512(), nil).NewRef()
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(*args.Gas)), args.GasPrice.ToInt())),
	}
	preTx.RefundTo = keys.Addr2PKr(args.From.ToUint512(), fromRand()).NewRef()
	preTx.GasPrice = (*big.Int)(args.GasPrice)
	preTx.Cmds = prepare.Cmds{}

	buyShareCmd := stx.BuyShareCmd{}
	buyShareCmd.Value = utils.U256(*args.Value.ToInt())
	buyShareCmd.Vote = common.AddrToPKr(*args.Vote)
	if args.Pool != nil {
		var pool keys.Uint256
		copy(pool[:], (*args.Pool)[:])
		buyShareCmd.Pool = &pool
	}
	preTx.Cmds.BuyShare = &buyShareCmd
	return preTx

}

const fromRandHex = "0x6e7d302d0c5ac4330dc5b006d9ad0a3bc88bcd45db01b030472fb00cfe3aa52"

func fromRand() *keys.Uint256 {
	var rand keys.Uint256
	out, _ := hexutil.Decode(fromRandHex)
	copy(rand[:], out)
	return &rand
}

func (s *PublicStakeApI) BuyShare(ctx context.Context, args BuyShareTxArg) (common.Hash, error) {
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

func (args *RegistStakePoolTxArg) toPreTxParam() prepare.PreTxParam {
	preTx := prepare.PreTxParam{}
	preTx.From = *args.From.ToUint512()
	preTx.RefundTo = keys.Addr2PKr(args.From.ToUint512(), fromRand()).NewRef()
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(*args.Gas)), args.GasPrice.ToInt())),
	}
	preTx.GasPrice = (*big.Int)(args.GasPrice)
	preTx.Cmds = prepare.Cmds{}
	registPoolCmd := stx.RegistPoolCmd{}
	registPoolCmd.Value = utils.U256(*args.Value.ToInt())
	registPoolCmd.Vote = common.AddrToPKr(*args.Vote)
	registPoolCmd.FeeRate = uint32(*args.Fee)*2/3 + 3334
	preTx.Cmds.RegistPool = &registPoolCmd
	return preTx

}

func (s *PublicStakeApI) RegistStakePool(ctx context.Context, args RegistStakePoolTxArg) (common.Hash, error) {
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

func (s *PublicStakeApI) PoolState(ctx context.Context, pool common.Hash) (map[string]interface{}, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	poolState := stake.NewStakeState(state).GetStakePool(pool)

	if poolState == nil {
		return nil, errors.New("stake pool not exists")
	}
	return newRPCStakePool(*poolState), nil
}

func (s *PublicStakeApI) SharePrice(ctx context.Context) (*hexutil.Big, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	price := stake.NewStakeState(state).CurrentPrice()
	return (*hexutil.Big)(price), nil
}

func (s *PublicStakeApI) SharePoolSize(ctx context.Context) (hexutil.Uint64, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return 0, err
	}
	size := stake.NewStakeState(state).ShareSize()
	return hexutil.Uint64(size), nil
}

type StakePool struct {
	PKr             keys.PKr
	VotePKr         keys.PKr
	TransactionHash common.Hash
	Amount          *big.Int `rlp:"nil"`
	Fee             uint16
	ShareNum        uint32
	ChoicedNum      uint32
	MissedNum       uint32
	WishVotNum      uint32
	Profit          *big.Int `rlp:"nil"`
	LastPayTime     uint64
	Closed          bool
}

func newRPCStakePool(pool stake.StakePool) map[string]interface{} {
	result := map[string]interface{}{}
	result["id"] = common.BytesToHash(pool.Id())
	result["own"] = common.BytesToAddress(pool.PKr[:])
	result["voteAddress"] = common.BytesToAddress(pool.VotePKr[:])
	result["fee"] = hexutil.Uint(pool.Fee)
	result["shareNum"] = hexutil.Uint64(pool.CurrentShareNum)
	result["choicedNum"] = hexutil.Uint64(pool.ChoicedShareNum)
	result["missedNum"] = hexutil.Uint64(pool.MissedVoteNum)
	result["lastPayTime"] = hexutil.Uint64(pool.LastPayTime)
	result["closed"] = pool.Closed
	return result
}

func (s *PublicStakeApI) StakePools(ctx context.Context) []map[string]interface{} {
	pools := stakeservice.CurrentStakeService().StakePools()
	result := []map[string]interface{}{}
	for _, pool := range pools {
		result = append(result, newRPCStakePool(*pool))
	}
	return result
}

func getAccountAddrByPKr(wallets []accounts.Wallet, PKr keys.PKr) interface{} {
	addr := common.Address{}
	copy(addr[:], PKr[:])
	for _, wallet := range wallets {
		if wallet.IsMine(addr) {
			return wallet.Accounts()[0].Address
		}
	}
	return addr
}

func newRPCShare(wallets []accounts.Wallet, share stake.Share) map[string]interface{} {
	s := map[string]interface{}{}
	s["id"] = common.BytesToHash(share.Id())
	s["addr"] = getAccountAddrByPKr(wallets, share.PKr)
	s["voteAddr"] = getAccountAddrByPKr(wallets, share.VotePKr)
	s["total"] = share.InitNum
	s["num"] = share.Num
	s["missed"] = share.WillVoteNum
	s["price"] = share.Value
	s["status"]=share.Status
	if share.PoolId != nil {
		s["pool"] = share.PoolId
	}
	s["returnAmount"]=share.ReturnAmount
	s["tx"] = share.TransactionHash
	return s
}

type RPCStatisticsShare struct{
	Address interface{}     `json:"addr"`
	VoteAddress []interface{}  `json:"voteAddr"`
	Total uint32            `json:"total"`
	Remaining uint32        `json:"remaining"`
	Missed uint32           `json:"missed"`
	Expired uint32           `json:"expired"`
	ShareIds []common.Hash  `json:"shareIds"`
	Pools    []common.Hash   `json:"pools"`
}

func containsVoteAddr(vas[]interface{},item interface{}) bool{
	for _, v := range vas {
		if v== item {
			return true
		}
	}
	return false
}

func containsHash(vas[]common.Hash,item common.Hash) bool{
	for _, v := range vas {
		if v== item {
			return true
		}
	}
	return false
}




func newRPCStatisticsShare(wallets []accounts.Wallet, shares []*stake.Share) []RPCStatisticsShare{
	result := map[string]*RPCStatisticsShare{}
	var key interface{}
	for _, share := range shares {
		key = getAccountAddrByPKr(wallets, share.PKr)
		var keystr string
		switch inst := key.(type) {

		case common.Address:
			keystr = inst.String()
		case address.AccountAddress:
			keystr = inst.String()
		}
		if s, ok := result[keystr]; ok {

			s.Total += share.InitNum
			if share.Status ==stake.STATUS_VALID{
				s.Remaining += share.Num
			}else {
				s.Expired+=share.Num
			}

			s.Missed += share.WillVoteNum
			s.ShareIds=append(	s.ShareIds,common.BytesToHash(share.Id()))
			if !containsVoteAddr(s.VoteAddress,getAccountAddrByPKr(wallets, share.VotePKr)){
				s.VoteAddress=append(s.VoteAddress,getAccountAddrByPKr(wallets, share.VotePKr))
			}
			if share.PoolId!=nil{
				if !containsHash(s.Pools,*share.PoolId){
					s.Pools=append(s.Pools,*share.PoolId)
				}
			}
		} else {
			s := &RPCStatisticsShare{}
			s.Address=key
			s.Total= share.InitNum
			s.Remaining = share.Num
			s.Missed= share.WillVoteNum
			if share.Status ==stake.STATUS_VALID{
				s.Remaining =share.Num
			}else{
				s.Expired =share.Num
			}
			s.VoteAddress=append(s.VoteAddress,getAccountAddrByPKr(wallets, share.VotePKr))
			if share.PoolId!=nil {
				s.Pools=append(s.Pools,*share.PoolId)
			}
			s.ShareIds = append(s.ShareIds, common.BytesToHash(share.Id()))
			result[keystr]=s
		}
	}
	statistics:=[]RPCStatisticsShare{}
	for _,v:= range result{
		statistics=append(statistics,*v)
	}
	return statistics

}

func (s *PublicStakeApI) MyShare(ctx context.Context, addr common.Address) []RPCStatisticsShare {
	var pk keys.Uint512
	wallets := s.b.AccountManager().Wallets()
	if addr.IsAccountAddress() {
		pk = *common.AddrToAccountAddr(addr).ToUint512()
	} else {
		for _, wallet := range wallets {
			if wallet.IsMine(addr) {
				pk = *wallet.Accounts()[0].Address.ToUint512()
				break
			}
		}
	}
	shares := stakeservice.CurrentStakeService().SharesByPk(pk)
	return newRPCStatisticsShare(wallets,shares)
}

func (s *PublicStakeApI) GetShare(ctx context.Context, shareId common.Hash) map[string]interface{} {
	share := stakeservice.CurrentStakeService().SharesById(shareId)
	if share == nil {
		return nil
	}
	wallets := s.b.AccountManager().Wallets()
	return newRPCShare(wallets,*share)
	return nil
}
