package accounts

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/wallet/lstate/lstate_types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Out struct {
	OS localdb.OutState
}

func (self *DB) getStateOut(tk *keys.Uint512, num uint64, root *keys.Uint256, os *localdb.OutState) (ret *lstate_types.OutState) {
	if os.IsO() {
		out_o := os.Out_O
		if out_o.Asset.Tkn == nil && out_o.Asset.Tkt == nil {
			return
		}
		no_tkn_value := false
		if out_o.Asset.Tkn != nil {
			if out_o.Asset.Tkn.Value.Cmp(&utils.U256_0) <= 0 {
				no_tkn_value = true
			}
		} else {
			no_tkn_value = true
		}
		no_tkt_value := false
		if out_o.Asset.Tkt != nil {
			if out_o.Asset.Tkt.Value == keys.Empty_Uint256 {
				no_tkt_value = true
			}
		} else {
			no_tkt_value = true
		}

		if no_tkt_value && no_tkn_value {
			return
		}

		if out_o.Addr == (keys.PKr{}) {
			return
		}

		if succ := keys.IsMyPKr(tk, &out_o.Addr); succ {
			out_z := &stx.Out_Z{}
			{
				desc_info := cpt.EncOutputInfo{}

				asset := os.Out_O.Asset.ToFlatAsset()
				desc_info.Tkn_currency = asset.Tkn.Currency
				desc_info.Tkn_value = asset.Tkn.Value.ToUint256()
				desc_info.Tkt_category = asset.Tkt.Category
				desc_info.Tkt_value = asset.Tkt.Value
				desc_info.Rsk = os.ToIndexRsk()
				desc_info.Memo = os.Out_O.Memo
				cpt.EncOutput(&desc_info)
				out_z = &stx.Out_Z{}
				out_z.PKr = os.Out_O.Addr
				out_z.EInfo = desc_info.Einfo
				out_z.OutCM = *os.ToOutCM()
			}
			ret = &lstate_types.OutState{}
			ret.Root = *root
			ret.RootCM = *os.ToRootCM()
			ret.Tk = *tk
			ret.Out_O = *os.Out_O
			ret.OutIndex = os.Index
			ret.Out_Z = out_z
			ret.Z = false
			ret.Trace = cpt.GenTil(tk, os.ToRootCM())
			ret.Num = num
			return
		} else {
			return
		}
	} else {
		if succ := keys.IsMyPKr(tk, &os.Out_Z.PKr); succ {
			key, flag := keys.FetchKey(tk, &os.Out_Z.RPK)

			info_desc := cpt.InfoDesc{}
			info_desc.Key = key
			info_desc.Flag = flag
			info_desc.Einfo = os.Out_Z.EInfo

			cpt.DecOutput(&info_desc)

			if e := stx.ConfirmOut_Z(&info_desc, os.Out_Z); e == nil {
				ret = &lstate_types.OutState{}
				ret.Out_O.Addr = os.Out_Z.PKr
				ret.Out_O.Asset = assets.NewAsset(
					&assets.Token{
						info_desc.Tkn_currency,
						utils.NewU256_ByKey(&info_desc.Tkn_value),
					},
					&assets.Ticket{
						info_desc.Tkt_category,
						info_desc.Tkt_value,
					},
				)
				ret.Out_O.Memo = info_desc.Memo
				ret.Out_Z = os.Out_Z.Clone().ToRef()
				ret.Root = *root
				ret.RootCM = *os.ToRootCM()
				ret.Tk = *tk
				ret.OutIndex = os.Index
				ret.Z = true
				ret.Trace = cpt.GenTil(tk, os.ToRootCM())
				ret.Num = num
				return
			} else {
				log.Error("My out_z confirm error", "root", hexutil.Encode(os.ToRootCM()[:]))
				return
			}
		} else {
			return
		}
	}
}

type TkRoot struct {
	Tk   keys.Uint512
	Root keys.Uint256
}

func (self *TkRoot) Bytes() (ret []byte) {
	return append(self.Tk[:], self.Root[:]...)
}

func NewTkRoot(bytes []byte) (ret TkRoot) {
	copy(ret.Tk[:], bytes[:64])
	copy(ret.Root[:], bytes[64:])
	return
}

const ROOT_OUT_KEY = "ROOT$OUT$KEY-"
const NIL_ROOT_KEY = "NIL$ROOT$KEY-"
const TKROOT_ROOT_KEY = "TKROOT$ROOT$KEY-"

func (self *DB) GetOut(root *keys.Uint256) (out lstate_types.OutState, e error) {
	if v, err := self.db.Get(Bytes2Key(ROOT_OUT_KEY, root[:]), nil); err != nil {
		e = err
		return
	} else {
		if err := rlp.DecodeBytes(v, &out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
func (self *DB) GetOuts(tk *keys.Uint512) (outs []*lstate_types.OutState, e error) {
	iter := self.db.NewIterator(util.BytesPrefix(Bytes2Key(TKROOT_ROOT_KEY, tk[:])), nil)
	defer iter.Release()
	for iter.Next() {
		v := iter.Value()
		if vo, err := self.db.Get(Bytes2Key(ROOT_OUT_KEY, v), nil); err != nil {
			e = err
			return
		} else {
			var out lstate_types.OutState
			if err := rlp.DecodeBytes(vo, &out); err != nil {
				panic(err)
			} else {
				outs = append(outs, &out)
			}
		}
	}
	return
}

func (self *DB) AddOut(batch *leveldb.Batch, a *Account, num uint64, root *keys.Uint256, os *localdb.OutState) (ret bool) {
	if out := self.getStateOut(&a.Tk, num, root, os); out != nil {
		if out.Out_O.Asset.Tkn != nil {
			a.AddToken(out.Out_O.Asset.Tkn)
		}
		if out.Out_O.Asset.Tkt != nil {
			a.AddTicket(out.Out_O.Asset.Tkt)
		}
		if v, err := rlp.EncodeToBytes(out); err != nil {
			panic(err)
		} else {
			batch.Put(Bytes2Key(ROOT_OUT_KEY, root[:]), v[:])
			batch.Put(Bytes2Key(NIL_ROOT_KEY, out.Trace[:]), root[:])
			batch.Put(Bytes2Key(NIL_ROOT_KEY, root[:]), root[:])
			tkroot := TkRoot{a.Tk, *root}
			batch.Put(Bytes2Key(TKROOT_ROOT_KEY, tkroot.Bytes()), root[:])
			return true
		}
	} else {
		return false
	}
}

func (self *DB) AddNil(batch *leveldb.Batch, del *keys.Uint256) (tkroot TkRoot, out lstate_types.OutState, e error) {
	if root, err := self.db.Get(Bytes2Key(NIL_ROOT_KEY, del[:]), nil); err != nil {
		e = err
	} else {
		if v, err := self.db.Get(Bytes2Key(ROOT_OUT_KEY, root), nil); err != nil {
			panic(err)
		} else {
			if err := rlp.DecodeBytes(v, &out); err != nil {
				panic(err)
			} else {
				if _, err := self.db.Get(Bytes2Key(NIL_ROOT_KEY, del[:]), nil); err != nil {
					panic(err)
				} else {
					batch.Delete(Bytes2Key(NIL_ROOT_KEY, del[:]))
				}
				tkroot = TkRoot{out.Tk, out.Root}
				if _, err := self.db.Get(Bytes2Key(TKROOT_ROOT_KEY, tkroot.Bytes()), nil); err != nil {
					//panic(err)
					e = errors.New("AddNil already been deleted")
				} else {
					batch.Delete(Bytes2Key(TKROOT_ROOT_KEY, tkroot.Bytes()))
				}
			}
		}
	}
	return
}
