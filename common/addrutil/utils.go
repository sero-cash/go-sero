package addrutil

import "github.com/sero-cash/go-sero/common/base58"

func FromBase58(s string, out []byte) {
	base58.DecodeString(s, out)
}
