// Copyright 2015 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package zconfig

import (
    "path/filepath"
    "os"
    "io/ioutil"
    "fmt"
    "time"
)

var is_dev=false

func Init_Dev(dev bool) {
    is_dev=dev
}
func Is_Dev()bool {
    return is_dev
}



var dir=""

func State1_dir() string {
    return filepath.Join(dir,"state1")
}

var last_remove_time=int64(0)
func Remove_State1_dir_files(height int) {
    current_remove_time:=time.Now().Unix()
    if current_remove_time-last_remove_time > 30 {
        state1_dir := State1_dir()
        if files, err := ioutil.ReadDir(state1_dir); err != nil {
            panic(err)
        } else {
            for _, file := range files {
                name := file.Name()
                var index int
                fmt.Sscanf(name, "%d.", &index)
                if height-index > 35 {
                    path := filepath.Join(state1_dir, name)
                    os.Remove(path)
                    fmt.Printf("remove state1 file: %s\n", name)
                }
            }
        }
        last_remove_time=current_remove_time
    } else {}
}

func Init_State1_dir(d string) {
    dir=d
    state1_dir:=State1_dir()
    os.Mkdir(state1_dir, os.ModePerm)
}

func State1_file(last_fork string) string {
    state1_dir:=State1_dir()
    os.Mkdir(state1_dir, os.ModePerm)
    file:=filepath.Join(state1_dir,last_fork)
    return file
}

