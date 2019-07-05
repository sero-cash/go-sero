package txtool

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type Reception struct {
	Addr     keys.PKr
	Currency string
	Value    *big.Int
}

type PreTxParam struct {
	From       keys.Uint512
	RefundTo   *keys.PKr
	Receptions []Reception
	Gas        uint64
	GasPrice   *big.Int
	Roots      []keys.Uint256
}

type TxParamGenerator interface {
	FindRoots(pk *keys.Uint512, currency string, amount *big.Int) (roots []keys.Uint256, remain big.Int)
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
	txParam, e = BuildTxParam(utxos, param.RefundTo, param.Receptions, param.Gas, param.GasPrice)
	return
}

func PreGenTx(param *PreTxParam, gen TxParamGenerator) (roots []keys.Uint256, err error) {
	if len(param.Roots) > 0 {
		roots = param.Roots
	} else {
		amounts := map[string]*big.Int{}
		for _, each := range param.Receptions {
			if amount, ok := amounts[each.Currency]; ok {
				amount.Add(amount, each.Value)
			} else {
				amounts[each.Currency] = new(big.Int).Set(each.Value)
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
				return roots, errors.New(fmt.Sprintf("not enough token, maximum available token is %s", new(big.Int).Sub(amount, &remain).String()))
			} else {
				roots = append(roots, list...)
			}
		}
	}
	return
}

func BuildTxParam(
	roots []keys.Uint256,
	refundTo *keys.PKr,
	receptions []Reception,
	gas uint64,
	gasPrice *big.Int) (txParam *GTxParam, e error) {

	txParam = new(GTxParam)
	txParam.Gas = gas
	txParam.GasPrice = *gasPrice

	txParam.From = Kr{PKr: *refundTo}

	Ins := []GIn{}
	wits, err := SRI_Inst.GetAnchor(roots)
	if err != nil {
		e = err
		return
	}

	amounts := make(map[string]*big.Int)
	ticekts := make(map[keys.Uint256]keys.Uint256)
	oins_count := 0
	for index, utxo := range roots {
		if out := GetOut(&utxo, 0); out != nil {
			need_add := false
			if out.OS.Out_O.Asset.Tkn != nil {
				if out.OS.Out_O.Asset.Tkn.Value.Cmp(&utils.U256_0) != 0 {
					currency := strings.Trim(string(out.OS.Out_O.Asset.Tkn.Currency[:]), string([]byte{0}))
					if amount, ok := amounts[currency]; ok {
						amount.Add(amount, out.OS.Out_O.Asset.Tkn.Value.ToIntRef())
					} else {
						amounts[currency] = new(big.Int).Set(out.OS.Out_O.Asset.Tkn.Value.ToIntRef())
					}
					need_add = true
				}
			}
			if out.OS.Out_O.Asset.Tkt != nil {
				if out.OS.Out_O.Asset.Tkt.Value != keys.Empty_Uint256 {
					ticekts[out.OS.Out_O.Asset.Tkt.Value] = out.OS.Out_O.Asset.Tkt.Category
					need_add = true
				}
			}

			if need_add {
				Ins = append(Ins, GIn{Out: Out{Root: utxo, State: *out}, Witness: wits[index]})
				if out.OS.Out_Z == nil {
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
		currency := strings.ToUpper(reception.Currency)
		if amount, ok := amounts[currency]; ok && amount.Cmp(reception.Value) >= 0 {

			if IsPk(reception.Addr) {
				pk := reception.Addr.ToUint512()
				pkr := CreatePkr(&pk, 1)
				Outs = append(Outs, GOut{PKr: pkr, Asset: assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
					Value:    utils.U256(*reception.Value),
				}}})
			} else {
				Outs = append(Outs, GOut{PKr: reception.Addr, Asset: assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte(currency), 32)).HashToUint256(),
					Value:    utils.U256(*reception.Value),
				}}})
			}

			amount.Sub(amount, reception.Value)
			if amount.Sign() == 0 {
				delete(amounts, currency)
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
