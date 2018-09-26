package state1

import (
	"fmt"

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
	NewState(hash *common.Hash) *zstate.State
	GetTks() []keys.Uint512
}

func state1_file_name(num uint64, hash *common.Hash) (ret string) {
	ret = fmt.Sprintf("%010d.%s", num, hexutil.Encode(hash[:])[3:])
	return
}

var current_state1 *State1

func CurrentState1() *State1 {
	return current_state1
}

func Run(bc BlockChain) {
	go run(bc)
}
func run(bc BlockChain) {

	for {
		var current_header *types.Header
		current_header = bc.GetCurrenHeader()
		tks := bc.GetTks()

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
					continue
				}
			} else {
				break
			}
		}

		var st1 *State1
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
			state := bc.NewState(&current_hash)
			if len(load_name) > 0 {
				log.Info("STATE1 REPARSE BLOCK : ", "loadname", load_name[5:17], "savename", saved_name[5:17], "cmnum", len(state.Block.Commitments))
			}
			if st1 == nil {
				s1 := LoadState1(&state.State0, load_name)
				st1 = &s1
			} else {
				st1.State0 = &state.State0
			}
			st1.UpdateWitness(tks)
			current_state1 = st1
			if i < 30 {
				st1.Finalize(saved_name)
				st1 = nil
			}
		}

		if current_state1 == nil {
			current_num := current_header.Number.Uint64()
			current_hash := current_header.Hash()
			state_name := state1_file_name(current_num, &current_hash)
			st := bc.NewState(&current_hash)
			st1 := LoadState1(&st.State0, state_name)
			current_state1 = &st1
		}

		if len(need_load) > 0 {
			log.Info("End parse block chain for state1 : ", "out_num", len(current_state1.G2wouts))
		}

		if len(need_load) == 0 {
			time.Sleep(1000 * 1000 * 1000 * 5)
		} else {
			time.Sleep(1000 * 1000 * 1000)
		}
	}
}
