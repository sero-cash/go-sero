package zconfig

import (
	"os"
	"path/filepath"
)

func Checkpoint_dir() string {
	return filepath.Join(dir, "checkpoint")
}

func Init_Checkpoint_dir() {
	os.Mkdir(Checkpoint_dir(), os.ModePerm)
}
