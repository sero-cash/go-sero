package utils

import (
	"fmt"
	"time"
)

var TR_enable = false

type TimeRecord struct {
	start time.Time
	name  string
}

func TR_enter(name string) (tr TimeRecord) {
	tr.start = time.Now()
	tr.name = name
	return
}

func (tr *TimeRecord) Renter(name string) {
	if TR_enable {
		td := time.Since(tr.start)
		fmt.Printf("TR-----("+tr.name+")     s=%v\n", td)
		tr.name = name
		tr.start = time.Now()
	}
}

func (tr *TimeRecord) Leave() time.Duration {
	td := time.Since(tr.start)
	if TR_enable {
		fmt.Printf("TR-----("+tr.name+")     s=%v\n", td)
	}
	return td
}
