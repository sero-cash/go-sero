package ethash

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/common/hexutil"
)

func TestProgpow(t *testing.T) {

	cache_size := calcCacheSize(0)
	data_size := calcDatasetSize(0)
	seed := seedHash(0*epochLength + 1)

	cache := make([]uint32, cache_size/4)
	generateCache(cache, 0, seed)
	header, _ := hexutil.Decode("0x5ffee07b6b16bc6f364c45b84d412138a0b1588edb74e4123e419384435e1691")

	cDag := make([]uint32, progpowCacheWords)

	d, r := progpowLightWithoutCDag(data_size, cache, cDag, header, 15017396847274520746, 50)
	fmt.Printf("d: %v,r: %v", hexutil.Encode(d), hexutil.Encode(r))
}

func TestProgpowFull(t *testing.T) {
	csize := cacheSize(1)
	dsize := datasetSize(1)
	cache := make([]uint32, csize/4)
	seed := seedHash(0*epochLength + 1)
	generateCache(cache, 0, seed)

	dataset := make([]uint32, dsize/4)
	generateDataset(dataset, 0, cache)

	header, _ := hexutil.Decode("0x5ffee07b6b16bc6f364c45b84d412138a0b1588edb74e4123e419384435e1691")
	d, r := progpowFull(dataset, header, 15017396847274520746, 50)
	fmt.Printf("d: %v,r: %v", hexutil.Encode(d), hexutil.Encode(r))
}
