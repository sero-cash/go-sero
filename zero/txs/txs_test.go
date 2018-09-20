// Copyright 2015 The sero.cash Authors
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

	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txs/tx"
	"github.com/sero-cash/go-sero/zero/txs/zstate"
	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/witness/merkle"
)

type user struct {
	i      int
	seed   keys.Uint256
	addr   keys.Uint512
	zstate *zstate.State
}

var seeds = []keys.Uint256{}

func newUser(i int, zstate *zstate.State) (ret user) {
	fmt.Printf("\n\n===========new user(%v)============\n", i)
	ret = user{}
	ret.i = i
	ret.seed = keys.Uint256{byte(i)}
	ret.addr = keys.Seed2Addr(&ret.seed)
	ret.zstate = zstate
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
	out.Value = utils.NewU256(uint64(v))
	self.zstate.AddOut_O(&out, &keys.Uint256{})
	self.zstate.Update()
	self.zstate.FinalizeGenWitness(keys.Seeds2Tks(seeds))
}

func (self *user) Logout() {
	db := ethdb.NewMemDatabase()
	ca := state.NewDatabase(db)
	st, _ := state.New(common.Hash{}, ca, 0)
	fmt.Printf("\n\n===========user(%v)============\n", self.i)
	if outs, e := GetOuts(keys.Seed2Tk(&self.seed).NewRef(), st.GetZState()); e != nil {
		fmt.Printf("user(%v) get outs error: %v", self.i, e)
	} else {
		for _, out := range outs {
			fmt.Printf("(%v)---%v-----%v\n", out.Witness.Tree.Root()[1], out.Out_O.Currency[0], out.Out_O.Out.Value)
		}
	}
	fmt.Printf("===========user(%v)============\n\n", self.i)
}

func (self *user) Send(v int, fee int, u user, z bool) {
	fmt.Printf("user(%v) send %v:%v to user(%v)\n", self.i, v, fee, u.i)
	if outs, e := GetOuts(keys.Seed2Tk(&self.seed).NewRef(), self.zstate); e != nil {
		fmt.Printf("user(%v) get outs error: %v", self.i, e)
	} else {
		in := tx.In{}
		in.Root = outs[0].Witness.Tree.RootKey()
		out0 := tx.Out{}
		out0.Addr = u.addr
		out0.Value = utils.NewU256(uint64(v))
		if z {
			out0.Z = tx.TYPE_Z
		} else {
			out0.Z = tx.TYPE_O
		}

		out1 := tx.Out{}
		out1.Addr = self.addr
		out1.Value.AddU(&outs[0].Out_O.Out.Value)
		out1.Value.SubU(utils.NewU256(uint64(v)).ToRef())
		out1.Value.SubU(utils.NewU256(uint64(fee)).ToRef())

		if z {
			out1.Z = tx.TYPE_Z
		} else {
			out1.Z = tx.TYPE_O
		}

		t := tx.T{}
		t.CTxs = append(t.CTxs, tx.CTx{})
		t.CTxs[0].Fee = utils.NewU256(uint64(fee))
		t.CTxs[0].Ins = append(t.CTxs[0].Ins, in)
		t.CTxs[0].Outs = append(t.CTxs[0].Outs, out0)
		t.CTxs[0].Outs = append(t.CTxs[0].Outs, out1)

		s, e := Gen(&self.seed, &t, self.zstate)
		if e != nil {
			fmt.Printf("user(%v) send gen error: %v", self.i, e)
		}

		if e := Verify(&s, self.zstate); e != nil {
			fmt.Printf("user(%v) send verify error: %v", self.i, e)
		}

		self.zstate.AddStx(&s)
		self.zstate.Update()
		self.zstate.FinalizeGenWitness(keys.Seeds2Tks(seeds))
	}
}

func TestTxs(t *testing.T) {
	db := serodb.NewMemDatabase()
	ca := state.NewDatabase(db)
	st, _ := state.NewGenesis(common.Hash{}, ca)

	//-----miner m dig block-----
	user_m := newUser(1, st.GetZState())
	user_a := newUser(2, st.GetZState())
	user_b := newUser(3, st.GetZState())
	user_c := newUser(4, st.GetZState())

	user_m.addOut(100)
	user_m.addOut(100)
	user_m.addOut(100)
	user_m.addOut(100)
	user_m.Logout()

	user_m.Send(50, 10, user_a, false)
	user_m.Logout()

	user_m.addOut(100)
	user_m.Logout()
	user_a.Logout()

	user_a.Send(20, 5, user_b, true)
	user_a.Logout()
	user_b.Logout()

	user_b.Send(10, 5, user_c, true)
	user_b.Logout()
	user_c.Logout()

}

type TT struct {
	i int
}

func TestXXX(t *testing.T) {
	sl := []TT{{1}, {2}}
	var p *TT = nil

	for _, v := range sl {
		r := v
		fmt.Print(v)
		p = &r
	}

	fmt.Print(p)
}

func TestZstateRLP(t *testing.T) {

	type testOutSate struct {
		Tree   merkle.Tree
		Desc_Z *stx.Desc_Z `rlp:"nil"`
	}

	//Desc_Z :=&stx.Desc_Z{R:keys.Uint256{25}}

	out := zstate.OutState0{}
	t.Logf("%t", out)
	enc, err := rlp.EncodeToBytes(out)
	if err != nil {
		t.Logf("%v", err)
	}
	decodeOut := zstate.OutState0{}
	err = rlp.DecodeBytes(enc, &decodeOut)
	if err != nil {
		t.Logf("%v", err)
	}
	t.Logf("%t", decodeOut)

}

func TestLowrRLP(t *testing.T) {

	type testOutSate struct {
		Root keys.Uint256
	}

	out := testOutSate{keys.Uint256{1}}
	t.Logf("%v", out)
	enc, err := rlp.EncodeToBytes(out)
	if err != nil {
		t.Logf("%v", err)
	}
	decodeOut := testOutSate{}
	err = rlp.DecodeBytes(enc, &decodeOut)
	if err != nil {
		t.Logf("%v", err)
	}
	t.Logf("%v", decodeOut)

}

func TestMain(m *testing.M) {
	cpt.ZeroInit()
	m.Run()
}
