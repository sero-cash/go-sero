package utils

import (
	"fmt"
	"sync"

	"github.com/sero-cash/go-czero-import/cpt"
)

type Proc interface {
	Run() bool
}

type Procs struct {
	ch   chan int
	wait sync.WaitGroup
	Runs []Proc
	succ bool
}

func NewProcs(num int) (ret Procs) {
	ret = Procs{
		make(chan int, num),
		sync.WaitGroup{},
		nil,
		true,
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
	if cpt.Is_czero_debug() {
		if !run.Run() {
			self.succ = false
		}
	} else {
		self.wait.Add(1)
		go func(run Proc) {
			self.ch <- 0
			defer func() {
				<-self.ch
				self.wait.Done()
			}()
			if !run.Run() {
				self.succ = false
			}
		}(run)
	}
}

func (self *Procs) Wait() []Proc {
	self.wait.Wait()
	if self.succ {
		p := self.Runs
		self.Runs = nil
		return p
	} else {
		return nil
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
	if ret == nil {
		panic(fmt.Errorf("GetProcsFromPool error: fetch nil!"))
	}
	return
}

func (self *ProcsPool) PutProcs(p *Procs) {
	self.pool.Put(p)
}
