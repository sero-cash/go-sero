package verify

import (
	"errors"

	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_output_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_output_desc struct {
	desc cpt.OutputVerifyDesc
	pkr  keys.PKr
	e    error
}

func (self *verify_output_desc) Run() bool {
	if keys.PKrValid(&self.pkr) {
		if err := cpt.VerifyOutput(&self.desc); err != nil {
			self.e = err
			return false
		} else {
			return true
		}
	} else {
		self.e = errors.New("z_out pkr is invalid !")
		return false
	}
}
