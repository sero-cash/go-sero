package ethapi

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rpc"

	"github.com/sero-cash/go-sero/accounts"

	"github.com/sero-cash/go-sero/zero/wallet/stakeservice"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-sero/zero/txtool/prepare"

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
	preTx.RefundTo = c_czero.Addr2PKr(args.From.ToUint512(), nil).NewRef()
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(*args.Gas)), args.GasPrice.ToInt())),
	}
	preTx.GasPrice = (*big.Int)(args.GasPrice)
	preTx.Cmds = prepare.Cmds{}

	buyShareCmd := stx.BuyShareCmd{}
	buyShareCmd.Value = utils.U256(*args.Value.ToInt())
	buyShareCmd.Vote = common.AddrToPKr(*args.Vote)
	if args.Pool != nil {
		var pool c_type.Uint256
		copy(pool[:], (*args.Pool)[:])
		buyShareCmd.Pool = &pool
	}
	preTx.Cmds.BuyShare = &buyShareCmd
	return preTx

}

func (s *PublicStakeApI) EstimateShares(ctx context.Context, args BuyShareTxArg) (map[string]interface{}, error) {
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	num, avprice, baseprice := stake.NewStakeState(state).CaleAvgPrice(args.Value.ToInt())

	result := map[string]interface{}{}
	result["total"] = hexutil.Uint64(num)
	result["avPrice"] = hexutil.Big(*avprice)
	result["basePrice"] = hexutil.Big(*baseprice)
	return result, nil
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

	if uint32(*args.Fee) < seroparam.LOWEST_STAKING_NODE_FEE_RATE {
		return errors.New(fmt.Sprintf("fee rate can not less then %v", seroparam.LOWEST_STAKING_NODE_FEE_RATE))
	}
	if uint32(*args.Fee) > seroparam.HIGHEST_STAKING_NODE_FEE_RATE {
		return errors.New(fmt.Sprintf("fee rate can not large then  %v", seroparam.HIGHEST_STAKING_NODE_FEE_RATE))
	}

	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	return nil
}

func (args *RegistStakePoolTxArg) toPreTxParam() prepare.PreTxParam {
	preTx := prepare.PreTxParam{}
	preTx.From = *args.From.ToUint512()
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(*args.Gas)), args.GasPrice.ToInt())),
	}
	preTx.GasPrice = (*big.Int)(args.GasPrice)
	preTx.Cmds = prepare.Cmds{}
	registPoolCmd := stx.RegistPoolCmd{}
	registPoolCmd.Value = utils.U256(*args.Value.ToInt())
	registPoolCmd.Vote = common.AddrToPKr(*args.Vote)
	registPoolCmd.FeeRate = uint32(*args.Fee)
	preTx.Cmds.RegistPool = &registPoolCmd
	return preTx
}

func (s *PublicStakeApI) RegistStakePool(ctx context.Context, args RegistStakePoolTxArg) (common.Hash, error) {

	if !seroparam.Is_Dev() {
		peerCount := s.b.PeerCount()
		if peerCount < 10 {
			return common.Hash{}, errors.New("connected peer < 10")
		}
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	wallet, err := s.b.AccountManager().Find(accounts.Account{Address: args.From})
	if err != nil {
		return common.Hash{}, err
	}
	fromPkr := getStakePoolPkr(wallet.Accounts()[0])

	poolId := getStakePoolId(fromPkr)
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return common.Hash{}, err
	}
	pool := stake.NewStakeState(state).GetStakePool(poolId)

	if pool != nil {
		return common.Hash{}, errors.New("stake pool has exists poolId=" + poolId.String())
	}

	log.Info("RegistStakePool", "idPkr", common.BytesToAddress(fromPkr[:]).String())
	preTx := args.toPreTxParam()
	preTx.RefundTo = &fromPkr
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

func getStakePoolPkr(account accounts.Account) c_type.PKr {
	randHash := crypto.Keccak256Hash(account.Tk[:])
	var rand c_type.Uint256
	copy(rand[:], randHash[:])
	return c_czero.Addr2PKr(account.Address.ToUint512(), &rand)

}
func getStakePoolId(from c_type.PKr) common.Hash {
	return crypto.Keccak256Hash(from[:])
}

func (s *PublicStakeApI) CloseStakePool(ctx context.Context, from common.Address) (common.Hash, error) {
	wallets := s.b.AccountManager().Wallets()
	var own address.AccountAddress
	var fromPkr c_type.PKr
	if from.IsAccountAddress() {
		own = common.AddrToAccountAddr(from)
		wallet, err := s.b.AccountManager().Find(accounts.Account{Address: own})
		if err != nil {
			return common.Hash{}, err
		}
		fromPkr = getStakePoolPkr(wallet.Accounts()[0])
	} else {
		localAddr := getLocalAccountAddressByPkr(wallets, from)
		if localAddr == nil {
			return common.Hash{}, errors.New("can not find local account")
		}
		fromPkr = *from.ToPKr()
	}
	poolId := getStakePoolId(fromPkr)
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return common.Hash{}, err
	}
	log.Info("close stakepool", "poolId", hexutil.Encode(poolId[:]))
	pool := stake.NewStakeState(state).GetStakePool(poolId)

	if pool == nil {
		return common.Hash{}, errors.New("stake pool not exists")
	}

	if pool.Closed {
		return common.Hash{}, errors.New("stake pool has closed")
	}
	preTx := prepare.PreTxParam{}
	preTx.From = *from.ToUint512()
	preTx.RefundTo = &fromPkr
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(25000)), new(big.Int).SetUint64(defaultGasPrice))),
	}
	preTx.GasPrice = new(big.Int).SetUint64(defaultGasPrice)
	preTx.Cmds = prepare.Cmds{}
	closePoolCmd := stx.ClosePoolCmd{}
	preTx.Cmds.ClosePool = &closePoolCmd

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

func (s *PublicStakeApI) ModifyStakePoolFee(ctx context.Context, from common.Address, fee hexutil.Uint64) (common.Hash, error) {

	if uint32(fee) < seroparam.LOWEST_STAKING_NODE_FEE_RATE {
		return common.Hash{}, errors.New(fmt.Sprintf("fee rate can not less then %v", seroparam.LOWEST_STAKING_NODE_FEE_RATE))
	}
	if uint32(fee) > seroparam.HIGHEST_STAKING_NODE_FEE_RATE {
		return common.Hash{}, errors.New(fmt.Sprintf("fee rate can not large then  %v", seroparam.HIGHEST_STAKING_NODE_FEE_RATE))
	}

	wallets := s.b.AccountManager().Wallets()
	var own address.AccountAddress
	var fromPkr c_type.PKr
	if from.IsAccountAddress() {
		own = common.AddrToAccountAddr(from)
		wallet, err := s.b.AccountManager().Find(accounts.Account{Address: own})
		if err != nil {
			return common.Hash{}, err
		}
		fromPkr = getStakePoolPkr(wallet.Accounts()[0])
	} else {

		localAddr := getLocalAccountAddressByPkr(wallets, from)
		if localAddr == nil {
			return common.Hash{}, errors.New("can not find local account")
		}
		fromPkr = *from.ToPKr()
	}

	poolId := getStakePoolId(fromPkr)
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return common.Hash{}, err
	}

	pool := stake.NewStakeState(state).GetStakePool(poolId)

	if pool == nil {
		return common.Hash{}, errors.New("stake pool not exists")
	}

	if pool.Closed {
		return common.Hash{}, errors.New("stake pool has closed")
	}
	preTx := prepare.PreTxParam{}
	preTx.From = *(own.ToUint512())
	preTx.RefundTo = &fromPkr
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(25000)), new(big.Int).SetUint64(defaultGasPrice))),
	}
	preTx.GasPrice = new(big.Int).SetUint64(defaultGasPrice)
	preTx.Cmds = prepare.Cmds{}
	registPoolCmd := stx.RegistPoolCmd{}
	registPoolCmd.Vote = pool.VotePKr
	registPoolCmd.FeeRate = uint32(fee)
	preTx.Cmds.RegistPool = &registPoolCmd
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

func (s *PublicStakeApI) ModifyStakePoolVote(ctx context.Context, from common.Address, vote common.Address) (common.Hash, error) {
	wallets := s.b.AccountManager().Wallets()
	var own address.AccountAddress
	var fromPkr c_type.PKr
	if from.IsAccountAddress() {
		own = common.AddrToAccountAddr(from)
		wallet, err := s.b.AccountManager().Find(accounts.Account{Address: own})
		if err != nil {
			return common.Hash{}, err
		}
		fromPkr = getStakePoolPkr(wallet.Accounts()[0])
	} else {

		localAddr := getLocalAccountAddressByPkr(wallets, from)
		if localAddr == nil {
			return common.Hash{}, errors.New("can not find local account")
		}
		fromPkr = *from.ToPKr()
	}

	poolId := getStakePoolId(fromPkr)
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return common.Hash{}, err
	}

	pool := stake.NewStakeState(state).GetStakePool(poolId)

	if pool == nil {
		return common.Hash{}, errors.New("stake pool not exists")
	}

	if pool.Closed {
		return common.Hash{}, errors.New("stake pool has closed")
	}
	votePkr := common.AddrToPKr(vote)
	preTx := prepare.PreTxParam{}
	preTx.From = *own.ToUint512()
	preTx.RefundTo = &fromPkr
	preTx.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*big.NewInt(0).Mul(big.NewInt(int64(25000)), new(big.Int).SetUint64(defaultGasPrice))),
	}
	preTx.GasPrice = new(big.Int).SetUint64(defaultGasPrice)
	preTx.Cmds = prepare.Cmds{}
	registPoolCmd := stx.RegistPoolCmd{}
	registPoolCmd.Vote = votePkr
	registPoolCmd.FeeRate = uint32(pool.Fee)
	preTx.Cmds.RegistPool = &registPoolCmd
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

func (s *PublicStakeApI) PoolState(ctx context.Context, poolId common.Hash) (map[string]interface{}, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, -1)
	if err != nil {
		return nil, err
	}
	wallets := s.b.AccountManager().Wallets()
	poolState := stake.NewStakeState(state).GetStakePool(poolId)

	if poolState == nil {
		return nil, errors.New("stake pool not exists")
	}

	timestamp := uint64(0)
	block, _ := s.b.BlockByNumber(ctx, rpc.BlockNumber(poolState.BlockNumber))

	if block != nil {
		timestamp = block.Header().Time.Uint64()
	}

	ret := newRPCStakePool(wallets, *poolState, timestamp)

	if poolState.LastPayTime != 0 {
		header, _ := s.b.HeaderByNumber(ctx, rpc.BlockNumber(poolState.LastPayTime))
		snapshot := stake.GetStakePoolByBlockNumber(s.b.ChainDb(), poolId, header.Hash(), header.Number.Uint64())
		if snapshot != nil {
			ret["returnProfit"] = hexutil.Big(*snapshot.Profit)
		}
	}
	return ret, nil
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
	PKr             c_type.PKr
	VotePKr         c_type.PKr
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

func newRPCStakePool(wallets []accounts.Wallet, pool stake.StakePool, timestamp uint64) map[string]interface{} {
	result := map[string]interface{}{}
	result["id"] = common.BytesToHash(pool.Id())
	result["idPkr"] = common.BytesToAddress(pool.PKr[:])
	result["own"] = getAccountAddrByPKr(wallets, pool.PKr)
	result["voteAddress"] = getAccountAddrByPKr(wallets, pool.VotePKr)
	result["fee"] = hexutil.Uint(pool.Fee)
	result["shareNum"] = hexutil.Uint64(pool.CurrentShareNum)
	result["choicedNum"] = hexutil.Uint64(pool.ChoicedShareNum)
	result["wishVoteNum"] = hexutil.Uint64(pool.WishVoteNum)
	result["expireNum"] = hexutil.Uint64(pool.ExpireNum)
	result["missedNum"] = hexutil.Uint64(pool.MissedVoteNum)
	result["profit"] = hexutil.Big(*pool.Profit)
	result["lastPayTime"] = hexutil.Uint64(pool.LastPayTime)
	result["closed"] = pool.Closed
	result["tx"] = pool.TransactionHash
	result["createAt"] = hexutil.Uint64(pool.BlockNumber)
	result["timestamp"] = hexutil.Uint64(timestamp)
	return result
}

func (s *PublicStakeApI) StakePools(ctx context.Context) []map[string]interface{} {
	pools := stakeservice.CurrentStakeService().StakePools()
	result := []map[string]interface{}{}
	wallets := s.b.AccountManager().Wallets()
	for _, pool := range pools {
		timestamp := uint64(0)
		block, _ := s.b.BlockByNumber(ctx, rpc.BlockNumber(pool.BlockNumber))

		if block != nil {
			timestamp = block.Header().Time.Uint64()
		}
		result = append(result, newRPCStakePool(wallets, *pool, timestamp))
	}
	return result
}

func getAccountAddrByPKr(wallets []accounts.Wallet, PKr c_type.PKr) interface{} {
	addr := common.Address{}
	copy(addr[:], PKr[:])
	for _, wallet := range wallets {
		if wallet.IsMine(addr) {
			return wallet.Accounts()[0].Address
		}
	}
	return addr
}

func newRPCShare(wallets []accounts.Wallet, share stake.Share, timestamp uint64) map[string]interface{} {
	s := map[string]interface{}{}
	s["id"] = common.BytesToHash(share.Id())
	s["addr"] = getAccountAddrByPKr(wallets, share.PKr)
	s["voteAddr"] = getAccountAddrByPKr(wallets, share.VotePKr)
	s["total"] = hexutil.Uint64(share.InitNum)
	s["missed"] = hexutil.Uint64(share.WillVoteNum)
	if share.Value != nil {
		s["price"] = hexutil.Big(*share.Value)
	}
	if share.Status == stake.STATUS_VALID {
		s["remaining"] = hexutil.Uint64(share.Num)
	} else {
		s["expired"] = hexutil.Uint64(share.Num)
	}
	s["status"] = hexutil.Uint64(share.Status)
	if share.PoolId != nil {
		s["pool"] = share.PoolId
	}
	if share.Profit != nil {
		s["profit"] = hexutil.Big(*share.Profit)
	}
	s["fee"] = hexutil.Uint64(share.Fee)

	s["tx"] = share.TransactionHash
	s["at"] = hexutil.Uint64(share.BlockNumber)
	s["timestamp"] = hexutil.Uint64(timestamp)
	return s
}

func newRPCStaticsShareMap(rs RPCStatisticsShare) map[string]interface{} {
	s := map[string]interface{}{}
	s["addr"] = rs.Address
	s["voteAddr"] = rs.VoteAddress
	s["total"] = hexutil.Uint64(rs.Total)
	s["missed"] = hexutil.Uint64(rs.Missed)
	s["remaining"] = hexutil.Uint64(rs.Remaining)
	s["expired"] = hexutil.Uint64(rs.Expired)
	s["shareIds"] = rs.ShareIds
	s["profit"] = hexutil.Big(*rs.Profit)
	s["pools"] = rs.Pools
	if rs.TotalAmount != nil {
		s["totalAmount"] = hexutil.Big(*rs.TotalAmount)
	}
	return s
}

type RPCStatisticsShare struct {
	Address     interface{}   `json:"addr"`
	VoteAddress []interface{} `json:"voteAddr"`
	Total       uint32        `json:"total"`
	Remaining   uint32        `json:"remaining"`
	Missed      uint32        `json:"missed"`
	Expired     uint32        `json:"expired"`
	ShareIds    []common.Hash `json:"shareIds"`
	Pools       []common.Hash `json:"pools"`
	Profit      *big.Int      `json:"profit"`
	TotalAmount *big.Int      `json:"totalAmount"`
}

func containsVoteAddr(vas []interface{}, item interface{}) bool {
	for _, v := range vas {
		if v == item {
			return true
		}
	}
	return false
}

func containsHash(vas []common.Hash, item common.Hash) bool {
	for _, v := range vas {
		if v == item {
			return true
		}
	}
	return false
}

func newRPCStatisticsShare(wallets []accounts.Wallet, shares []*stake.Share, api *PublicStakeApI, ctx context.Context) []map[string]interface{} {
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
		var s *RPCStatisticsShare
		if _, ok := result[keystr]; ok {
			s = result[keystr]
			s.Total += share.InitNum
			if share.Status == stake.STATUS_VALID {
				s.Remaining += share.Num
			} else {
				s.Expired += share.Num
			}

			s.Missed += share.WillVoteNum
			s.ShareIds = append(s.ShareIds, common.BytesToHash(share.Id()))
			if !containsVoteAddr(s.VoteAddress, getAccountAddrByPKr(wallets, share.VotePKr)) {
				s.VoteAddress = append(s.VoteAddress, getAccountAddrByPKr(wallets, share.VotePKr))
			}
			if share.PoolId != nil {
				if !containsHash(s.Pools, *share.PoolId) {
					s.Pools = append(s.Pools, *share.PoolId)
				}
			}
			if share.Profit != nil {
				s.Profit = big.NewInt(0).Add(s.Profit, share.Profit)
			}

		} else {
			s = &RPCStatisticsShare{}
			s.Address = key
			s.Total = share.InitNum
			s.Missed = share.WillVoteNum
			if share.Status == stake.STATUS_VALID {
				s.Remaining = share.Num
			} else {
				s.Expired = share.Num
			}
			s.VoteAddress = append(s.VoteAddress, getAccountAddrByPKr(wallets, share.VotePKr))
			if share.PoolId != nil {
				s.Pools = append(s.Pools, *share.PoolId)
			}
			s.Profit = new(big.Int).Set(share.Profit)
			s.ShareIds = append(s.ShareIds, common.BytesToHash(share.Id()))

			result[keystr] = s
		}

		if share.LastPayTime > share.BlockNumber+198720 {
			continue
		}
		remain := share.InitNum
		if share.LastPayTime != 0 {
			header, _ := api.b.HeaderByNumber(ctx, rpc.BlockNumber(share.LastPayTime))
			snapshot := stake.GetShareByBlockNumber(api.b.ChainDb(), common.BytesToHash(share.Id()), header.Hash(), header.Number.Uint64())
			if snapshot != nil {
				remain = snapshot.Num + snapshot.WillVoteNum
			}
		}
		if s.TotalAmount == nil {
			s.TotalAmount = new(big.Int)
		}
		s.TotalAmount = new(big.Int).Add(s.TotalAmount, new(big.Int).Mul(big.NewInt(int64(remain)), share.Value))
	}
	statistics := []map[string]interface{}{}
	for _, v := range result {
		statistics = append(statistics, newRPCStaticsShareMap(*v))
	}
	return statistics

}

func (s *PublicStakeApI) MyShare(ctx context.Context, addr common.Address) []map[string]interface{} {
	var pk c_type.Uint512
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
	return newRPCStatisticsShare(wallets, shares, s, ctx)
}

func (s *PublicStakeApI) GetShare(ctx context.Context, shareId common.Hash) map[string]interface{} {
	share := stakeservice.CurrentStakeService().SharesById(shareId)
	if share == nil {
		return nil
	}
	wallets := s.b.AccountManager().Wallets()
	timestamp := uint64(0)
	block, _ := s.b.BlockByNumber(ctx, rpc.BlockNumber(share.BlockNumber))

	if block != nil {
		timestamp = block.Header().Time.Uint64()
	}
	ret := newRPCShare(wallets, *share, timestamp)

	if share.LastPayTime != 0 {
		header, _ := s.b.HeaderByNumber(ctx, rpc.BlockNumber(share.LastPayTime))
		snapshot := stake.GetShareByBlockNumber(s.b.ChainDb(), shareId, header.Hash(), header.Number.Uint64())
		if snapshot != nil {
			if snapshot.Status == 1 {
				ret["returnNum"] = hexutil.Uint64(snapshot.InitNum - snapshot.WillVoteNum)
			} else if snapshot.Status == 2 {
				ret["returnNum"] = hexutil.Uint64(snapshot.InitNum)
			} else {
				ret["returnNum"] = hexutil.Uint64(snapshot.InitNum - snapshot.Num - snapshot.WillVoteNum)
			}

			ret["returnProfit"] = hexutil.Big(*snapshot.Profit)
		}
		ret["lastPayTime"] = hexutil.Uint64(share.LastPayTime)
	}
	return ret
}
func (s *PublicStakeApI) GetShareByPkr(ctx context.Context, pkr PKrAddress) []map[string]interface{} {
	wallets := s.b.AccountManager().Wallets()
	shares := stakeservice.CurrentStakeService().SharesByPkr(*(pkr.ToPKr()))
	return newRPCStatisticsShare(wallets, shares, s, ctx)
}

func (s *PublicStakeApI) GetStakeInfo(ctx context.Context, poolId common.Hash, start, end hexutil.Uint64) (ret map[string][]interface{}) {
	pools := []interface{}{}
	shares := []interface{}{}
	for start < end {
		header, err := s.b.HeaderByNumber(ctx, rpc.BlockNumber(start))
		if err != nil || header == nil {
			return
		}
		shareList, poolList := stake.GetBlockRecords(s.b.ChainDb(), header.Hash(), uint64(start))
		for _, each := range shareList {
			if each.PoolId != nil && *each.PoolId == poolId {
				share := map[string]interface{}{}
				share["id"] = common.BytesToHash(each.Id())
				share["own"] = base58.Encode(each.PKr[:])
				share["blockNumber"] = hexutil.Uint64(header.Number.Uint64())
				share["total"] = hexutil.Uint64(each.InitNum)
				share["missed"] = hexutil.Uint64(each.WillVoteNum)
				share["price"] = hexutil.Big(*each.Value)
				share["remaining"] = hexutil.Uint64(each.Num)
				share["status"] = hexutil.Uint64(each.Status)
				if each.PoolId != nil {
					share["pool"] = each.PoolId
				}
				share["profit"] = hexutil.Big(*each.Profit)
				share["lastPayTime"] = hexutil.Uint64(each.LastPayTime)
				share["fee"] = hexutil.Uint64(each.Fee)
				share["tx"] = each.TransactionHash
				shares = append(shares, share)
			}
		}

		for _, each := range poolList {
			if bytes.Equal(each.Id(), poolId[:]) {
				pool := map[string]interface{}{}
				pool["id"] = common.BytesToHash(each.Id())
				pool["own"] = base58.Encode(each.PKr[:])
				pool["blockNumber"] = hexutil.Uint64(header.Number.Uint64())
				pool["fee"] = hexutil.Uint64(each.Fee)
				pool["shareNum"] = hexutil.Uint64(each.CurrentShareNum)
				pool["choicedNum"] = hexutil.Uint64(each.ChoicedShareNum)
				pool["wishVoteNum"] = hexutil.Uint64(each.WishVoteNum)
				pool["expireNum"] = hexutil.Uint64(each.ExpireNum)
				pool["missedNum"] = hexutil.Uint64(each.MissedVoteNum)
				pool["profit"] = hexutil.Big(*each.Profit)
				pool["lastPayTime"] = hexutil.Uint64(each.LastPayTime)
				pool["tx"] = each.TransactionHash
				pool["createAt"] = hexutil.Uint64(each.BlockNumber)
				pool["closed"] = each.Closed

				pools = append(pools, pool)
			}
		}
		start++
	}
	ret = map[string][]interface{}{}
	ret["pools"] = pools
	ret["shares"] = shares
	return
}
func (s *PublicStakeApI) Shares(ctx context.Context) (shares []*stake.Share) {
	return stakeservice.CurrentStakeService().Shares()
}

func (s *PublicStakeApI) GetShareAtNumber(ctx context.Context, shareId common.Hash, num hexutil.Uint64) (share *stake.Share) {

	header, _ := s.b.HeaderByNumber(ctx, rpc.BlockNumber(num))
	share = stake.GetShareByBlockNumber(s.b.ChainDb(), shareId, header.Hash(), header.Number.Uint64())
	return
}
