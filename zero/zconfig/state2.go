package zconfig

import (
	"os"
	"path/filepath"
)

func State2_dir() string {
	return filepath.Join(dir, "state2")
}

func Init_State2() {
	os.Mkdir(State2_dir(), os.ModePerm)
}
