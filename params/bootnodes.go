// Copyright 2015 The go-ethereum Authors
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

// BetanetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Ethereum network.
var BetanetBootnodes = []string{
	"snode://c4c6aa3e3c73c1d86d53f904bb82ffbbc50edd75530487f053d37bbcf75e27ca4c7ec215f56274cf60176b5d7fb896c9bd54354ba2fedb7d3adb35efc053cd14@13.251.221.90:53719",
	"snode://1eb4b0bfa17bd70601b70da8df9070f5b4d1757864e0361e730bba04ccdc8e02cb0d51b8428921570684357eb9c053c433e23085a51e6ffde1c0d734f8dbbfe6@140.143.152.43:53719",
	"snode://f5f92eb85242a4ad9853d7315caeb5ff38ff90aa137e2348d42fef7ea90eb3012d369f4eae1efefb26ee20ee41c6380798f77108109039336b77ec9d65def74f@111.230.148.82:53719",
	"snode://81abed64774e612362f133e045289d4e6644cbfe6421c21856f0f01eb4145531256ce0a2b61bd1baa100f14b1e940fda79643fbdf34a75a5a3e470a1ab4e8a8e@118.25.146.113:53719",
}

// AlphanetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var AlphanetBootnodes = []string{
	"snode://29ad1a59fbb8984ec167b890ddb52e8690c6c1699c6796ca209cc2d0a5603760f25747fdae61ecb937260491f51632e6651a9292b2b3ce871e6c2e0c7c0f7c00@118.25.146.113:60609",
}

var DevBootnodes = []string{
	//"snode://7105c6d1e9a8d0c12ecda61d4db860d62d84e2a53e777681e2d098b0b8cc592f48e0368fd82b353d7f8acec7321151441638164a21285879263f9c6e05a63f6e@192.168.15.220:60609",
	//"snode://91d9d0c77947d376f66c2a60146b235ab3aab39116bfae238d9c2aabea318b15b5df0154fc5ff82c855b09bbadd362fae6dc8533729bdc6345d3c3b566ca7c02@192.168.15.165:60609",
	//"snode://63aa57c5f311a1e848f3b77b16ec00f9afe2456f171af8f4838a7211688bb9c1cc836aa20852c8aef3487b988e4021594eb3aa03a604b868482a6c764bf46fe5@192.168.15.169:60609",
	//"snode://f081cf0dd496069518e92462c8fce964247b816008259efe8f37ac35fa1ed17d2a6759d64544884b3dec430e001dfae7e1224a22f4a0a992fe9fe94cf020f551@118.25.146.113:60606",
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"snode://bb50e2d89a4ed70663d080659fe0ad4b9bc3e06c17a227433966cb59ceee020decddbf6e00192011648d13b1c00af770c0c1bb609d4d3a5c98a43772e0e18ef4@192.168.15.252:60601",
}
