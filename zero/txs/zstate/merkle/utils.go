package merkle

import (
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
)

const DEPTH = c_type.DEPTH

func Combine(l *c_type.Uint256, r *c_type.Uint256) (out c_type.Uint256) {
	return c_czero.Combine(l, r)
}

func createEmpty() (ret [c_type.DEPTH + 1]c_type.Uint256) {
	ret[0] = c_type.Empty_Uint256
	for i := 1; i <= c_type.DEPTH; i++ {
		ret[i] = Combine(&ret[i-1], &ret[i-1])
	}
	return
}

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

func CalcRoot(value *c_type.Uint256, pos uint64, paths *[DEPTH]c_type.Uint256) (ret c_type.Uint256) {
	ret = *value
	for _, path := range paths {
		if pos%2 == 0 {
			ret = Combine(&ret, &path)
		} else {
			ret = Combine(&path, &ret)
		}
		pos >>= 1
	}
	return
}
