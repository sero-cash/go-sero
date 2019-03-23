package data

import (
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/rlp"
)

type Block struct {
	Pkgs []keys.Uint256
}

func (self *Block) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}

type BlockGet struct {
	Out *Block
}

func (self *BlockGet) Unserial(v []byte) (e error) {
	if len(v) < 2 {
		self.Out = nil
		return
	} else {
		self.Out = &Block{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			self.Out = nil
			return
		} else {
			return
		}
	}
}
