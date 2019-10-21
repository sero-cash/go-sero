package zconfig

import (
	"fmt"
	"testing"
)

func TestSign(t *testing.T) {
	is := []int{}
	is = append(is, 1)
	is = append(is, 2)
	r_is := &is[0]
	(*r_is) = 2
	fmt.Println(is)
}
