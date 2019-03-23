package lstate

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/sero-cash/go-sero/zero/localdb"

	"github.com/sero-cash/go-sero/serodb"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/log"

	"time"

	"os"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

type BlockChain interface {
	GetCurrenHeader() *types.Header
	GetHeader(hash *common.Hash) *types.Header
	NewState(hash *common.Hash) *zstate.ZState
	GetTks() []keys.Uint512
	CashChose() *atomic.Value
	GetDB() serodb.Database
}

func state1_file_name(num uint64, hash *common.Hash) (ret string) {
	ret = fmt.Sprintf("%010d.%s", num, hexutil.Encode(hash[:])[3:])
	return
}

var current_state1 *State
var current_bc BlockChain

func CurrentState1() *State {
	return current_state1
}

func BC() BlockChain {
	return current_bc
}

func Run(bc BlockChain) {
	current_bc = bc
	go run(bc)
	for current_state1 != nil {
		time.Sleep(time.Second * 1)
	}
}

const delay_block_count = 6

func parse_block_chain(bc BlockChain, last_cmd_count int) (current_cm_count int, e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("parse block chain error : ", "number", bc.GetCurrenHeader().Number, "recover", r)
			debug.PrintStack()
			e = errors.Errorf("parse block chain error %v %v", bc.GetCurrenHeader().Number, r)
		}
	}()

	var current_header *types.Header
	current_header = bc.GetCurrenHeader()
	tks := bc.GetTks()

	current_chose := bc.CashChose()
	if current_chose != nil {
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
					return
				}
			}
		}
	}

	hash := current_header.Hash()
	var stz = bc.NewState(&hash)

	chose := current_header.Number.Uint64()

	progress := utils.NewProgress("STATE1_PROCESS : ", current_header.Number.Uint64())

	need_load := []*types.Header{}

	for {
		current_hash := current_header.Hash()
		current_name := state1_file_name(current_header.Number.Uint64(), &current_hash)
		current_file := zconfig.State1_file(current_name)
		var is_exist bool

		if _, err := os.Stat(current_file); err != nil {
			if os.IsNotExist(err) {
				is_exist = false
			} else {
				time.Sleep(1000 * 1000 * 1000 * 10)
				break
			}
		} else {
			is_exist = true
		}

		if !is_exist {
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

	var st1 *State
	parse_count := 0
	for i := len(need_load) - 1; i >= 0; i-- {

		header := need_load[i]
		current_num := header.Number.Uint64()
		current_hash := header.Hash()
		saved_name := state1_file_name(current_num, &current_hash)
		var load_name string
		if current_num == 0 {
			load_name = ""
		} else {
			parent_num := current_num - 1
			parent_hash := header.ParentHash
			load_name = state1_file_name(parent_num, &parent_hash)
		}

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
			block.Pkgs = temp_state.Pkgs.GetBlockDetails()
			block.Dels = temp_state.State.Block.Dels
			block.Roots = temp_state.State.Block.Roots
		}

		//state := bc.NewState(&current_hash)

		current_cm_count += len(block.Roots)

		t.Renter("PARSE_BLOCK_CHAIN----LoadState")

		if st1 == nil {
			s1 := LoadState(stz, load_name)
			st1 = &s1
		}

		commitment_len := len(st1.State.State.Block.Roots)
		t.Renter(fmt.Sprintf("PARSE_BLOCK_CHAIN----UpdateWiteness(count=%d)", commitment_len))
		st1.UpdateWitness(tks, current_num, block)
		current_state1 = st1

		t.Renter("PARSE_BLOCK_CHAIN----Finalize")
		if parse_count%2000 == 0 {
			st1.Finalize(saved_name, current_num)
			st1 = nil
		} else {
			if i < 30 {
				st1.Finalize(saved_name, current_num)
				st1 = nil
			}
		}
		td := t.Leave()
		progress.Tick(current_num, "len", commitment_len, "d", td)
		parse_count++
	}

	if current_state1 == nil {
		current_num := current_header.Number.Uint64()
		current_hash := current_header.Hash()
		state_name := state1_file_name(current_num, &current_hash)
		st := bc.NewState(&current_hash)
		st1 := LoadState(st, state_name)
		current_state1 = &st1
	}

	cashChose := bc.CashChose()
	cashChose.Store(chose)
	return current_cm_count, nil

}

func run(bc BlockChain) {
	cmd_count := 2
	for {
		cmd_count, _ := parse_block_chain(bc, cmd_count)
		if cmd_count <= 1 {
			time.Sleep(1000 * 1000 * 1000 * 8)
		} else {
			time.Sleep(1000 * 1000 * 10)
		}
	}
}
