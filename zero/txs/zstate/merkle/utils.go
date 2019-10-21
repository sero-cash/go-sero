package merkle

import (
	"github.com/sero-cash/go-czero-import/c_type"
)

const DEPTH = c_type.DEPTH

func toDepth(index uint64) (ret uint8) {
	ret = 0
	for index != 0 {
		index >>= 1
		ret++
	}
	return DEPTH + 1 - ret
}

func toPow2(index int) (ret uint64) {
	ret = uint64(1) << uint64(index)
	return
}
