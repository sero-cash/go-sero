package utils

import (
	"fmt"
	"testing"

	"github.com/sero-cash/go-czero-import/keys"

	"github.com/mohae/deepcopy"
)

type copytest struct {
	M map[string]bool
	L []string
	P *keys.Uint256
}

func TestSnapshots(t *testing.T) {
	ct := copytest{}
	ct.M = make(map[string]bool)
	ct.M["hello"] = true
	ct.L = append(ct.L, "list")
	ct.P = &keys.Uint256{1}

	ss := Snapshots{}
	ss.Push(0, &ct)

	ct1 := copytest{}
	ct1 = *ss.Revert(1).(*copytest)
	fmt.Print(ct1)
}

func TestDeepCopy(t *testing.T) {
	ct := copytest{}
	ct.M = make(map[string]bool)
	ct.M["hello"] = true
	ct.L = append(ct.L, "list")
	ct.P = &keys.Uint256{1}

	tr := TR_enter("dp")
	cp := deepcopy.Copy(ct).(copytest)
	tr.Renter("cp")
	ctd := copytest{}
	DeepCopy(&ctd, &ct)
	tr.Leave()

	fmt.Print(cp)

	DeepCopy(&ct, &ct)
}
