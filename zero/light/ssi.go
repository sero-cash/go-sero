package light

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/light/light_issi"
	"github.com/sero-cash/go-sero/zero/light/light_ref"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type SSI struct {
}

var SSI_Inst = SSI{}

func (self *SSI) GetBlocksInfo(start uint64, count uint64) (blocks []light_issi.Block, e error) {

	if bs, err := SRI_Inst.GetBlocksInfo(start, count); err != nil {
		e = err
		return
	} else {
		for _, b := range bs {
			block := light_issi.Block{}
			block.Num = b.Num
			block.Nils = b.Nils
			for _, o := range b.Outs {
				block.Outs = append(
					block.Outs,
					light_issi.Out{
						o.Root,
						*o.State.OS.ToPKr(),
					},
				)
			}
			blocks = append(blocks, block)
		}
	}

	return
}

func (self *SSI) Detail(roots []keys.Uint256, skr *keys.PKr) (douts []light_types.DOut, e error) {

	outs := []light_types.Out{}
	for _, r := range roots {
		if root := localdb.GetRoot(light_ref.Ref_inst.Bc.GetDB(), &r); root == nil {
			e = fmt.Errorf("SSI Detail Error for root %v", r)
			return
		} else {
			outs = append(outs, light_types.Out{r, *root})
		}
	}
	douts = SLI_Inst.DecOuts(outs, skr)

	return
}

var txMap sync.Map

func (self *SSI) GenTx(param *light_issi.GenTxParam) (hash keys.Uint256, e error) {
	p := light_types.GenTxParam{}
	p.Gas = param.Gas
	p.GasPrice = *big.NewInt(0).SetUint64(param.GasPrice)
	p.From = param.From
	p.Outs = param.Outs

	roots := []keys.Uint256{}
	outs := []light_types.Out{}

	for _, in := range param.Ins {
		roots = append(roots, in.Root)
		if root := localdb.GetRoot(light_ref.Ref_inst.Bc.GetDB(), &in.Root); root == nil {
			e = fmt.Errorf("SSI GenTx Error for root %v", in.Root)
			return
		} else {
			outs = append(outs, light_types.Out{in.Root, *root})
		}
	}

	wits := []light_types.Witness{}

	if wits, e = SRI_Inst.GetAnchor(roots); e != nil {
		return
	}

	for i := 0; i < len(wits); i++ {
		in := light_types.GIn{}
		in.SKr = param.Ins[i].SKr
		in.Out = outs[i]
		in.Witness = wits[i]
	}

	if gtx, err := SLI_Inst.GenTx(&p); err != nil {
		e = err
		return
	} else {
		hash = gtx.Tx.ToHash()
		txMap.Store(&hash, &gtx)
	}

	return
}

func (self *SSI) GetTx(txhash *keys.Uint256) (tx *light_types.GTx, e error) {
	if ld, ok := txMap.Load(txhash); !ok {
		e = fmt.Errorf("SSI GetTx Failed : %v", txhash)
	} else {
		if ld == nil {
			e = fmt.Errorf("SSI GetTx Nil : %v", txhash)
		} else {
			tx = ld.(*light_types.GTx)
		}
	}
	return
}
