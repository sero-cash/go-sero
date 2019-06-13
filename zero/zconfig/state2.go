package zconfig

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

var dir = ""

var last_remove_time = int64(0)

func Init_State_dir(d string) {
	dir = d
}

func IsDirExists(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	} else {
		if len(files) > 0 {
			return true
		} else {
			return false
		}
	}
}

func State2_dir() string {
	return filepath.Join(dir, "balance")
}

func Init_State2() {
	os.Mkdir(State2_dir(), os.ModePerm)
}
