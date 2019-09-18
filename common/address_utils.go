package common

import (
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/address"
)

func AddrToAccountAddr(addr Address) address.AccountAddress {
	var accountAddr address.AccountAddress
	copy(accountAddr[:], addr[:])
	return accountAddr
}

func AddrToPKr(addr Address) c_type.PKr {
	if addr.IsAccountAddress() {
		return c_czero.Addr2PKr(addr.ToUint512(), nil)
	} else {
		return *addr.ToPKr()
	}
}
