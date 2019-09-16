package txs

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/txs/pkg"
)

func WatchPkg(id *c_type.Uint256, key *c_type.Uint256) (ret pkg.Pkg_O, pkr c_type.PKr, e error) {
	/*
		st1 := lstate.CurrentLState()
		if st1 == nil {
			e = errors.New("Watch Pkg but lstate is nil")
			return
		}
		pg := st1.CurrentZState().Pkgs.GetPkgById(id)
		if pg == nil || pg.Closed {
			e = errors.New("Watch Pkg but has been closed")
			return
		}
		pkr = pg.Pack.PKr
		ret, e = pkg.DePkg(key, &pg.Pack.Pkg)
	*/
	return
}
