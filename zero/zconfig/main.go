package zconfig

import "io/ioutil"

var dir = ""

func Init_Data_dir(d string) {
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
