package common

import (
	"fmt"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/keys"
)

var ZEROBYTES = [32]byte{}

func getAddressSuffix(addr Address) [32]byte {
	suffix := [32]byte{}
	copy(suffix[:], addr[64:])
	return suffix
}

func AddrToAccountAddr(addr Address) address.AccountAddress {
	accountAddress := address.AccountAddress{}
	copy(accountAddress[:], addr[:64])
	return accountAddress
}

func IsPkr(addr *Address) (bool, error) {
	if addr == nil {
		return false, errors.New("addr cannot be nil")
	}
	if *addr == (Address{}) {
		return false, errors.New("addr cannot be zero address")
	}
	suffix := getAddressSuffix(*addr)
	if ZEROBYTES == suffix {
		return false, nil
	} else {
		if keys.PKrValid(addr.ToPKr()) {
			return true, nil
		} else {
			return false, errors.New(fmt.Sprintf("invalid addr %v", addr.String()))
		}
	}
}

/**
  do not support contract address
*/
func AddressToPkr(addr Address) (keys.PKr, error) {
	flag, err := IsPkr(&addr)
	if err != nil {
		return keys.PKr{}, err
	}
	if flag {
		pkr := addr.ToPKr()
		return *pkr, nil
	} else {
		accountAddress := AddrToAccountAddr(addr)
		return keys.Addr2PKr(accountAddress.ToUint512(), keys.RandUint256().NewRef()), nil
	}

}
