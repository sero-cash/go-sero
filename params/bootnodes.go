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
	"snode://bb50e2d89a4ed70663d080659fe0ad4b9bc3e06c17a227433966cb59ceee020decddbf6e00192011648d13b1c00af770c0c1bb609d4d3a5c98a43772e0e18ef4@bootnode0-us-dev.sero.cash:60609",
	"snode://13967b95dd14651a28e9b91e8a0f1d6a803fdd63b88040ad8734e474cd0b24248e9aa64768a0d175282f8ecbb1e936cfd42d663f88727f645455aa56bc17cd9e@54.193.115.126:60609",
	"snode://8a2ae32aa351d92cc777b6867b7dfd2f2b37bc7266355809ad0370a953a4329f2b33a4f0a10db42814f5161e4cac7dfa1c752535fb86de0e025e915d8a5f6396@13.251.221.90:60609",
	"snode://32575f813feaf24546f0e9d30d5ef15816f614c185f239d583162e16fab94379089f3a2194043cb72eef0d2b308b8b7017ed6df1a7e7c6eca09ab9b67900b0f2@52.58.7.220:60609",
	"snode://a64756b1859b7d51a95cfd183e5a01b6fa581c5a572c62a6685a3c1f05adbd58f68d8153c2dad466b8d3c5903d0e09054776a8ee5d3e9fae50f61e68266744dc@140.143.152.43:60609",
}

// AlphanetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var AlphanetBootnodes = []string{
	"snode://29ad1a59fbb8984ec167b890ddb52e8690c6c1699c6796ca209cc2d0a5603760f25747fdae61ecb937260491f51632e6651a9292b2b3ce871e6c2e0c7c0f7c00@118.25.146.113:60609",
}

var DevBootnodes = []string{
	"snode://ea30c41faf986f827ac5c6436e84645453eb0fd7974708b6172a49adfd87515e1c9043ab6e759da2678055eb296dded91cfd7ecbb90db0465dd50320eefd0a07@192.168.15.220:60609",
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"snode://bb50e2d89a4ed70663d080659fe0ad4b9bc3e06c17a227433966cb59ceee020decddbf6e00192011648d13b1c00af770c0c1bb609d4d3a5c98a43772e0e18ef4@192.168.15.252:60601",
}
