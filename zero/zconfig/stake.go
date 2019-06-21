package zconfig

import "path/filepath"

func Stake_dir() string {
	return filepath.Join(dir, "stake")
}
