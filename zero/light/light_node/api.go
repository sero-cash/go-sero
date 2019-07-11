package light_node

import (
	"github.com/sero-cash/go-czero-import/keys"
	"fmt"
	"bytes"
	"github.com/sero-cash/go-sero/zero/light/light_types"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/log"
)

var current_light *LightNode

func (self *LightNode) CurrentLight() *LightNode {
	return current_light
}

func (self *LightNode) GetOutsByPKr(pkrs []keys.PKr, start, end uint64) (br BlockOutResp, e error) {
	fmt.Printf("start=[%d],end=[%d]\n", start, end)
	br.CurrencyNum = self.getLastNumber()
	blockOuts := []BlockOut{}
	for _, pkr := range pkrs {
		uPKr := pkr.ToUint512()
		prefix := append(pkrPrefix, uPKr[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)

		for ok := iterator.Seek(pkrKey(pkr.ToUint512(), start)); ok; ok = iterator.Next() {

			key := iterator.Key()
			num := bytesToUint64(key[len(key)-8:])
			fmt.Println("getOutsByPKr:", num)
			if num > end {
				break
			}
			var outs []light_types.Out
			if err := rlp.Decode(bytes.NewReader(iterator.Value()), &outs); err != nil {
				log.Error("Light Invalid block RLP", "Num:", num, "err:", err)
				return br, err
			} else {
				blockOut := BlockOut{Num: num, Outs: outs}
				blockOuts = append(blockOuts, blockOut)
			}
		}
	}
	br.BlockOuts = blockOuts
	return br, nil
}

func (self *LightNode) CheckNil(Nils []keys.Uint256, start uint64, end uint64) (delNils []BlockDelNil, e error) {
	if len(Nils) == 0 {
		return
	}
	for _, Nil := range Nils {
		prefix := append(nilPrefix, Nil[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)
		for ok := iterator.Seek(nilKey(Nil, start)); ok; ok = iterator.Next() {
			delNil := BlockDelNil{}
			key := iterator.Key()
			num := bytesToUint64(key[len(key)-8:])
			if num > end {
				break
			}
			fmt.Println("getOutsByNil:", num)
			value := bytesToUint64(iterator.Value())
			if value == 1 {
				delNil.Num = num
				delNil.Nil = Nil
				delNils = append(delNils, delNil)
			}
		}
	}
	return delNils, nil
}

type BlockOutResp struct {
	CurrencyNum uint64
	BlockOuts   []BlockOut
}

type BlockOut struct {
	Num  uint64
	Outs []light_types.Out
}

type BlockDelNil struct {
	Num uint64
	Nil keys.Uint256
}
