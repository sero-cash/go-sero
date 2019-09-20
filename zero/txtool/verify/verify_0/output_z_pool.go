package verify_0

import (
	"errors"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_output_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_output_desc struct {
	desc c_czero.OutputVerifyDesc
	pkr  c_type.PKr
	e    error
}

func (self *verify_output_desc) Run() error {
	if superzk.IsPKrValid(&self.pkr) {
		if err := c_czero.VerifyOutput(&self.desc); err != nil {
			self.e = err
			return err
		} else {
			return nil
		}
	} else {
		self.e = errors.New("z_out pkr is invalid !")
		return self.e
	}
}
