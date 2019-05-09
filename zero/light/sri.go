package light

import (
	"fmt"

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

func (self *SRI) GetBlocksInfo(start uint64, count uint64) (blocks []light_types.Block, e error) {
	stable_num := light_ref.Ref_inst.GetDelayedNum(12)
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
					root := localdb.GetRoot(light_ref.Ref_inst.Bc.GetDB(), &k)
					if root != nil {
						block.Outs = append(block.Outs, light_types.Out{k, *root})
					} else {
						e = fmt.Errorf("GetBlocksInfo.GetOut Failed, num: %v root: %v", num, k)
						return
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
			out := localdb.GetRoot(light_ref.Ref_inst.Bc.GetDB(), &root)
			if out == nil {
				e = errors.New("GetAnchor use root but out is nil !!!")
				return
			}
			pos, paths, anchor := state.State.MTree.GetPaths(*out.OS.RootCM)
			wit.Pos = hexutil.Uint64(pos)
			wit.Paths = paths
			wit.Anchor = anchor
			wits = append(wits, wit)
		}
		return
	} else {
		e = errors.New("State is nil")
		return
	}
	return
}
