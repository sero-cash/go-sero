package txs

import (
	"errors"

	"github.com/sero-cash/go-sero/zero/wallet/lstate"
	"github.com/sero-cash/go-sero/zero/wallet/lstate/lstate_types"

	"github.com/sero-cash/go-sero/zero/txs/pkg"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
)

func WatchPkg(id *keys.Uint256, key *keys.Uint256) (ret pkg.Pkg_O, pkr keys.PKr, e error) {
	st1 := lstate.CurrentLState()
	if st1 == nil {
		e = errors.New("Watch Pkg but lstate is nil")
		return
	}
	pg := st1.CurrentZState().Pkgs.GetPkgById(id)
	if pg == nil || pg.Closed {
		e = errors.New("Watch Pkg but has been closed")
		return
	}
	pkr = pg.Pack.PKr
	ret, e = pkg.DePkg(key, &pg.Pack.Pkg)
	return
}

func GetOuts(tk *keys.Uint512) (outs []*lstate_types.OutState, e error) {
	st1 := lstate.CurrentLState()
	if st1 == nil {
		e = errors.New("Get outs but lstate is nil")
		return
	}
	return st1.GetOuts(tk)
}

func GetRoots(tk *keys.Uint512, costTkns map[keys.Uint256]utils.U256, costTkts map[keys.Uint256][]keys.Uint256) (roots []keys.Uint256, tknMap map[keys.Uint256]utils.U256, tktMap map[keys.Uint256][]keys.Uint256, e error) {
	tknMap = make(map[keys.Uint256]utils.U256)
	tktMap = make(map[keys.Uint256][]keys.Uint256)
	if outs, err := GetOuts(tk); err != nil {
		e = err
		return
	} else {
		for cy, value := range costTkns {
			tknRoots, amount, tkts, err := GetTknRoots(outs, &value, &cy)
			if err != nil {
				e = err
				return
			} else {
				tknMap[cy] = amount
				roots = append(roots, tknRoots...)
				for catg, value := range tkts {
					tktMap[catg] = append(tktMap[catg], value...)
				}
			}
		}

		for catg, value := range costTkts {
			if _, ok := tktMap[catg]; ok {
				for _, v := range value {
					tktMap[catg] = uint256Remove(tktMap[catg], v)
				}
				if len(tktMap[catg]) == 0 {
					delete(tktMap, catg)
				}
			}
		}
		for catg, value := range costTkts {
			tktRoots, tkns, err := GeTktRoots(outs, &catg, value, roots)
			if err != nil {
				e = err
				return
			} else {
				roots = append(roots, tktRoots...)
				for cy, value := range tkns {
					if _, ok := tknMap[cy]; ok {
						amount := tknMap[cy]
						amount.AddU(&value)
						tknMap[cy] = amount
					} else {
						tknMap[cy] = value
					}
				}
			}
		}
		for cy, value := range costTkns {
			if balance, ok := tknMap[cy]; ok {
				balance.SubU(&value)
				if balance.Cmp(&utils.U256_0) > 0 {
					tknMap[cy] = balance
				} else {
					delete(tknMap, cy)
				}

			}
		}

	}
	return

}

func GetTknRoots(outs []*lstate_types.OutState, v *utils.U256, currency *keys.Uint256) (roots []keys.Uint256, amount utils.U256, tkts map[keys.Uint256][]keys.Uint256, e error) {
	tkts = make(map[keys.Uint256][]keys.Uint256)
	value := v.ToI256()
	for _, out := range outs {
		root := out.Root
		if out.Out_O.Asset.Tkn != nil {
			if out.Out_O.Asset.Tkn.Currency == *currency {
				roots = append(roots, root)
				amount.AddU(&out.Out_O.Asset.Tkn.Value)
				out_o := out.Out_O
				if out_o.Asset.Tkt != nil {
					if ts, ok := tkts[out_o.Asset.Tkt.Category]; ok {
						ts = append(ts, out_o.Asset.Tkt.Value)
						tkts[out_o.Asset.Tkt.Category] = ts
					} else {
						tkts[out_o.Asset.Tkt.Category] = []keys.Uint256{out_o.Asset.Tkt.Value}
					}
				}
				if value.Cmp(out_o.Asset.Tkn.Value.ToI256().ToRef()) <= 0 {
					value = utils.NewI256(0)
					break
				} else {
					value.SubU(&out_o.Asset.Tkn.Value)
				}
			} else {
			}
		}
	}
	if value.Cmp(&utils.I256_0) == 0 {
		return
	} else {
		e = errors.New("can not find enough outs")
		return
	}

}

func GeTktRoots(outs []*lstate_types.OutState, categroy *keys.Uint256, tkts []keys.Uint256, exits []keys.Uint256) (roots []keys.Uint256, tkns map[keys.Uint256]utils.U256, e error) {
	tkns = map[keys.Uint256]utils.U256{}
	tktSize := len(tkts)
	for _, out := range outs {
		root := out.Root
		if categroy != nil && tktSize > 0 {
			if out.Out_O.Asset.Tkt != nil {
				if out.Out_O.Asset.Tkt.Category == *categroy {
					if uint256Contains(tkts, out.Out_O.Asset.Tkt.Value) {
						if !uint256Contains(exits, root) {
							if out.Out_O.Asset.Tkn != nil {
								if tkn, ok := tkns[out.Out_O.Asset.Tkn.Currency]; ok {
									tkn.AddU(&out.Out_O.Asset.Tkn.Value)
									tkns[out.Out_O.Asset.Tkn.Currency] = tkn
								} else {
									tkns[out.Out_O.Asset.Tkn.Currency] = out.Out_O.Asset.Tkn.Value
								}

							}
							roots = append(roots, root)
						}
						tktSize--
					}
				} else {

				}
			}
		}
		if tktSize == 0 {
			break
		}
	}
	if tktSize == 0 {
		return
	} else {
		e = errors.New("can not find enough ticket outs")
		return
	}

}

func uint256Contains(arrs []keys.Uint256, item keys.Uint256) bool {
	for _, a := range arrs {
		if a == item {
			return true
		}
	}
	return false

}

func uint256Remove(slice []keys.Uint256, elem keys.Uint256) []keys.Uint256 {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == elem {
			slice = append(slice[:i], slice[i+1:]...)
			return uint256Remove(slice, elem)
			break
		}
	}
	return slice
}
