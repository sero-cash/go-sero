package light

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/seroparam"

	"github.com/sero-cash/go-sero/log"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-sero/zero/light/light_types"

	"github.com/sero-cash/go-sero/zero/light/light_ref"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type SRI struct {
}

var SRI_Inst = SRI{}

func GetOut(root *keys.Uint256, num uint64) (out *localdb.RootState) {
	rs := localdb.GetRoot(light_ref.Ref_inst.Bc.GetDB(), root)
	if rs != nil {
		return rs
	} else {
		zst := light_ref.Ref_inst.GetState()
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

func (self *SRI) GetBlocksInfo(start uint64, count uint64) (blocks []light_types.Block, e error) {
	stable_num := light_ref.Ref_inst.GetDelayedNum(seroparam.DefaultConfirmedBlock())
	if start <= stable_num {
		if stable_num-start+1 < count {
			count = stable_num - start + 1
		}
		for i := uint64(0); i < count; i++ {
			num := start + i
			chain_block := light_ref.Ref_inst.Bc.GetBlockByNumber(num)
			hash := chain_block.Hash()
			local_block := localdb.GetBlock(light_ref.Ref_inst.Bc.GetDB(), num, hash.HashToUint256())
			if local_block != nil {
				block := light_types.Block{}
				block.Hash = *hash.HashToUint256()
				block.Num = hexutil.Uint64(num)
				for _, k := range local_block.Dels {
					block.Nils = append(block.Nils, k)
				}
				for _, k := range local_block.Roots {
					if out := GetOut(&k, num); out == nil {
						log.Error("GetBlocksInfo ERROR", "num", num, "root", k)
					} else {
						block.Outs = append(block.Outs, light_types.Out{k, *out})
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

func (self *SRI) GetAnchor(roots []keys.Uint256) (wits []light_types.Witness, e error) {
	state := light_ref.Ref_inst.GetState()
	if state != nil {
		for _, root := range roots {
			wit := light_types.Witness{}
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
