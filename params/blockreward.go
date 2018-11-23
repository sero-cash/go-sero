// Copyright 2016 The go-ethereum Authors
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

package params

import (
	"math/big"
)

var (

	base = big.NewInt(1e+9)
	defaultReward = new(big.Int).Mul(big.NewInt(5e+10), base)
	devReward = new(big.Int).Mul(big.NewInt(3e+10), base)

	BetaBlockRewards = []*big.Int{
		new(big.Int).Mul(big.NewInt(34214780857), base),
		new(big.Int).Mul(big.NewInt(24757439115), base),
		new(big.Int).Mul(big.NewInt(17914210648), base),
		new(big.Int).Mul(big.NewInt(12962525794), base),
		new(big.Int).Mul(big.NewInt(9379541096), base),
		new(big.Int).Mul(big.NewInt(6786932775), base),
		new(big.Int).Mul(big.NewInt(4910949910), base),
		new(big.Int).Mul(big.NewInt(3553509341), base),
		new(big.Int).Mul(big.NewInt(2571280276), base),
		new(big.Int).Mul(big.NewInt(1860550127), base),
	}

	AiphaBlockRewards = []*big.Int{
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
	}

	DevBlockRewards = []*big.Int{
		devReward,
		devReward,
		devReward,
		devReward,
		devReward,
		devReward,
		devReward,
		devReward,
		devReward,
		devReward,
	}

	DefaultBlockRewards = []*big.Int{
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
		defaultReward,
	}
)
