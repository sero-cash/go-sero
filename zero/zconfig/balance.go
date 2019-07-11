package zconfig

import (
	"os"
	"path/filepath"
)

func Balance_dir() string {
	return filepath.Join(dir, "balance")
}

func Init_Balance_dir() {
	os.Mkdir(Balance_dir(), os.ModePerm)
}
