package verify

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_pkg_procs_pool = utils.NewProcsPool(func() int { return G_v_thread_num })

type verify_pkg_desc struct {
	desc cpt.PkgVerifyDesc
	e    error
}

func (self *verify_pkg_desc) Run() bool {
	if err := cpt.VerifyPkg(&self.desc); err != nil {
		self.e = err
		return false
	} else {
		return true
	}
}
