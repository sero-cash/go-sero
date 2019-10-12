package utils

import (
	"fmt"
	"sync"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/pkg/errors"
)

type Proc interface {
	Run() error
}

type Procs struct {
	ch   chan int
	wait sync.WaitGroup
	Runs []Proc
	E    error
	ERun Proc
}

func NewProcs(num int) (ret Procs) {
	ret = Procs{
		make(chan int, num),
		sync.WaitGroup{},
		nil,
		nil,
		nil,
	}
	return
}

func (self *Procs) HasProc() bool {
	if len(self.Runs) > 0 {
		return true
	} else {
		return false
	}
}

func (self *Procs) StartProc(run Proc) {
	self.Runs = append(self.Runs, run)
	if c_superzk.Is_czero_debug() {
		if e := run.Run(); e != nil {
			self.E = e
		}
	} else {
		self.wait.Add(1)
		go func(run Proc) {
			defer func() {
				if r := recover(); r != nil {
					self.E = errors.Errorf("process panic: %v", r)
					self.ERun = run
				}
				<-self.ch
				self.wait.Done()
			}()
			self.ch <- 0
			if self.E == nil {
				if e := run.Run(); e != nil {
					self.E = e
					self.ERun = run
				}
			}
		}(run)
	}
}

func (self *Procs) End() error {
	self.wait.Wait()
	if self.E == nil {
		return nil
	} else {
		return self.E
	}
}

type ProcsPool struct {
	pool sync.Pool
}

func NewProcsPool(numget func() int) (ret ProcsPool) {
	ret = ProcsPool{
		sync.Pool{
			New: func() interface{} {
				procs := NewProcs(numget())
				return &procs
			},
		},
	}
	return
}

func (self *ProcsPool) GetProcs() (ret *Procs) {
	ret = self.pool.Get().(*Procs)
	ret.Runs = []Proc{}
	if ret == nil {
		panic(fmt.Errorf("GetProcsFromPool error: fetch nil!"))
	}
	return
}

func (self *ProcsPool) PutProcs(p *Procs) {
	self.pool.Put(p)
}
