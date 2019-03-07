package utils

import (
	"fmt"
	"time"
)

type TimeRecord struct {
	start  time.Time
	name   string
	enable bool
}

func TR_enter(name string) (tr TimeRecord) {
	tr.start = time.Now()
	tr.name = name
	tr.enable = false
	return
}

func TR_enter_f(name string) (tr TimeRecord) {
	tr.start = time.Now()
	tr.name = name
	tr.enable = true
	return
}

func (tr *TimeRecord) Renter(name string) {
	if tr.enable {
		td := time.Since(tr.start)
		fmt.Printf("TR-----("+tr.name+")     s=%v\n", td)
		tr.name = name
		tr.start = time.Now()
	}
}

func (tr *TimeRecord) Leave() time.Duration {
	td := time.Since(tr.start)
	if tr.enable {
		fmt.Printf("TR-----("+tr.name+")     s=%v\n", td)
	}
	return td
}
