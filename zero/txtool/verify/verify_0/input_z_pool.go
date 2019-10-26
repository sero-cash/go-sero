package verify_0

import (
	"encoding/hex"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

var verify_input_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_input_desc struct {
	desc c_czero.InputVerifyDesc
}

func (self *verify_input_desc) Run() error {
	if err := c_czero.VerifyInput(&self.desc); err != nil {
		log.Warn("verify in_z proof error", "nil", hex.EncodeToString(self.desc.Nil[:]))
		return err
	} else {
		return nil
	}
}
