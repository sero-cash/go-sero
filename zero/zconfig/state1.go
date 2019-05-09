package zconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func State1_dir() string {
	return filepath.Join(dir, "state1")
}

func Remove_State1_dir_files(height uint64) {
	current_remove_time := time.Now().Unix()
	if current_remove_time-last_remove_time > 30 {
		state1_dir := State1_dir()
		if files, err := ioutil.ReadDir(state1_dir); err != nil {
			panic(err)
		} else {
			reserveds := NewReserveds(height)

			for _, file := range files {
				name := file.Name()
				var index int
				fmt.Sscanf(name, "%d.", &index)
				path := filepath.Join(state1_dir, name)
				if files := reserveds.Insert(uint64(index), path); files != nil {
					for _, file := range files {
						os.Remove(file)
					}
				}
			}
		}
		last_remove_time = current_remove_time
	} else {
	}
}

func State1_file(last_fork string) string {
	state1_dir := State1_dir()
	os.Mkdir(state1_dir, os.ModePerm)
	file := filepath.Join(state1_dir, last_fork)
	return file
}

func Init_State1() {
	os.Mkdir(State1_dir(), os.ModePerm)
}
