package ethash_hash

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/hexutil"
)

func TestHash(t *testing.T) {
	k := c_type.RandUint256()
	o := Miner_Hash_0(k[:], 0)
	fmt.Print(hexutil.Encode(o))
}
