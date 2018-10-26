package tx

import (
	"fmt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
	"testing"
)

func TestT_TokenCost(t *testing.T) {
	seroCy:=stringToUint256("sero")
	fmt.Printf("%t\n",seroCy)
	cy:=stringToUint256("d")
	ret := make(map[keys.Uint256]utils.U256)
	ret[seroCy]=utils.NewU256(24)
	if cost,ok:=ret[seroCy];ok {
		add := utils.NewU256(12)
		cost.AddU(&add)
		ret[seroCy] =cost;
	}else{
		cost := utils.NewU256(48)
		ret[cy] =cost;
	}

	fmt.Printf("%t",ret)

}
