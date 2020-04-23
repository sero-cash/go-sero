package statistics

import (
	"fmt"
	"testing"
)

func f1() {
	Now()
	defer Since()
	fmt.Println("FuncName1 =", runFuncName())
}

func Test2(t *testing.T) {
	fmt.Println("FuncName2 =", runFuncName(), )
	f1()
	print();
}

