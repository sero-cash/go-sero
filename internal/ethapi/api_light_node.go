package ethapi

import (
	"context"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/wallet/light"
	"fmt"
)

type PublicLightNodeApi struct {
	b Backend
}

func (plna PublicLightNodeApi) GetOutsByPKr(ctx context.Context, addresses [] *MixAdrress, start, end uint64) (outBlockResp light.BlockOutResp, e error) {

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

func (plna PublicLightNodeApi) CheckNil(Nils []keys.Uint256) (nilResps []light.NilValue, e error) {

	return plna.b.CheckNil(Nils)
}
