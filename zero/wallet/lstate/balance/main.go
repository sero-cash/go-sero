package balance

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/wallet/lstate/balance/accounts"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/zconfig"
	"github.com/syndtr/goleveldb/leveldb"
)

func (self *Balance) MakesureEnv() {
	if !zconfig.IsDirExists(zconfig.State2_dir()) {
		zconfig.Init_State2()
		if self.db != nil {
			self.db.DB().Close()
			self.db = nil
		}
	}

	if self.db == nil {
		db := accounts.NewDB(zconfig.State2_dir())
		self.db = &db
	}
}

func GetOut(root *keys.Uint256) (src *localdb.OutState) {
	db := txtool.Ref_inst.Bc.GetDB()
	rt := localdb.GetRoot(db, root)
	if rt == nil {
		zst := txtool.Ref_inst.GetState()
		os := zst.State.GetOut(root)
		return os
	} else {
		src = &rt.OS
		return
	}
}

func (self *Balance) Parse() (num uint64) {

	self.MakesureEnv()

	if txtool.Ref_inst.Bc == nil || !txtool.Ref_inst.Bc.IsValid() {
		return 0
	}

	tks := txtool.Ref_inst.Bc.GetTks()

	if len(tks) == 0 {
		return uint64(0)
	}

	next_num := uint64(0)
	amap := make(map[keys.Uint512]*accounts.Account)
	for _, tk := range tks {
		account := self.db.GetAccount(&tk)
		at := txtool.Ref_inst.Bc.GetTkAt(&tk)
		if account.NextNum < at {
			account.NextNum = at
		}
		amap[tk] = &account
		if next_num == 0 {
			next_num = account.NextNum
		} else {
			if next_num > account.NextNum {
				next_num = account.NextNum
			}
		}
	}

	target_num := txtool.Ref_inst.GetDelayedNum(seroparam.DefaultConfirmedBlock())

	i := 0
	for ; (next_num <= target_num) && (i < 2000); i++ {
		batch := leveldb.Batch{}

		next_header := txtool.Ref_inst.Bc.GetHeaderByNumber(next_num)
		next_hash := next_header.Hash()
		block := localdb.GetBlock(txtool.Ref_inst.Bc.GetDB(), next_num, next_hash.HashToUint256())
		if block == nil {
			temp_state := txtool.Ref_inst.Bc.NewState(&next_hash)
			if temp_state == nil {
				panic(fmt.Sprintf("new zstate error: %v:%v !", next_num, next_hash))
			} else {
				log.Debug("STATE1_PARSE GO BACK TO STATE: ", "num", next_num, "hash", next_hash)
			}
			block = &localdb.Block{}
			block.Pkgs = temp_state.Pkgs.GetPkgHashes()
			block.Roots = temp_state.State.GetBlockRoots()
			block.Dels = temp_state.State.GetBlockDels()
		}

		for _, del := range block.Dels {
			if tkroot, out, err := self.db.AddNil(&batch, &del); err == nil {
				if account, ok := amap[tkroot.Tk]; ok {
					if out.Out_O.Asset.Tkn != nil {
						account.DelToken(out.Out_O.Asset.Tkn)
					}
					if out.Out_O.Asset.Tkt != nil {
						account.DelTicket(out.Out_O.Asset.Tkt)
					}
				}
			}
		}

		for _, root := range block.Roots {
			os := GetOut(&root)
			if os == nil {
				panic("BALANCE parse but can not find root -> out")
			} else {
			}
			for _, account := range amap {
				if account.NextNum == next_num {
					if self.db.AddOut(&batch, account, next_num, &root, os) {
						break
					}
				} else {
					continue
				}
			}
		}
		for _, hash := range block.Pkgs {
			pg := localdb.GetPkg(txtool.Ref_inst.Bc.GetDB(), &hash)
			if pg == nil {
				panic("BALANCE parse but can not find hash -> pkg")
			} else {
				for _, account := range amap {
					if account.NextNum == next_num {
						if self.db.AddPkg(account, pg) {
							break
						}
					} else {
						continue
					}
				}
			}
		}

		for _, account := range amap {
			if account.NextNum == next_num {
				account.Next()
				self.db.SetAccount(&batch, account)
			} else {
				continue
			}
		}

		next_num++

		if batch.Len() > 0 {
			if err := self.db.DB().Write(&batch, nil); err != nil {
				panic(err)
			}
		}
	}
	if i > 1 {
		log.Info("BALANCE PARSE", "t", target_num, "c", next_num-1)
	}

	return uint64(i)
}

func (self *Balance) update(tk *keys.Uint512, num uint64, block *localdb.Block) {
	return
}
