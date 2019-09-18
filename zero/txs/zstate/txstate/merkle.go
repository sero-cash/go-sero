package txstate

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/zero/txs/zstate/merkle"
)

var CzeroAddress = c_type.NewPKrByBytes(crypto.Keccak512(nil))
var CzeroMerkleParam = merkle.NewParam(&CzeroAddress)

var SzkAddress = c_type.NewPKrByBytes(crypto.Keccak256([]byte("$SuperZK$MerkleTree")))
var SzkMerkleParam = merkle.NewParam(&SzkAddress)
