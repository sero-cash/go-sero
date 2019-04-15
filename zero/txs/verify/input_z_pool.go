package verify

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_input_procs_pool = utils.NewProcsPool(func() int { return G_v_thread_num })

type verify_input_desc struct {
	desc cpt.InputVerifyDesc
	e    error
}

func (self *verify_input_desc) Run() bool {
	if err := cpt.VerifyInput(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}
