package state2

import (
	"github.com/sero-cash/go-sero/zero/lstate"
	"github.com/sero-cash/go-sero/zero/lstate/state1"
)

func InitLState(bc lstate.BlockChain) {
	ns := state1.NewState1(bc)
	lstate.Run(bc, &ns)
	return
}

func (self *State2) Parse(last_chose uint64) (chose uint64) {
	return last_chose
}
