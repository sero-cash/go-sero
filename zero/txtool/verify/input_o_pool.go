package verify

import (
	"github.com/sero-cash/go-sero/zero/zconfig"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"
)

var verify_input_o_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_input_o_desc struct {
	hash_z   keys.Uint256
	src      localdb.OutState
	in       stx.In_S
	asset_cc keys.Uint256
	e        error
}

func (self *verify_input_o_desc) Run() error {
	g := cpt.VerifyInputSDesc{}
	g.Ehash = self.hash_z
	g.Nil = self.in.Nil
	g.RootCM = *self.src.ToRootCM()
	g.Sign = self.in.Sign
	g.Pkr = *self.src.ToPKr()
	if err := cpt.VerifyInputS(&g); err != nil {
		self.e = err
		return err
	} else {
		self.asset_cc = self.src.Out_O.ToAssetCC()
		return nil
	}
}
