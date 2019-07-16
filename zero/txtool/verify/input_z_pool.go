package verify

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

var verify_input_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_input_desc struct {
	desc cpt.InputVerifyDesc
}

func (self *verify_input_desc) Run() error {
	if err := cpt.VerifyInput(&self.desc); err != nil {
		return err
	} else {
		return nil
	}
}
