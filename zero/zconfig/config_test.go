package zconfig

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestFile(t *testing.T) {

	files, err := ioutil.ReadDir("/Users")
	fmt.Print(files, err)

}
