package stx

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/crypto/sha3"
)

func (self *T) Tx1_Hash_From() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *T) Tx1_Hash_Cmd() (ret c_type.Uint256) {
	return self.Desc_Cmd.ToHash()
}

func (self *T) Tx1_Hash_Pkg() (ret c_type.Uint256) {
	return self.Desc_Pkg.Tx1_Hash()
}

func (self *T) Tx1_Hash_Tx1() (ret c_type.Uint256) {
	return self.Tx1.Tx1_Hash()
}

func (self *T) Tx1_Hash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Tx1_Hash_From().NewRef()[:])
	d.Write(self.Tx1_Hash_Cmd().NewRef()[:])
	d.Write(self.Tx1_Hash_Pkg().NewRef()[:])
	d.Write(self.Tx1_Hash_Tx1().NewRef()[:])
	copy(ret[:], d.Sum(nil))
	return ret
}
