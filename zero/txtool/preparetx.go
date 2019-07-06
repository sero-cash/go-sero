package txtool

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/sero-cash/go-sero/zero/txs/stx"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Reception struct {
	Addr  keys.PKr
	Asset assets.Asset
}

type PkgCloseCmd struct {
	Id      keys.Uint256
	Owner   keys.PKr
	AssetCM keys.Uint256
	Ar      keys.Uint256
}

type PkgTransferCmd struct {
	Id    keys.Uint256
	Owner keys.PKr
	PKr   keys.PKr
}

type PkgCreateCmd struct {
	Id    keys.Uint256
	PKr   keys.PKr
	Asset assets.Asset
	Memo  keys.Uint512
	Ar    keys.Uint256
}

type Cmds struct {
	//Share
	BuyShare *stx.BuyShareCmd
	//Pool
	RegistPool *stx.RegistPoolCmd
	ClosePool  *stx.ClosePoolCmd
	//Contract
	Contract *stx.ContractCmd
	//Package
	PkgCreate   *PkgCreateCmd
	PkgTransfer *PkgTransferCmd
	PkgClose    *PkgCloseCmd
}

func (self *Cmds) Asset() *assets.Asset {
	if self.PkgCreate != nil {
		return &self.PkgCreate.Asset
	}
	if self.BuyShare != nil {
		asset := self.BuyShare.Asset()
		return &asset
	}
	if self.RegistPool != nil {
		asset := self.RegistPool.Asset()
		return &asset
	}
	if self.Contract != nil {
		return &self.Contract.Asset
	}
	return nil
}

func (self *Cmds) Valid() bool {
	count := 0
	if self.PkgCreate != nil {
		count++
	}
	if self.PkgTransfer != nil {
		count++
	}
	if self.PkgClose != nil {
		count++
	}
	if self.BuyShare != nil {
		count++
	}
	if self.RegistPool != nil {
		count++
	}
	if self.ClosePool != nil {
		count++
	}
	if self.Contract != nil {
		count++
	}
	if count <= 1 {
		return true
	} else {
		return false
	}
}

type PreTxParam struct {
	From       keys.Uint512
	RefundTo   *keys.PKr
	Receptions []Reception
	Cmds       Cmds
	Gas        uint64
	GasPrice   *big.Int
	Roots      []keys.Uint256
}

type Utxo struct {
	Root  keys.Uint256
	Asset assets.Asset
}

type Utxos []Utxo

func (self *Utxos) Roots() (roots []keys.Uint256) {
	for _, utxo := range *self {
		roots = append(roots, utxo.Root)
	}
	return
}

type TxParamGenerator interface {
	FindRoots(pk *keys.Uint512, currency string, amount *big.Int) (utxos Utxos, remain big.Int)
	FindRootsByTicket(pk *keys.Uint512, tickets map[keys.Uint256]keys.Uint256) (roots Utxos, remain map[keys.Uint256]keys.Uint256)
	GetRoot(root *keys.Uint256) (utxos *Utxo)
	DefaultRefundTo(from *keys.Uint512) (ret *keys.PKr)
}

func GenTxParam(param *PreTxParam, gen TxParamGenerator) (txParam *GTxParam, e error) {
	if len(param.Receptions) > 500 {
		return nil, errors.New("receptions count must <= 500")
	}
	utxos, err := PreGenTx(param, gen)
	if err != nil {
		return nil, err
	}

	if param.RefundTo == nil {
		if param.RefundTo = gen.DefaultRefundTo(&param.From); param.RefundTo == nil {
			return nil, errors.New("can not find default refund to")
		}
	}
	txParam, e = BuildTxParam(utxos, param.RefundTo, param.Receptions, &param.Cmds, param.Gas, param.GasPrice)
	return
}

func PreGenTx(param *PreTxParam, gen TxParamGenerator) (utxos Utxos, err error) {
	if len(param.Roots) > 0 {
		for _, root := range param.Roots {
			if utxo := gen.GetRoot(&root); utxo == nil {
				return utxos, fmt.Errorf("can not find the utxo for root : %v", hexutil.Encode(root[:]))
			} else {
				utxos = append(utxos, *utxo)
			}
		}
	} else {
		amounts := map[string]*big.Int{}
		tickets := map[keys.Uint256]keys.Uint256{}
		if len(param.Receptions) > 0 {
			for _, each := range param.Receptions {
				if each.Asset.Tkn != nil {
					currency := common.BytesToString(each.Asset.Tkn.Currency[:])
					if amount, ok := amounts[currency]; ok {
						amount.Add(amount, each.Asset.Tkn.Value.ToInt())
					} else {
						amounts[currency] = new(big.Int).Set(each.Asset.Tkn.Value.ToInt())
					}
				}
				if each.Asset.Tkt != nil {
					if _, ok := tickets[each.Asset.Tkt.Value]; ok {
						return utxos, errors.New("duplicate ticket")
					} else {
						tickets[each.Asset.Tkt.Value] = each.Asset.Tkt.Category
					}
				}
			}
		}

		if asset := param.Cmds.Asset(); asset != nil {
			if asset.Tkn != nil {
				currency := common.BytesToString(asset.Tkn.Currency[:])
				if amount, ok := amounts[currency]; ok {
					amount.Add(amount, asset.Tkn.Value.ToInt())
				} else {
					amounts[currency] = new(big.Int).Set(asset.Tkn.Value.ToInt())
				}
			}
			if asset.Tkt != nil {
				if _, ok := tickets[asset.Tkt.Value]; ok {
					return utxos, errors.New("duplicate ticket")
				} else {
					tickets[asset.Tkt.Value] = asset.Tkt.Category
				}
			}
		}

		if amount, ok := amounts["SERO"]; ok {
			amount.Add(amount, new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), param.GasPrice))
		} else {
			amounts["SERO"] = new(big.Int).Mul(new(big.Int).SetUint64(param.Gas), param.GasPrice)
		}

		for currency, amount := range amounts {
			list, remain := gen.FindRoots(&param.From, currency, amount)
			if remain.Sign() > 0 {
				return list, errors.New(fmt.Sprintf("not enough token, maximum available token is %s", new(big.Int).Sub(amount, &remain).String()))
			} else {
				utxos = append(utxos, list...)
				for _, utxo := range list {
					if utxo.Asset.Tkt != nil {
						if category, ok := tickets[utxo.Asset.Tkt.Value]; ok {
							if category != utxo.Asset.Tkt.Category {
								return list, errors.New("ticket error")
							}
							delete(tickets, utxo.Asset.Tkt.Value)
						}
					}
				}
			}
		}

		list, remainTicket := gen.FindRootsByTicket(&param.From, tickets)
		if len(remainTicket) > 0 {
			return list, errors.New("not enough ticket")
		}
		utxos = append(utxos, list...)
	}
	return
}

func BuildTxParam(
	utxos Utxos,
	refundTo *keys.PKr,
	receptions []Reception,
	cmds *Cmds,
	gas uint64,
	gasPrice *big.Int) (txParam *GTxParam, e error) {

	txParam = new(GTxParam)
	txParam.Gas = gas
	txParam.GasPrice = *gasPrice

	txParam.From = Kr{PKr: *refundTo}

	wits, err := SRI_Inst.GetAnchor(utxos.Roots())
	if err != nil {
		e = err
		return
	}

	Ins := []GIn{}
	amounts := make(map[string]*big.Int)
	ticekts := make(map[keys.Uint256]keys.Uint256)
	oins_count := 0
	for index, utxo := range utxos {
		if out := GetOut(&utxo.Root, 0); out != nil {
			need_add := false
			if utxo.Asset.Tkn != nil {
				if utxo.Asset.Tkn.Value.Cmp(&utils.U256_0) != 0 {
					currency := strings.Trim(string(utxo.Asset.Tkn.Currency[:]), string([]byte{0}))
					if amount, ok := amounts[currency]; ok {
						amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
					} else {
						amounts[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
					}
					need_add = true
				}
			}
			if utxo.Asset.Tkt != nil {
				if utxo.Asset.Tkt.Value != keys.Empty_Uint256 {
					ticekts[utxo.Asset.Tkt.Value] = utxo.Asset.Tkt.Category
					need_add = true
				}
			}

			if need_add {
				Ins = append(Ins, GIn{Out: Out{Root: utxo.Root, State: *out}, Witness: wits[index]})
				if out.OS.Out_O != nil {
					oins_count++
				}
			}
		}
	}

	if oins_count > 2500 {
		e = fmt.Errorf("o_ins count > 2500")
		return
	}

	Outs := []GOut{}
	for _, reception := range receptions {
		asset := &assets.Asset{}
		pkr := reception.Addr
		if IsPk(reception.Addr) {
			pk := reception.Addr.ToUint512()
			pkr = CreatePkr(&pk, 1)
		}
		if reception.Asset.Tkn == nil && reception.Asset.Tkt == nil {
			continue
		} else {
			if reception.Asset.Tkn != nil {
				e = checkTokenAsset(reception.Asset.Tkn, amounts)
				if e != nil {
					return
				}
			}
			if reception.Asset.Tkt != nil {
				e = checkTicketAsset(reception.Asset.Tkt, ticekts)
				if e != nil {
					return
				}
			}
		}
		Outs = append(Outs, GOut{PKr: pkr, Asset: *asset})
	}

	if cmds != nil {
		if asset := cmds.Asset(); asset != nil {
			if asset.Tkn == nil && asset.Tkt == nil {
			} else {
				if asset.Tkn != nil {
					e = checkTokenAsset(asset.Tkn, amounts)
					if e != nil {
						return
					}
				}
				if asset.Tkt != nil {
					e = checkTicketAsset(asset.Tkt, ticekts)
					if e != nil {
						return
					}
				}
			}
		}
	}

	fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), gasPrice)
	if amount, ok := amounts["SERO"]; !ok || amount.Cmp(fee) < 0 {
		e = fmt.Errorf("Exchange Error: not enough")
		return
	} else {
		amount.Sub(amount, fee)
		if amount.Sign() == 0 {
			delete(amounts, "SERO")
		}
	}

	if len(amounts) > 0 {
		for currency, value := range amounts {
			Outs = append(Outs, GOut{PKr: txParam.From.PKr, Asset: assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
				Value:    utils.U256(*value),
			}}})
		}
	}
	if len(ticekts) > 0 {
		for value, category := range ticekts {
			Outs = append(Outs, GOut{PKr: txParam.From.PKr, Asset: assets.Asset{Tkt: &assets.Ticket{
				Category: category,
				Value:    value,
			}}})
		}
	}

	txParam.Ins = Ins
	txParam.Outs = Outs
	txParam.Cmds = *cmds

	return
}

func checkTokenAsset(token *assets.Token, amounts map[string]*big.Int) (e error) {
	currency := common.BytesToString(token.Currency[:])
	if amount, ok := amounts[currency]; ok && amount.Cmp(token.Value.ToInt()) >= 0 {
		amount.Sub(amount, token.Value.ToInt())
		if amount.Sign() == 0 {
			delete(amounts, currency)
		}
	} else {
		e = errors.New("build token error")
		return
	}
	return
}

func checkTicketAsset(ticket *assets.Ticket, tickets map[keys.Uint256]keys.Uint256) (e error) {
	if category, ok := tickets[ticket.Value]; ok && category == ticket.Category {
		delete(tickets, ticket.Value)
	} else {
		e = errors.New("build ticket error")
		return
	}
	return
}

func IsPk(addr keys.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func CreatePkr(pk *keys.Uint512, index uint64) keys.PKr {
	r := keys.Uint256{}
	copy(r[:], common.LeftPadBytes(EncodeNumber(index), 32))
	return keys.Addr2PKr(pk, &r)
}

func EncodeNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func DecodeNumber(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}
