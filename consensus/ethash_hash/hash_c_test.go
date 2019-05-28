package ethash_hash

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/common/hexutil"

	"github.com/sero-cash/go-czero-import/keys"
)

func TestHash(t *testing.T) {
	k := keys.RandUint256()
	o := Miner_Hash_0(k[:], 0)
	fmt.Print(hexutil.Encode(o))
}
