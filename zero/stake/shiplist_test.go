package stake

import (
	"fmt"
	"github.com/sero-cash/go-sero/common"
	"math/big"
	"math/rand"
	"testing"
)

func TestSkipKist(t *testing.T) {
	list := newSkipList()
	for i := 0; i < 20; i++ {
		hash := common.BigToHash(new(big.Int).SetInt64(int64(rand.Intn(100))))
		fmt.Println(hash)
		list.insert(hash)
	}
	print(list)

	fmt.Println("\n--------------------------------------")

	list.delete(common.BigToHash(new(big.Int).SetInt64(int64(10))))
	print(list)

	fmt.Println("\n--------------------------------------")
}

func print(s *skipList) {
	fmt.Println()

	for i := s.level - 1; i >= 0; i-- {
		current := s.head
		for current.forward[i] != nil {
			fmt.Printf("%d \n", current.forward[i].key)
			current = current.forward[i]
		}
		fmt.Printf("***************** Level %d \n", i+1)
	}
}
