package flight

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type SRI struct {
}

var SRI_Inst = SRI{}

func Trace2Root(tk *keys.Uint512, trace *keys.Uint256) (root *keys.Uint256) {
	root_cm := cpt.FetchRootCM(tk, trace)
	root = localdb.GetRootByRootCM(txtool.Ref_inst.Bc.GetDB(), &root_cm)
	return
}

func GetOut(root *keys.Uint256, num uint64) (out *localdb.RootState) {
	rs := localdb.GetRoot(txtool.Ref_inst.Bc.GetDB(), root)
	if rs != nil {
		return rs
	} else {
		zst := txtool.Ref_inst.CurrentState()
		if os := zst.State.GetOut(root); os == nil {
			return nil
		} else {
			out := localdb.RootState{
				*os,
				keys.Uint256{},
				num,
			}
			return &out
		}
	}
}

func GetBlock(num uint64, hash *common.Hash) (ret *localdb.Block) {
	ret = localdb.GetBlock(txtool.Ref_inst.Bc.GetDB(), num, hash.HashToUint256())
	if ret == nil {
		temp_state := txtool.Ref_inst.Bc.CurrentState(hash)
		if temp_state == nil {
			panic(fmt.Sprintf("new zstate error: %v:%v !", num, hash))
		} else {
			log.Debug("STATE1_PARSE GO BACK TO STATE: ", "num", num, "hash", hash)
		}
		ret = &localdb.Block{}
		ret.Pkgs = temp_state.Pkgs.GetPkgHashes()
		ret.Roots = temp_state.State.GetBlockRoots()
		ret.Dels = temp_state.State.GetBlockDels()
	}
	return
}

func (self *SRI) GetBlocksInfoByDelay(start uint64, count uint64, delay uint64) (blocks []txtool.Block, e error) {
	stable_num := txtool.Ref_inst.GetDelayedNum(delay)
	if start <= stable_num {
		if stable_num-start+1 < count {
			count = stable_num - start + 1
		}
		for i := uint64(0); i < count; i++ {
			num := start + i
			chain_block := txtool.Ref_inst.Bc.GetBlockByNumber(num)
			hash := chain_block.Hash()
			local_block := GetBlock(num, &hash)
			if local_block != nil {
				block := txtool.Block{}
				block.Hash = *hash.HashToUint256()
				block.Num = hexutil.Uint64(num)
				for _, k := range local_block.Dels {
					block.Nils = append(block.Nils, k)
				}
				for _, k := range local_block.Roots {
					if out := GetOut(&k, num); out == nil {
						log.Error("GetBlocksInfo ERROR", "num", num, "root", k)
					} else {
						block.Outs = append(block.Outs, txtool.Out{k, *out})
					}
				}
				for _, k := range local_block.Pkgs {
					if pkg := localdb.GetPkg(txtool.Ref_inst.Bc.GetDB(), &k); pkg == nil {
						log.Error("GetBlocksInfo ERROR", "num", num, "pkg", k)
					} else {
						block.Pkgs = append(block.Pkgs, *pkg)
					}
				}
				blocks = append(blocks, block)
			} else {
				e = fmt.Errorf("GetBlocksInfo.GetBlock Failed, num: %v", num)
				return
			}
		}
		return
	} else {
		return
	}
}

func (self *SRI) GetBlocksInfo(start uint64, count uint64) (blocks []txtool.Block, e error) {
	return self.GetBlocksInfoByDelay(start, count, seroparam.DefaultConfirmedBlock())
}

func (self *SRI) GetAnchor(roots []keys.Uint256) (wits []txtool.Witness, e error) {
	state := txtool.Ref_inst.CurrentState()
	if state != nil {
		for _, root := range roots {
			wit := txtool.Witness{}
			if out := GetOut(&root, 0); out == nil {
				e = errors.New("GetAnchor use root but out is nil !!!")
				return
			} else {
				pos, paths, anchor := state.State.MTree.GetPaths(*out.OS.RootCM)
				wit.Pos = hexutil.Uint64(pos)
				wit.Paths = paths
				wit.Anchor = anchor
				wits = append(wits, wit)
			}
		}
		return
	} else {
		e = errors.New("State is nil")
		return
	}
	return
}

func GenTxParam(param *PreTxParam, tk keys.Uint512) (p txtool.GTxParam, e error) {
	log.Debug("genTx start")
	p.Gas = param.Gas
	p.GasPrice = big.NewInt(0).SetUint64(param.GasPrice)
	p.Fee = assets.Token{
		utils.CurrencyToUint256("SERO"),
		utils.U256(*new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), new(big.Int).SetUint64(param.GasPrice))),
	}
	p.From.PKr = param.From

	p.Outs = param.Outs

	skr := keys.PKr{}
	copy(skr[:], tk[:])

	roots := []keys.Uint256{}
	outs := []txtool.Out{}

	amounts := make(map[string]*big.Int)
	ticekts := make(map[keys.Uint256]keys.Uint256)
	for _, in := range param.Ins {
		roots = append(roots, in)
		if root := localdb.GetRoot(txtool.Ref_inst.Bc.GetDB(), &in); root == nil {
			e = fmt.Errorf("SRI.GenTxParam get root Error for root %v", in)
			return
		} else {
			out := txtool.Out{in, *root}
			dOuts := DecTraceOuts([]txtool.Out{out}, &skr)
			if len(dOuts) == 0 {
				e = fmt.Errorf("SRI.GenTxParam dec outs Error for root %v", in)
				return
			}
			oOut := dOuts[0]
			if len(oOut.Nils) == 0 {
				e = fmt.Errorf("SRI.GenTxParam dec outs Error for root %v", in)
				return
			}
			if oOut.Asset.Tkn != nil {
				currency := strings.Trim(string(oOut.Asset.Tkn.Currency[:]), string([]byte{0}))
				if amount, ok := amounts[currency]; ok {
					amount.Add(amount, oOut.Asset.Tkn.Value.ToIntRef())
				} else {
					amounts[currency] = oOut.Asset.Tkn.Value.ToIntRef()
				}

			}
			if oOut.Asset.Tkt != nil {
				ticekts[oOut.Asset.Tkt.Value] = oOut.Asset.Tkt.Category
			}
			outs = append(outs, txtool.Out{in, *root})
		}
	}

	for _, out := range param.Outs {
		if out.Asset.Tkn != nil {
			currency := strings.Trim(string(out.Asset.Tkn.Currency[:]), string([]byte{0}))
			token := out.Asset.Tkn.Value.ToIntRef()
			if amount, ok := amounts[currency]; ok && amount.Cmp(token) >= 0 {
				amount.Sub(amount, token)
				if amount.Sign() == 0 {
					delete(amounts, currency)
				}
			} else {
				e = fmt.Errorf("SSI GenTx Error: balance is not enough")
				return
			}
		}
		if out.Asset.Tkt != nil {
			if value, ok := ticekts[out.Asset.Tkt.Value]; ok && value == out.Asset.Tkt.Category {
				delete(ticekts, out.Asset.Tkt.Value)
			} else {
				e = fmt.Errorf("SSI GenTx Erro: balance is not enough")
				return
			}
		}
	}

	if amount, ok := amounts[utils.Uint256ToCurrency(&p.Fee.Currency)]; !ok || amount.Cmp(p.Fee.Value.ToInt()) < 0 {
		e = fmt.Errorf("SSI GenTx Error: sero amount < Fee")
		return
	} else {
		amount.Sub(amount, p.Fee.Value.ToInt())
		if amount.Sign() == 0 {
			delete(amounts, utils.Uint256ToCurrency(&p.Fee.Currency))
		}
	}

	if len(amounts) > 0 || len(ticekts) > 0 {
		for currency, value := range amounts {
			p.Outs = append(p.Outs, txtool.GOut{PKr: p.From.PKr, Asset: assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*value),
			}}})
		}
		for value, category := range ticekts {
			p.Outs = append(p.Outs, txtool.GOut{PKr: p.From.PKr, Asset: assets.Asset{Tkt: &assets.Ticket{
				Category: category,
				Value:    value,
			}}})
		}
	}

	wits := []txtool.Witness{}

	if wits, e = SRI_Inst.GetAnchor(roots); e != nil {
		return
	}

	for i := 0; i < len(wits); i++ {
		in := txtool.GIn{}
		in.Out = outs[i]
		in.Witness = wits[i]
		p.Ins = append(p.Ins, in)
	}

	log.Debug("genTxParam ins : %v, outs : %v", len(p.Ins), len(p.Outs))
	return
}
