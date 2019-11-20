package verify_1

import (
	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

var verify_output_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_output_desc struct {
	pkr      c_type.PKr
	asset_cm c_type.Uint256
	proof    c_type.Proof
	isEx     bool
	e        error
}

func (self *verify_output_desc) Run() error {
	if c_superzk.IsPKrValid(&self.pkr) {
		if err := c_superzk.VerifyOutput(&self.asset_cm, &self.proof, self.isEx); err != nil {
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
