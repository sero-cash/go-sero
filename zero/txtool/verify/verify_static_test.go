package verify

import (
	"fmt"
	"sync/atomic"
	"testing"
)

type Test struct {
	t atomic.Value
}

func TestConsRecord2(t *testing.T) {
	tm := []Test{}
	tm = append(tm, Test{})

	for _, v := range tm {
		v.t.Store(1)
	}

	a := tm[0].t.Load().(int)
	fmt.Println(a)
}
