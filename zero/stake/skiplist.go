package stake

import (
	"github.com/sero-cash/go-sero/common"
	"math/big"
	"math/rand"
	"time"
)

const MaxLevel = 32
const Probability = 0.25 // 基于时间与空间综合 best practice 值, 越上层概率越小

func randLevel() (level int) {
	rand.Seed(time.Now().UnixNano())
	for level = 1; rand.Float32() < Probability && level < MaxLevel; level++ {
		//fmt.Println(rand.Float32())
	}
	//fmt.Printf("up to %d level\n", level)
	return
}

type node struct {
	forward []*node
	key     common.Hash
}

type skipList struct {
	head  *node
	level int
	state StakeState
}

func newNode(key common.Hash, level int) *node {
	return &node{key: key, forward: make([]*node, level)}
}

func newSkipList() *skipList {
	return &skipList{head: newNode(common.Hash{}, MaxLevel), level: 1}
}

func (s *skipList) cmpHash(key1, key2 common.Hash) int {
	return new(big.Int).SetBytes(key1[0:24]).Cmp(new(big.Int).SetBytes(key1[0:24]))
}

func (s *skipList) insert(key common.Hash) {
	current := s.head
	update := make([]*node, MaxLevel) // 新节点插入以后的前驱节点
	for i := s.level - 1; i >= 0; i-- {
		if current.forward[i] == nil || current.forward[i].key.Big().Cmp(key.Big()) > 0 {
			update[i] = current
		} else {
			for current.forward[i] != nil && current.forward[i].key.Big().Cmp(key.Big()) < 0 {
				current = current.forward[i] // 指针往前推进
			}
			update[i] = current
		}
	}

	level := randLevel()
	if level > s.level {
		// 新节点层数大于跳表当前层数时候, 现有层数 + 1 的 head 指向新节点
		for i := s.level; i < level; i++ {
			update[i] = s.head
		}
		s.level = level
	}
	node := newNode(key, level)
	for i := 0; i < level; i++ {
		node.forward[i] = update[i].forward[i]
		update[i].forward[i] = node
	}
}

func (s *skipList) delete(key common.Hash) {
	current := s.head
	for i := s.level - 1; i >= 0; i-- {
		for current.forward[i] != nil {
			if current.forward[i].key.Big().Cmp(key.Big()) == 0 {
				tmp := current.forward[i]
				current.forward[i] = tmp.forward[i]
				tmp.forward[i] = nil
			} else if current.forward[i].key.Big().Cmp(key.Big()) > 0 {
				break
			} else {
				current = current.forward[i]
			}
		}
	}
}

func (s *skipList) search(key int) *node {
	// 类似 delete
	return nil
}
