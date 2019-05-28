package ethash_hash

/*
 -mno-stack-arg-probe disables stack probing which avoids the function
 __chkstk_ms being linked. this avoids a clash of this symbol as we also
 separately link the secp256k1 lib which ends up defining this symbol

 1. https://gcc.gnu.org/onlinedocs/gccint/Stack-Checking.html
 2. https://groups.google.com/forum/#!msg/golang-dev/v1bziURSQ4k/88fXuJ24e-gJ
 3. https://groups.google.com/forum/#!topic/golang-nuts/VNP6Mwz_B6o

*/

/*


#cgo CFLAGS: -std=gnu99 -Wall
#cgo windows CFLAGS: -mno-stack-arg-probe -Wunused-function
#cgo LDFLAGS: -lm

#include "blake2b.c"

*/
import "C"
import "unsafe"

func Miner_Hash_0(in []byte, num uint64) []byte {
	var bs [64]byte
	C.hash_enter(
		(*C.uchar)(unsafe.Pointer(&bs[0])),
		(*C.uchar)(unsafe.Pointer(&in[0])),
		C.ulonglong(num),
	)
	return bs[:]
}

func Miner_Hash_1(in []byte, num uint64) []byte {
	var bs [32]byte
	C.hash_leave(
		(*C.uchar)(unsafe.Pointer(&bs[0])),
		(*C.uchar)(unsafe.Pointer(&in[0])),
		C.ulonglong(num),
	)
	return bs[:]
}
