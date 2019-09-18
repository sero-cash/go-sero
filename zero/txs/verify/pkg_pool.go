package verify

import (
	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

var verify_pkg_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_pkg_desc struct {
	desc c_czero.PkgVerifyDesc
	e    error
}

func (self *verify_pkg_desc) Run() error {
	if err := c_czero.VerifyPkg(&self.desc); err != nil {
		self.e = err
		return err
	} else {
		return nil
	}
}
