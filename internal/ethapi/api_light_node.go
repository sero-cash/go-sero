package ethapi

import (
	"context"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/light/light_node"
	"fmt"
)

type PublicLightNodeApi struct {
	b Backend
}

func (plna PublicLightNodeApi) GetOutsByPKr(ctx context.Context, addresses [] *MixAdrress, start, end uint64) (outBlockResp light_node.BlockOutResp, e error) {

	pkrs := []keys.PKr{}
	for _,address := range addresses{
		addr := *address
		if len(addr) == 96 {
			var pkr keys.PKr
			copy(pkr[:], addr[:])
			pkrs = append(pkrs,pkr)
		} else {
			return outBlockResp, fmt.Errorf("address is invalid")
		}
	}

	return plna.b.GetOutByPKr(pkrs, start, end)
}

func (plna PublicLightNodeApi) CheckNil(Nils []keys.Uint256, start ,end uint64) (delNils []light_node.BlockDelNil, e error) {

	return plna.b.CheckNil(Nils, start,end)
}
