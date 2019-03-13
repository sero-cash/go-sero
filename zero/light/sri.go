package light

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type SRI struct {
}

var SRI_Inst = SRI{}

func (self *SRI) GetBlocksInfo(start uint64, count uint64) (blocks []Block, e error) {
	stable_num := Light_inst.GetDelayedNum(32)
	if start <= stable_num {
		if stable_num-start+1 < count {
			count = stable_num - start + 1
		}
		for i := uint64(0); i < count; i++ {
			num := start + i
			chain_block := Light_inst.Bc.GetBlockByNumber(num)
			hash := chain_block.Hash()
			local_block := localdb.GetBlock(Light_inst.Bc.GetDB(), num, hash.HashToUint256())
			if local_block != nil {
				block := Block{}
				block.Num = num
				for _, k := range local_block.Dels {
					block.Nils = append(block.Nils, k)
				}
				for _, k := range local_block.Roots {
					out := localdb.GetOut(Light_inst.Bc.GetDB(), &k)
					if out != nil {
						block.Outs = append(block.Outs)
					} else {
						e = fmt.Errorf("GetBlocksInfo.GetOut Failed, num: %v root: %v", num, k)
						return
					}
				}
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

func (self *SRI) GetAnchor(roots []keys.Uint256) (wits []Witness, e error) {
	return
}

func (self *SRI) CommitTx(tx *GTx) (e error) {
	return
}
