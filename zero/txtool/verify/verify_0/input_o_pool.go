package verify_0

import (
	"github.com/sero-cash/go-sero/zero/txs/stx/stx_v1"
	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_input_o_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_input_o_desc struct {
	hash_z   c_type.Uint256
	src      localdb.OutState
	in       stx_v1.In_S
	asset_cc c_type.Uint256
	e        error
}

func (self *verify_input_o_desc) Run() error {
	g := c_czero.VerifyInputSDesc{}
	g.Ehash = self.hash_z
	g.Nil = self.in.Nil
	g.RootCM = *self.src.ToRootCM()
	g.Sign = self.in.Sign
	g.Pkr = *self.src.ToPKr()
	if err := c_czero.VerifyInputS(&g); err != nil {
		self.e = err
		return err
	} else {
		self.asset_cc = self.src.Out_O.ToAssetCC()
		return nil
	}
}
