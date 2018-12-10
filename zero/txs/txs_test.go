// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package txs

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/zero/txs/pkg"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/txs/zstate"

	"github.com/sero-cash/go-sero/zero/txs/tx"

	"github.com/sero-cash/go-sero/zero/txs/zstate/state1"

	"github.com/sero-cash/go-czero-import/cpt"

	"github.com/sero-cash/go-sero/zero/txs/assets"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/core/state"

	"github.com/sero-cash/go-sero/serodb"
)

type Blocks struct {
	ca  state.Database
	sd  *state.StateDB
	st  *zstate.State
	st0 *zstate.State0
	st1 *state1.State1
}

var g_blocks Blocks

func NewBlock() {
	if g_blocks.ca == nil {
		db := serodb.NewMemDatabase()
		g_blocks.ca = state.NewDatabase(db)
		g_blocks.sd, _ = state.NewGenesis(common.Hash{}, g_blocks.ca)
		g_blocks.st = g_blocks.sd.GetZState()
		g_blocks.st0 = &g_blocks.st.State0
	} else {
		g_blocks.st0.Block = zstate.State0Block{}
		g_blocks.st0.Block.Tree = g_blocks.st0.Cur.Tree.Clone().ToRef()

	}
}

func EndBlock() {
	if g_blocks.st1 == nil {
		st1 := state1.LoadState1(g_blocks.st0, "")
		g_blocks.st1 = &st1
	} else {
		g_blocks.st1.State0 = g_blocks.st0
	}
	g_blocks.st1.UpdateWitness(keys.Seeds2Tks(seeds))
	NewBlock()
}

type user struct {
	i    int
	seed keys.Uint256
	addr keys.Uint512
}

var seeds = []keys.Uint256{}

func newUser(i int) (ret user) {
	fmt.Printf("\n\n===========new user(%v)============\n", i)
	ret = user{}
	ret.i = i
	ret.seed = keys.Uint256{byte(i)}
	ret.addr = keys.Seed2Addr(&ret.seed)
	seeds = append(seeds, ret.seed)
	fmt.Printf("\nseed: ")
	ret.seed.LogOut()
	fmt.Printf("\naddr: ")
	ret.addr.LogOut()
	return
}

func (self *user) getAR() (pkr keys.Uint512) {
	pkr = keys.Addr2PKr(&self.addr, nil)
	fmt.Printf("\nuser(%v):get pkr: ", self.i)
	pkr.LogOut()
	return
}

func (self *user) addOut(v int) {
	out := stx.Out_O{}
	out.Addr = self.getAR()
	out.Asset = assets.NewAsset(
		&assets.Token{
			utils.StringToUint256("SERO"),
			utils.NewU256(uint64(v)),
		},
		nil,
	)
	g_blocks.st.AddOut_O(&out)
	g_blocks.st.Update()
	EndBlock()
}

func (self *user) addTkt(v int) {
	out := stx.Out_O{}
	out.Addr = self.getAR()
	out.Asset = assets.Asset{
		&assets.Token{
			utils.StringToUint256("SERO"),
			utils.NewU256(uint64(v)),
		},
		&assets.Ticket{
			utils.StringToUint256("SERO_TICKET"),
			cpt.Random(),
		},
	}
	g_blocks.st.AddOut_O(&out)
	g_blocks.st.Update()
	EndBlock()
}

func (self *user) GetOuts() (outs []*state1.OutState1) {
	if os, e := g_blocks.st1.GetOuts(keys.Seed2Tk(&self.seed).NewRef()); e != nil {
		panic(e)
		return
	} else {
		outs = os
		return
	}
}

func (self *user) Gen(seed *keys.Uint256, t *tx.T) (s stx.T, e error) {
	return Gen_state1(seed, t, g_blocks.st1)
}

func (self *user) Verify(t *stx.T) (e error) {
	return Verify_state1(t, g_blocks.st1.State0)
}

func (self *user) Logout() (ret uint64) {
	fmt.Printf("\n\n===========user(%v)============\n", self.i)
	outs := self.GetOuts()
	for _, out := range outs {
		if out.Out_O.Asset.Tkn != nil {
			fmt.Printf("TKN: (%v:%v)---%v-----%v\n", out.Pg.Anchor[1], out.Pg.Index, out.Out_O.Asset.Tkn.Currency[0], out.Out_O.Asset.Tkn.Value.ToIntRef().Int64())
			ret += out.Out_O.Asset.Tkn.Value.ToIntRef().Uint64()
		}
		if out.Out_O.Asset.Tkt != nil {
			fmt.Printf("TKT: (%v:%v)---%v-----%v\n", out.Pg.Anchor[1], out.Pg.Index, out.Out_O.Asset.Tkt.Category[0], out.Out_O.Asset.Tkt.Value)
		}
	}
	fmt.Printf("===========user(%v)============\n\n", self.i)
	return
}

func (self *user) Package(v int, fee int, u user) {
	fmt.Printf("user(%v) send %v:%v to user(%v)\n", self.i, v, fee, u.i)
	outs := self.GetOuts()
	in := tx.In{}
	in.Root = *outs[0].Pg.Root.ToUint256()
	out0 := pkg.Pkg_O{}
	out0.PKr = u.addr
	out0.Asset = assets.Asset{
		&assets.Token{
			utils.StringToUint256("SERO"),
			utils.NewU256(uint64(v)),
		},
		nil,
	}

	out1 := tx.Out{}
	out1.Addr = self.addr
	out1.Asset = outs[0].Out_O.Asset.Clone()
	out1.Asset.Tkn.Value.SubU(utils.NewU256(uint64(v)).ToRef())
	out1.Asset.Tkn.Value.SubU(utils.NewU256(uint64(fee)).ToRef())

	out1.Z = tx.TYPE_Z

	t := tx.T{}
	t.Fee = utils.NewU256(uint64(fee))
	t.Ins = append(t.Ins, in)
	t.Outs = append(t.Outs, out1)
	t.PkgPack = &out0

	s, e := self.Gen(&self.seed, &t)
	if e != nil {
		fmt.Printf("user(%v) send gen error: %v", self.i, e)
	}

	if e := self.Verify(&s); e != nil {
		fmt.Printf("user(%v) send verify error: %v", self.i, e)
	}

	g_blocks.st.AddStx(&s)
	g_blocks.st.Update()
	EndBlock()
}

func (self *user) Send(v int, fee int, u user, z bool) {
	fmt.Printf("user(%v) send %v:%v to user(%v)\n", self.i, v, fee, u.i)
	outs := self.GetOuts()
	in := tx.In{}
	in.Root = *outs[0].Pg.Root.ToUint256()
	out0 := tx.Out{}
	out0.Addr = u.addr
	out0.Asset = assets.Asset{
		&assets.Token{
			utils.StringToUint256("SERO"),
			utils.NewU256(uint64(v)),
		},
		nil,
	}
	if z {
		out0.Z = tx.TYPE_Z
	} else {
		out0.Z = tx.TYPE_O
	}

	out1 := tx.Out{}
	out1.Addr = self.addr
	out1.Asset = outs[0].Out_O.Asset.Clone()
	out1.Asset.Tkn.Value.SubU(utils.NewU256(uint64(v)).ToRef())
	out1.Asset.Tkn.Value.SubU(utils.NewU256(uint64(fee)).ToRef())

	if z {
		out1.Z = tx.TYPE_Z
	} else {
		out1.Z = tx.TYPE_O
	}

	t := tx.T{}
	t.Fee = utils.NewU256(uint64(fee))
	t.Ins = append(t.Ins, in)
	t.Outs = append(t.Outs, out0)
	t.Outs = append(t.Outs, out1)

	s, e := self.Gen(&self.seed, &t)
	if e != nil {
		fmt.Printf("user(%v) send gen error: %v", self.i, e)
	}

	if e := self.Verify(&s); e != nil {
		fmt.Printf("user(%v) send verify error: %v", self.i, e)
	}

	g_blocks.st.AddStx(&s)
	g_blocks.st.Update()
	EndBlock()
}

func TestMain(m *testing.M) {
	cpt.ZeroInit("", cpt.NET_Dev)
	NewBlock()
	m.Run()
}

func TestTxs(t *testing.T) {
	user_m := newUser(1)
	user_a := newUser(2)
	user_b := newUser(3)
	user_c := newUser(4)

	user_m.addTkt(100)
	user_m.addOut(100)
	user_m.addOut(100)
	user_m.addOut(100)

	if user_m.Logout() != 400 {
		t.Fail()
	}

	user_m.Package(50, 10, user_a)
	g_blocks.st.OpenPkg(4, &pkg.Key{})
	g_blocks.st.Update()
	EndBlock()

	g_blocks.st.GetPkg(4)
	EndBlock()

	//user_m.Send(50, 10, user_a, true)

	if user_m.Logout() != 340 {
		t.Fail()
	}
	if user_a.Logout() != 50 {
		t.Fail()
	}

	user_m.addOut(100)

	if user_m.Logout() != 440 {
		t.Fail()
	}
	if user_a.Logout() != 50 {
		t.Fail()
	}

	user_a.Send(20, 5, user_b, true)

	if user_a.Logout() != 25 {
		t.Fail()
	}
	if user_b.Logout() != 20 {
		t.Fail()
	}

	user_b.Send(10, 5, user_c, true)

	if user_b.Logout() != 5 {
		t.Fail()
	}
	if user_c.Logout() != 10 {
		t.Fail()
	}
}
