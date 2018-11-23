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
	BetaBlockRewards = []*big.Int{
		big.NewInt(34214780857e+9),
		big.NewInt(24757439115e+9),
		big.NewInt(17914210648e+9),
		big.NewInt(12962525794e+9),
		big.NewInt(9379541096e+9),
		big.NewInt(6786932775e+9),
		big.NewInt(4910949910e+9),
		big.NewInt(3553509341e+9),
		big.NewInt(2571280276e+9),
		big.NewInt(1860550127e+9),
	}

	AiphaBlockRewards = []*big.Int{
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
	}

	DevBlockRewards = []*big.Int{
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
		big.NewInt(3e+19),
	}

	DefaultBlockRewards = []*big.Int{
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
		big.NewInt(5e+19),
	}
)
