package accounts

import (
	"fmt"
	"testing"
)

func TestCreatWallet(t *testing.T) {
	mnemonic, err := CreatWallet("")
	if err != nil {
		t.Error("create wallet failed")
	}
	fmt.Println(mnemonic)
}
