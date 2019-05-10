package state1

import (
	"fmt"

	"github.com/sero-cash/go-sero/zero/light/light_ref"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/log"

	"time"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
)

func state1_file_name(num uint64, hash *common.Hash) (ret string) {
	ret = fmt.Sprintf("%010d.%s", num, hexutil.Encode(hash[:])[3:])
	return
}

const delay_block_count = 6

func (self *State1) Parse(last_chose uint64) (chose uint64) {
	bc := light_ref.Ref_inst.Bc
	tks := bc.GetTks()

	if self.next_num == 0 {
		current_header := bc.GetCurrenHeader()
		for {
			current_hash := current_header.Hash()
			current_num := current_header.Number.Uint64()
			if need_parse, err := self.needParse(current_num, &current_hash); err != nil {
				time.Sleep(1000 * 1000 * 1000 * 10)
				return
			} else {
				if need_parse {
					self.next_num = current_header.Number.Uint64()
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
	}

	chose = light_ref.Ref_inst.GetDelayedNum(delay_block_count)
	if self.next_num > chose {
		return last_chose
	}

	chose_header := bc.GetHeaderByNumber(chose)
	hash := chose_header.Hash()
	self.begin(&hash, tks)

	parse_count := 0
	for parse_count < 2000 && self.next_num <= chose {
		header := bc.GetHeaderByNumber(self.next_num)
		hash := header.Hash()

		block := localdb.GetBlock(bc.GetDB(), self.next_num, hash.HashToUint256())
		if block == nil {
			temp_state := bc.NewState(&hash)
			if temp_state == nil {
				panic(fmt.Sprintf("new zstate error: %v:%v !", self.next_num, hash))
			} else {
				log.Debug("STATE1_PARSE GO BACK TO STATE: ", "num", self.next_num, "hash", hash)
			}
			block = &localdb.Block{}
			block.Pkgs = temp_state.Pkgs.GetPkgHashes()
			block.Roots = temp_state.State.GetBlockRoots()
			block.Dels = temp_state.State.GetBlockDels()
		}

		self.update(&header.ParentHash, self.next_num, &hash, block)

		self.next_num++
		parse_count++
	}

	if parse_count > 0 {
		self.save()
		log.Info("STATE1 PARSE", "t", chose, "c", self.next_num-1)
	}

	self.finalize()

	return

}
