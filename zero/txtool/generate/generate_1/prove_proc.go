package generate_1

import (
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

var gen_output_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_p_thread_num })

type gen_output_desc struct {
	asset    c_type.Asset
	ar       c_type.Uint256
	asset_cm c_type.Uint256
	proof    c_type.Proof
	index    int
	e        error
}

func (self *gen_output_desc) Run() error {
	if proof, err := c_superzk.ProveOutput(&self.asset, &self.ar, &self.asset_cm); err != nil {
		self.e = err
		return err
	} else {
		self.proof = proof
		return nil
	}
}

var gen_input_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_p_thread_num })

type gen_input_desc struct {
	asset_cm_new c_type.Uint256
	zpka         c_type.Uint256
	nil          c_type.Uint256
	anchor       c_type.Uint256
	asset_cc     c_type.Uint256
	ar_old       c_type.Uint256
	ar_new       c_type.Uint256
	index        uint64
	zpkr         c_type.Uint256
	vskr         c_type.Uint256
	baser        c_type.Uint256
	a            c_type.Uint256
	paths        [c_type.DEPTH * 32]byte
	pos          uint64
	proof        c_type.Proof
	e            error
}

func (self *gen_input_desc) Run() error {
	if proof, err := c_superzk.ProveInput(
		&self.asset_cm_new,
		&self.zpka,
		&self.nil,
		&self.anchor,
		&self.asset_cc,
		&self.ar_old,
		&self.ar_new,
		self.index,
		&self.zpkr,
		&self.vskr,
		&self.baser,
		&self.a,
		&self.paths,
		self.pos,
	); err != nil {
		self.e = err
		return err
	} else {
		self.proof = proof
		return nil
	}
}

var gen_pkg_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_p_thread_num })

type gen_pkg_desc struct {
	asset_cm c_type.Uint256
	asset    c_type.Asset
	ar       c_type.Uint256
	proof    c_type.Proof
	e        error
}

func (self *gen_pkg_desc) Run() error {
	if proof, e := c_superzk.ProveOutput(&self.asset, &self.ar, &self.asset_cm); e != nil {
		self.e = e
		return e
	} else {
		self.proof = proof
		return nil
	}
}
