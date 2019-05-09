package state1

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/lstate"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/log"

	"time"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/core/types"
)

func state1_file_name(num uint64, hash *common.Hash) (ret string) {
	ret = fmt.Sprintf("%010d.%s", num, hexutil.Encode(hash[:])[3:])
	return
}

const delay_block_count = 6

func (self *State1) Parse(last_chose uint64) (chose uint64) {
	bc := lstate.BC()

	var current_header *types.Header
	current_header = bc.GetCurrenHeader()
	tks := bc.GetTks()

	if current_chose := bc.CashChose(); current_chose != nil {
		chose := current_chose.Load().(uint64)
		if chose > 0 {
			delay := uint64(delay_block_count)
			if current_header.Number.Uint64() < delay_block_count {
				delay = 0
			} else {
				dist := current_header.Number.Uint64() - delay
				if delay > dist {
					delay = dist
				}
			}
			if (current_header.Number.Uint64() - chose) < delay {
				delay = (current_header.Number.Uint64() - chose)
			}
			for i := uint64(0); i < delay; i++ {
				parent_hash := current_header.ParentHash
				current_header = bc.GetHeader(&parent_hash)
				if current_header == nil {
					return last_chose
				}
			}
		}
	}

	hash := current_header.Hash()

	chose = current_header.Number.Uint64()

	progress := utils.NewProgress("STATE1_PROCESS : ", current_header.Number.Uint64())

	need_load := []*types.Header{}

	self.begin(&hash, tks)

	for {
		current_hash := current_header.Hash()
		current_num := current_header.Number.Uint64()
		if need_parse, err := self.needParse(current_num, &current_hash); err != nil {
			time.Sleep(1000 * 1000 * 1000 * 10)
			return
		} else {
			if need_parse {
				need_load = append(need_load, current_header)
				parent_hash := current_header.ParentHash
				current_header = bc.GetHeader(&parent_hash)
				if current_header == nil {
					break
				} else {
					hash := current_header.Hash()
					if hash != parent_hash {
						log.Error(
							"current.hash not equal the pre.parent_hash : ",
							"current.h", hexutil.Encode(hash[:])[:8],
							"pre.p_h",
							hexutil.Encode(parent_hash[:])[:8],
						)
						panic("parse block error")
					}
					continue
				}
			} else {
				break
			}
		}
	}

	parse_count := 0
	for i := len(need_load) - 1; parse_count < 2000 && i >= 0; i-- {

		header := need_load[i]
		current_num := header.Number.Uint64()
		current_hash := header.Hash()

		t := utils.TR_enter(fmt.Sprintf("PARSE_BLOCK_CHAIN----NewState(num=%v)", current_num))

		block := localdb.GetBlock(bc.GetDB(), current_num, current_hash.HashToUint256())

		if block == nil {
			temp_state := bc.NewState(&current_hash)
			if temp_state == nil {
				panic(fmt.Sprintf("new zstate error: %v:%v !", current_num, current_hash))
			} else {
				log.Debug("STATE1_PARSE GO BACK TO STATE: ", "num", current_num, "hash", current_hash)
			}
			block = &localdb.Block{}
			block.Pkgs = temp_state.Pkgs.GetPkgHashes()
			block.Roots = temp_state.State.GetBlockRoots()
			block.Dels = temp_state.State.GetBlockDels()
		}

		t.Renter("PARSE_BLOCK_CHAIN----LoadState")

		self.update(&header.ParentHash, current_num, &current_hash, block)

		t.Renter("PARSE_BLOCK_CHAIN----Finalize")

		if i < 30 {
			self.save()
		} else {
			if parse_count%2000 == 0 {
				self.save()
			}
		}
		td := t.Leave()
		progress.Tick(current_num, "len", len(block.Roots), "d", td)
		parse_count++
	}

	self.finalize()

	return

}
