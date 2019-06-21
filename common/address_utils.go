package common

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common/address"
)

func AddrToAccountAddr(addr Address) address.AccountAddress {
	var accountAddr address.AccountAddress
	copy(accountAddr[:], addr[:])
	return accountAddr
}

func AddrToPKr(addr Address) keys.PKr {
	if addr.IsAccountAddress() {
		return keys.Addr2PKr(addr.ToUint512(), nil)
	} else {
		return *addr.ToPKr()
	}
}
