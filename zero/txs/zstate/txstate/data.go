package txstate

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/localdb"
)

type Data struct {
	Cur    Current
	Block  StateBlock
	G2ins  map[keys.Uint256]bool
	G2outs map[keys.Uint256]*localdb.OutState

	Dirty_last_out bool
	Dirty_G2ins    map[keys.Uint256]bool
	Dirty_G2outs   map[keys.Uint256]bool
}

func (state *Data) clear() {
	state.Cur = NewCur()
	state.G2ins = make(map[keys.Uint256]bool)
	state.G2outs = make(map[keys.Uint256]*localdb.OutState)
	state.Block = StateBlock{}
	state.clear_dirty()
}

func (state *Data) clear_dirty() {
	state.Dirty_last_out = false
	state.Dirty_G2ins = make(map[keys.Uint256]bool)
	state.Dirty_G2outs = make(map[keys.Uint256]bool)
}
