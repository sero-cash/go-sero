package pkgstate

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type Data struct {
	G2pkgs       map[keys.Uint256]*localdb.ZPkg
	Block        Block
	Dirty_G2pkgs map[keys.Uint256]bool
}

func (state *Data) clear() {
	state.G2pkgs = make(map[keys.Uint256]*localdb.ZPkg)
	state.Block.Pkgs = []keys.Uint256{}
	state.Dirty_G2pkgs = make(map[keys.Uint256]bool)
}
