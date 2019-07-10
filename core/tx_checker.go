// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/zero/txtool/verify"

	"github.com/sero-cash/go-sero/zero/utils"
	"github.com/sero-cash/go-sero/zero/zconfig"
)

var txCheckerPool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type CheckDesc struct {
	tx    *types.Transaction
	num   uint64
	index int
}

func (self *CheckDesc) Run() error {
	if e := verify.VerifyWithoutState(self.tx.Ehash().NewRef(), self.tx.GetZZSTX(), self.num); e != nil {
		return e
	}
	return nil
}
