package utils

import (
	"fmt"

	"github.com/sero-cash/go-sero/log"
)

type Progress struct {
	start  int64
	target uint64
	label  string
}

func NewProgress(label string, target uint64) (ret Progress) {
	return Progress{
		start:  -1,
		target: target,
		label:  label,
	}
}

func (self *Progress) Tick(cur uint64, ctx ...interface{}) {
	if self.start == -1 {
		self.start = int64(cur)
	}

	dist := self.target - uint64(self.start)

	if dist == 0 {
		return
	}

	prog := cur - uint64(self.start)
	step := dist / 10
	if step == 0 {
		step = 1
	}

	if cur == self.target || cur == uint64(self.start) || prog%step == 0 {
		percent := (float32(prog) / float32(dist)) * 100
		p := []interface{}{
			"t", self.target,
			"c", cur,
			"p", fmt.Sprintf("%d%%", uint64(percent)),
		}
		p = append(p, ctx...)
		log.Info(
			self.label,
			p...,
		)
	}
}
